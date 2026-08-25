package revoltgo

import "time"

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
