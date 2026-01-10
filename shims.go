package revoltgo

import "time"

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
