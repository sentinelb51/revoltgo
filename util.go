package revoltgo

import (
	"reflect"
	"strings"
)

// clearByJSON will clear the value of a field based on its json tag.
func clearByJSON(object any, query string) {
	objectValue := reflect.ValueOf(object).Elem()
	objectType := objectValue.Type()

	for i := 0; i < objectValue.NumField(); i++ {
		field := objectValue.Field(i)
		structField := objectType.Field(i)
		tag := structField.Tag.Get("json")

		if !strings.EqualFold(tag, query) {
			continue
		}

		if field.CanSet() {
			field.SetZero()
		}
	}
}

func merge[T any](base T, contrast any) T {
	baseValue := reflect.ValueOf(base).Elem()
	contrastValue := reflect.ValueOf(contrast).Elem()

	for i := 0; i < baseValue.NumField(); i++ {
		contrastValuesField := contrastValue.Field(i)

		// Skip if the contrast value is nil or zero
		if contrastValuesField.Kind() == reflect.Ptr && contrastValuesField.IsNil() ||
			contrastValuesField.IsZero() {
			continue
		}

		baseValue.Field(i).Set(contrastValuesField)
	}

	return base
}

// sliceRemoveIndex removes the element at the specified index from slice.
// If the index is out of bounds, slice is returned unchanged.
func sliceRemoveIndex[T any](slice []T, index int) []T {
	if index < 0 {
		panic("index must be >= 0")
	} else if index >= len(slice) {
		panic("index must be < len(slice)")
	}

	// Pre-calculate size to avoid unnecessary len() calls
	size := len(slice) - 1

	// Swap the element to be removed with the last element
	slice[index], slice[size] = slice[size], slice[index]

	// Exclude last element, effectively removing it
	return slice[:size]
}
