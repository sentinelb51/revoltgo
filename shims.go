package revoltgo

import (
	json "encoding/json/v2"
	"fmt"
	"math"
	"slices"
	"strconv"
	"time"

	"github.com/tinylib/msgp/msgp"
)

// The shim below binds the gateway only, and does not contradict REST's RFC3339:
// the gateway is rmp_serde (not human-readable), where iso8601-timestamp emits an
// i64 of milliseconds, while REST is serde_json (human-readable), where the same
// type emits an ISO8601 string. One Go field, two encodings, both correct.

//go:generate msgp -tests=false -io=false
//msgp:shim time.Time as:int64 using:timeToMs/msToTime

/* timeToMs and msToTime are important shims;
ServerMember.JoinedAt will fail to decode otherwise, and the ready event will never be decoded
*/

// msToTime converts int64 (wire) -> time.Time
func msToTime(ms int64) time.Time {
	return time.UnixMilli(ms)
}

// timeToMs converts time.Time (wire) -> int64
func timeToMs(t time.Time) int64 {
	return t.UnixMilli()
}

/* Wire shapes that the generated codecs get wrong on their own */

//msgp:ignore NullableBytes NullableBool

// NullableBytes is a Vec<u8> as Stoat's two encoders emit it, neither of which
// msgp's generated []byte codec reads. Serde has no byte string: the gateway
// writes an array of integers, or nil for a None, and REST writes an array of
// numbers where encoding/json wants a base64 string. One refused field fails the
// whole frame, so a single thumbhash in Ready loses the entire snapshot.
type NullableBytes []byte

func (b NullableBytes) MarshalMsg(o []byte) ([]byte, error) {
	return msgp.AppendBytes(o, b), nil
}

func (b *NullableBytes) UnmarshalMsg(bts []byte) ([]byte, error) {

	// Three forms for one field, all of them legal MessagePack: rmp_serde writes
	// a Vec<u8> as an array of integers rather than a byte string, and a None as
	// nil. The byte string is what msgp itself would have written, so a value
	// this package encoded round-trips too.
	switch msgp.NextType(bts) {
	case msgp.NilType:
		*b = nil
		return msgp.ReadNilBytes(bts)

	case msgp.ArrayType:
		count, bts, err := msgp.ReadArrayHeaderBytes(bts)
		if err != nil {
			return bts, err
		}

		out := slices.Grow((*b)[:0], int(count))[:count]

		for i := range out {
			var value uint8
			if value, bts, err = msgp.ReadUint8Bytes(bts); err != nil {
				return bts, msgp.WrapError(err, i)
			}
			out[i] = value
		}

		*b = out

		return bts, nil

	default:
		var err error
		*b, bts, err = msgp.ReadBytesBytes(bts, (*b)[:0])

		return bts, err
	}
}

func (b NullableBytes) Msgsize() int {
	return msgp.BytesPrefixSize + len(b)
}

// MarshalJSON writes the array of numbers serde reads back, not the base64
// string the default []byte encoding would.
func (b NullableBytes) MarshalJSON() ([]byte, error) {
	if b == nil {
		return []byte("null"), nil
	}

	out := make([]byte, 0, 2+4*len(b))
	out = append(out, '[')

	for i, v := range b {
		if i > 0 {
			out = append(out, ',')
		}
		out = strconv.AppendUint(out, uint64(v), 10)
	}

	return append(out, ']'), nil
}

func (b *NullableBytes) UnmarshalJSON(data []byte) error {

	// Read through a wider element type: encoding/json reads a []byte from a
	// base64 string, and the numbers serde writes are not that.
	var values []uint16
	if err := json.Unmarshal(data, &values); err != nil {
		return fmt.Errorf("NullableBytes: %w", err)
	}

	out := slices.Grow((*b)[:0], len(values))[:len(values)]

	for i, value := range values {
		if value > math.MaxUint8 {
			return fmt.Errorf("NullableBytes: element %d is not a byte: %d", i, value)
		}
		out[i] = byte(value)
	}

	*b = out

	return nil
}

// NullableBool is a bool the gateway may send as nil, where msgp's generated
// decoder expects one of the two boolean bytes. A None means false: the field
// exists to say a thing is so, and nothing said is not so.
type NullableBool bool

func (v NullableBool) MarshalMsg(o []byte) ([]byte, error) {
	return msgp.AppendBool(o, bool(v)), nil
}

func (v *NullableBool) UnmarshalMsg(bts []byte) ([]byte, error) {
	if msgp.IsNil(bts) {
		*v = false
		return msgp.ReadNilBytes(bts)
	}

	value, bts, err := msgp.ReadBoolBytes(bts)
	*v = NullableBool(value)

	return bts, err
}

func (v NullableBool) Msgsize() int {
	return msgp.BoolSize
}
