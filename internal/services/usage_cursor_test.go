package services

import (
	"testing"
	"time"
)

func TestDecodeUsageCursorRejectsInvalidValues(t *testing.T) {
	for _, cursor := range []string{"not-base64", "", "e30"} {
		_, _, err := decodeUsageCursor(cursor)
		if cursor == "" && err != nil {
			t.Fatalf("empty cursor returned error: %v", err)
		}
		if cursor != "" && err == nil {
			t.Errorf("cursor %q succeeded; want error", cursor)
		}
	}
}

func TestUsageCursorRoundTrip(t *testing.T) {
	wantTime := time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC)
	raw := encodeUsageCursor(usageEventCursor{OccurredAt: wantTime, ID: "event-1"})

	gotTime, gotID, err := decodeUsageCursor(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !gotTime.Equal(wantTime) || gotID != "event-1" {
		t.Fatalf("decoded cursor = %s/%s, want %s/event-1", gotTime, gotID, wantTime)
	}
}
