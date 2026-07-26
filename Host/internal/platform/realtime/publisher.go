package realtime

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"time"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/subscriptions"
)

type Event[T any] struct {
	EventID    string `json:"eventId"`
	GameID     string `json:"gameId,omitempty"`
	Revision   int    `json:"revision,omitempty"`
	Kind       string `json:"kind"`
	OccurredAt string `json:"occurredAt"`
	Payload    T      `json:"payload"`
}

type Authorize func(auth *core.Record) bool

func NewEventID() string {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return time.Now().UTC().Format("20060102T150405.000000000")
	}
	const alphabet = "0123456789abcdef"
	encoded := make([]byte, len(value)*2)
	for i, current := range value {
		encoded[i*2] = alphabet[current>>4]
		encoded[i*2+1] = alphabet[current&0x0f]
	}
	return string(encoded)
}

func Publish[T any](app core.App, topic string, event Event[T], authorize Authorize) error {
	if event.EventID == "" || event.Kind == "" {
		return errors.New("realtime event requires an event id and kind")
	}
	if event.OccurredAt == "" {
		event.OccurredAt = time.Now().UTC().Format(time.RFC3339Nano)
	}
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	message := subscriptions.Message{Name: topic, Data: data}
	for _, chunk := range app.SubscriptionsBroker().ChunkedClients(100) {
		for _, client := range chunk {
			if !client.HasSubscription(topic) {
				continue
			}
			auth, _ := client.Get(apis.RealtimeClientAuthKey).(*core.Record)
			if authorize != nil && !authorize(auth) {
				continue
			}
			client.Send(message)
		}
	}
	return nil
}
