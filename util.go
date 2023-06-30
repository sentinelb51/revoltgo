package revoltgo

import (
	"math/rand"
	"reflect"
	"strings"
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

// clear will clear the value of a field based on its json tag.
func clear(object any, query string) {
	values := reflect.ValueOf(object).Elem()
	valuesType := values.Type()

	for i := 0; i < values.NumField(); i++ {
		field := values.Field(i)
		structField := valuesType.Field(i)
		tag := structField.Tag.Get("json")

		if !strings.EqualFold(tag, query) {
			continue
		}

		if field.CanSet() {
			field.SetZero()
		}
	}
}

func merge(base, contrast any) any {
	baseValues := reflect.ValueOf(base).Elem()
	contrastValues := reflect.ValueOf(contrast).Elem()

	for i := 0; i < baseValues.NumField(); i++ {
		contrastValuesField := contrastValues.Field(i)
		shouldUpdate := false

		if contrastValuesField.Kind() == reflect.Ptr {
			shouldUpdate = !contrastValuesField.IsNil()
		} else {
			shouldUpdate = !contrastValuesField.IsZero()
		}

		if shouldUpdate {
			baseValues.Field(i).Set(contrastValuesField)
		}
	}

	return base
}

// sliceRemoveIndex removes the element at the specified index from slice.
// If index is out of bounds, slice is returned unchanged.
func sliceRemoveIndex[T any](slice []T, index int) []T {

	if index < 0 {
		panic("index must be >= 0")
	}

	if index >= len(slice) {
		return slice
	}

	return append(slice[:index], slice[index+1:]...)
}
