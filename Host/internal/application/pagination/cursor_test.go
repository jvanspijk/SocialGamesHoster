package pagination

import (
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"
)

func TestCursorRoundTripUsesURLSafeJSON(t *testing.T) {
	cursor := Cursor{
		Created: time.Date(2026, time.August, 1, 12, 0, 0, 0, time.UTC),
		ID:      "same-timestamp-id",
	}

	encoded := Encode(cursor)
	data, err := json.Marshal(cursor)
	if err != nil {
		t.Fatal(err)
	}
	if want := base64.RawURLEncoding.EncodeToString(data); encoded != want {
		t.Fatalf("encoded cursor = %q, want %q", encoded, want)
	}
	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if !decoded.Created.Equal(cursor.Created) || decoded.ID != cursor.ID {
		t.Fatalf("decoded cursor = %#v, want %#v", decoded, cursor)
	}
}

func TestDecodeRejectsMalformedOrIncompleteCursor(t *testing.T) {
	for _, value := range []string{
		"not-a-cursor",
		encodedCursor(t, Cursor{}),
		encodedCursor(t, Cursor{Created: time.Date(2026, time.August, 1, 12, 0, 0, 0, time.UTC)}),
	} {
		if _, err := Decode(value); err == nil {
			t.Fatalf("Decode(%q) succeeded", value)
		}
	}
}

func TestWindowUsesLimitPlusOneAndFinalReturnedCursor(t *testing.T) {
	type record struct{ ID string }
	created := time.Date(2026, time.August, 1, 12, 0, 0, 0, time.UTC)
	for _, count := range []int{0, 2, PageSize, QueryLimit} {
		t.Run(string(rune('0'+count%10)), func(t *testing.T) {
			records := make([]record, count)
			for index := range records {
				records[index] = record{ID: string(rune('a' + index))}
			}

			items, next := Window(records, func(record record) Cursor {
				return Cursor{Created: created, ID: record.ID}
			})
			wantCount := count
			if wantCount > PageSize {
				wantCount = PageSize
			}
			if len(items) != wantCount {
				t.Fatalf("items = %d, want %d", len(items), wantCount)
			}
			if count <= PageSize {
				if next != "" {
					t.Fatalf("next cursor = %q, want empty", next)
				}
				return
			}
			cursor, err := Decode(next)
			if err != nil || !cursor.Created.Equal(created) || cursor.ID != records[PageSize-1].ID {
				t.Fatalf("next cursor = (%#v, %v)", cursor, err)
			}
		})
	}
}

func TestDescendingCreatedIDPredicate(t *testing.T) {
	if DescendingCreatedIDSort != "-created,-id" {
		t.Fatalf("sort = %q", DescendingCreatedIDSort)
	}
	const want = "(created < {:created} || (created = {:created} && id < {:id}))"
	if DescendingCreatedIDPredicate != want {
		t.Fatalf("predicate = %q, want %q", DescendingCreatedIDPredicate, want)
	}
}

func encodedCursor(t *testing.T, cursor Cursor) string {
	t.Helper()
	data, err := json.Marshal(cursor)
	if err != nil {
		t.Fatal(err)
	}
	return base64.RawURLEncoding.EncodeToString(data)
}
