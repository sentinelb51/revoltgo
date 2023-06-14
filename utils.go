package revoltgo

import (
	"math/rand"
	"time"

	"github.com/oklog/ulid/v2"
)

var (
	entropySrc = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func ULID() string {
	t := time.Now()
	entropy := ulid.Monotonic(entropySrc, 0)

	return ulid.MustNew(ulid.Timestamp(t), entropy).String()
}
