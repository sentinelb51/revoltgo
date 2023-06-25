package revoltgo

import (
	"math/rand"
	"reflect"
	"time"

	"github.com/oklog/ulid/v2"
)

var (
	entropySrc = rand.New(rand.NewSource(time.Now().UnixNano()))
	entropy    = ulid.Monotonic(entropySrc, 0)
)

func NewULID() string {
	return ulid.MustNew(ulid.Timestamp(time.Now()), entropy).String()
}

func merge(base, contrast any) any {
	baseValues := reflect.ValueOf(base).Elem()
	contrastValues := reflect.ValueOf(contrast).Elem()

	for i := 0; i < baseValues.NumField(); i++ {
		baseValuesField := baseValues.Field(i)
		contrastValuesField := contrastValues.Field(i)

		shouldUpdate := false

		if contrastValuesField.Kind() == reflect.Ptr {
			shouldUpdate = !contrastValuesField.IsNil()
		} else {
			shouldUpdate = !contrastValuesField.IsZero()
		}

		if shouldUpdate {
			baseValuesField.Set(contrastValuesField)
		}
	}

	return base
}
