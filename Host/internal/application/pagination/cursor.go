// Package pagination provides the shared descending created/id cursor protocol
// used by bounded collection listings.
package pagination

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

const (
	PageSize   = 50
	QueryLimit = PageSize + 1

	DescendingCreatedIDSort      = "-created,-id"
	DescendingCreatedIDPredicate = "(created < {:created} || (created = {:created} && id < {:id}))"
)

// Cursor is the stable position of an item in a descending created/id listing.
type Cursor struct {
	Created time.Time `json:"created"`
	ID      string    `json:"id"`
}

// Encode returns the URL-safe cursor representation.
func Encode(cursor Cursor) string {
	data, _ := json.Marshal(cursor)
	return base64.RawURLEncoding.EncodeToString(data)
}

// Decode validates and returns a URL-safe cursor representation.
func Decode(value string) (Cursor, error) {
	data, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return Cursor{}, err
	}
	var cursor Cursor
	if err := json.Unmarshal(data, &cursor); err != nil || cursor.Created.IsZero() || cursor.ID == "" {
		return Cursor{}, fmt.Errorf("invalid cursor")
	}
	return cursor, nil
}

// Window applies the shared limit-plus-one paging rule and calculates the next
// cursor from the final returned item.
func Window[T any](records []T, cursorFor func(T) Cursor) ([]T, string) {
	if len(records) <= PageSize {
		return records, ""
	}
	records = records[:PageSize]
	return records, Encode(cursorFor(records[len(records)-1]))
}
