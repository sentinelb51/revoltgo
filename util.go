package revoltgo

import (
	"fmt"
	"log"
	"net/url"
	"strings"
	"unicode"

	"github.com/goccy/go-json"
)

// Ptr is a quality-of-life function to quickly return a pointer to any value
func Ptr[T any](v T) *T {
	return &v
}

// WebhookFromURL extracts the webhook ID and token from a full webhook URL, for example:
// https://stoat.chat/api/webhooks/08UE096D31N8ZM7PJFAHFSST8R/5uC1w7KjCixSD49XFOQ2qEkLo3ukvWDbK6_EEfZ-1gqRYRiVaXI8mYrSDjRv0T1y
func WebhookFromURL(uri string) (wID, wToken string, err error) {
	u, err := url.Parse(uri)
	if err != nil {
		err = fmt.Errorf("cannot parse url: %w", err)
		return
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) != 4 {
		err = fmt.Errorf("too few parts")
		return
	}

	if parts[0] != "api" || parts[1] != "webhooks" {
		err = fmt.Errorf("missing api or webhooks segment")
		return
	}

	return parts[2], parts[3], nil
}

func StrTrimAfter(s, substr string) string {
	index := strings.Index(s, substr)
	if index == -1 {
		return s
	}

	return s[:index]
}

// mergeJSON deserializes the object into JSON, then merges the data into the object.
// It will also remove any fields specified in the clear map.
func mergeJSON[T any](object *T, data json.RawMessage, clearFields []string) {

	decoded := make(map[string]any)
	err := json.Unmarshal(data, &decoded)
	if err != nil {
		log.Printf("Error unmarshalling data: %s\n", err)
		return
	}

	// Marshal the object to JSON
	objectBytes, err := json.Marshal(object)
	if err != nil {
		log.Printf("Error marshalling object: %s\n", err)
		return
	}

	// Unmarshal the JSON into a map
	objectMap := make(map[string]any)
	err = json.Unmarshal(objectBytes, &objectMap)
	if err != nil {
		log.Printf("Error unmarshalling object: %s\n", err)
		return
	}

	// Merge the data into the object map
	for key, value := range decoded {
		objectMap[key] = value
	}

	// Remove any fields specified in the clearFields slice
	hasClear := len(clearFields) > 0
	if hasClear {
		for _, key := range clearFields {
			delete(objectMap, toSnakeCase(key))
		}
	}

	// Marshal the map back into JSON
	objectBytes, err = json.Marshal(objectMap)
	if err != nil {
		log.Printf("Error marshalling object: %s\n", err)
		return
	}

	// Determine if we need to create a new object (burden the GC) or update the existing one
	// If anything was deleted (cleared), a new object is required because Unmarshal will not overwrite existing fields
	var result *T
	if !hasClear {
		result = object // Re-use old object
	} else {
		result = new(T) // Allocate new object
	}

	err = json.Unmarshal(objectBytes, result)
	if err != nil {
		log.Printf("Error unmarshalling new object: %s\n", err)
		return
	}

	// If required, overwrite the original object with the new object
	if hasClear {
		*object = *result
	}
}

// ToSnakeCase converts a CamelCase string to snake_case
func toSnakeCase(str string) string {
	var (
		result strings.Builder
		size   = len(str)
		growBy = size % 4 // Assume every 4 characters, there can underscore
	)

	// Return if small string
	if size < 2 {
		return str
	}

	// Grow buffer to avoid re-allocations
	result.Grow(size + growBy)

	// Skip processing first letter
	result.WriteRune(unicode.ToLower(rune(str[0])))

	// Start loop after 1 character
	for _, r := range str[1:] {

		if !unicode.IsUpper(r) {
			result.WriteRune(r)
			continue
		}

		result.WriteRune('_')
		result.WriteRune(unicode.ToLower(r))
	}

	return result.String()
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

func mustParseURL(raw string) *url.URL {
	u, err := url.Parse(raw)
	if err != nil {
		panic(err)
	}
	return u
}

func validateBaseURL(newURL string) (u *url.URL, err error) {
	newURL = strings.TrimSuffix(strings.TrimSpace(newURL), "/")

	u, err = url.Parse(newURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	if u.Scheme != "https" {
		return nil, fmt.Errorf("base URL must use HTTPS")
	}

	if u.Path != "" {
		return nil, fmt.Errorf("base URL must not have a path (trailing /)")
	}

	if u.Host == "" {
		return nil, fmt.Errorf("base URL must have a host")
	}

	if strings.Count(u.Hostname(), ".") < 1 {
		return nil, fmt.Errorf("base URL must have a domain and TLD")
	}

	return
}
