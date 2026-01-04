package revoltgo

// This file provides utility functions for working with Updatable entities.
// The Updatable interface unifies the update pattern across different entity types.

// UpdateEntity is a generic helper that updates an entity using its update method.
// This demonstrates how the Updatable interface enables generic code.
func UpdateEntity[T any, U Updatable[T]](entity U, data T) {
	entity.update(data)
}

// ClearEntityFields is a generic helper that clears specific fields from an entity.
// This demonstrates how the Updatable interface enables generic code.
func ClearEntityFields[T any, U Updatable[T]](entity U, fields []string) {
	entity.clear(fields)
}

// ApplyPartialUpdate applies both an update and field clearing in a single operation.
// This is useful for handling API responses that include both updates and removals.
func ApplyPartialUpdate[T any, U Updatable[T]](entity U, data T, clearFields []string) {
	entity.update(data)
	entity.clear(clearFields)
}
