package revoltgo

import (
	"testing"
)

var sinkFn func() any

// validFrame builds a MessagePack frame: 6-byte prefix, 0xA7 fixstr header
// (length 7), then "Message" — the happy path for eventTypeFromMSGP.
func validFrame() []byte {
	f := make([]byte, 6, 14)
	f = append(f, 0xA7)
	return append(f, []byte("Message")...)
}

// TestHandleLookupZeroAlloc guards the hot path: extracting the event type and
// looking it up as a map key must not allocate. The zero-alloc property relies
// on the compiler's m[string(b)] optimization, which only fires when the
// conversion is written inline in the index expression. Hoisting it into a
// local that escapes (e.g. string(eventType) passed to a logger on the happy
// path) would reintroduce an allocation per event and fail this test.
func TestHandleLookupZeroAlloc(t *testing.T) {
	raw := validFrame()
	allocs := testing.AllocsPerRun(1000, func() {
		et, err := eventTypeFromMSGP(raw)
		if err != nil {
			t.Fatal(err)
		}
		sinkFn = eventConstructors[string(et)]
	})
	if allocs != 0 {
		t.Errorf("event-type lookup allocated %v times per run; want 0", allocs)
	}
}
