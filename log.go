package revoltgo

import (
	"log"
	"sync/atomic"
)

// logger is where this package writes its diagnostics. A library must not write
// to the program's log unless it is asked to, so it discards until SetLogger
// installs one. It is atomic because the gateway logs from its own goroutine.
var logger atomic.Pointer[log.Logger]

// SetLogger routes this package's diagnostics to l. A nil l silences them,
// which is the default.
func SetLogger(l *log.Logger) {
	logger.Store(l)
}

// logf writes one diagnostic. With no logger installed it costs an atomic load
// and does not evaluate its arguments' String methods.
func logf(format string, v ...any) {
	if l := logger.Load(); l != nil {
		l.Printf(format, v...)
	}
}
