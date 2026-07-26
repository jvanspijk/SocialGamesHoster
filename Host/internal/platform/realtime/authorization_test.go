package realtime

import "testing"

func TestGameMasterTopicRemainsDistinctFromPublicGameTopic(t *testing.T) {
	kind, id, ok := splitTopic("game:game123:game-masters")
	if !ok || kind != "game-master" || id != "game123" {
		t.Fatalf("unexpected game-master topic: %q %q %t", kind, id, ok)
	}
	kind, _, ok = splitTopic("game:game123:public")
	if !ok || kind != "game" {
		t.Fatalf("unexpected public topic: %q %t", kind, ok)
	}
}

func TestRejectsUnknownRealtimeTopics(t *testing.T) {
	for _, topic := range []string{"game:game123:private", "game:game123:game-masters:extra", "room:", "collections/games"} {
		if _, _, ok := splitTopic(topic); ok {
			t.Fatalf("expected %q to be rejected", topic)
		}
	}
}

func TestProfileRequestCapabilityTopicParsing(t *testing.T) {
	capability := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	requestID, actual, ok := splitProfileRequestTopic("profile-request:request123:" + capability)
	if !ok || requestID != "request123" || actual != capability {
		t.Fatalf("unexpected capability topic: %q %q %t", requestID, actual, ok)
	}
	if _, _, ok := splitProfileRequestTopic("profile-request:request123:short"); ok {
		t.Fatal("expected malformed capability to be rejected")
	}
}
