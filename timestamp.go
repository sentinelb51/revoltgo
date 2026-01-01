package revoltgo

import (
	"time"

	"github.com/tinylib/msgp/msgp"
)

//go:generate msgp -tests=false -io=false
//msgp:ignore Timestamp

type Timestamp struct {
	time.Time
}

// UnmarshalMsg implements msgp.Unmarshaler
func (t *Timestamp) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var ms int64

	ms, o, err = msgp.ReadInt64Bytes(bts)
	if err != nil {
		return
	}

	t.Time = time.UnixMilli(ms)
	return
}

// MarshalMsg implements msgp.Marshaler
func (t *Timestamp) MarshalMsg(b []byte) (o []byte, err error) {
	return msgp.AppendInt64(b, t.UnixMilli()), nil
}

// Msgsize implements msgp.Sizer
func (t *Timestamp) Msgsize() int {
	return msgp.Int64Size
}
