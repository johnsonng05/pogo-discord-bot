package api

import "testing"

func TestFetchEventsLive(t *testing.T) {
    client := New()
    events, err := client.FetchEvents()
    if err != nil {
        t.Fatal(err)
    }
    if len(events) == 0 {
        t.Fatal("got 0 events")
    }
    if events[0].Name == "" || events[0].Heading == "" || events[0].Start == "" || events[0].End == "" {
        t.Fatalf("fields empty — check json tags: %+v", events[0])
    }
}
