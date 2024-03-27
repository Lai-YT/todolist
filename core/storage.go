package core

// StorageAccessor is an interface that defines the functions that the core package will use to interact with the storage layer.
type StorageAccessor interface {
	// Create creates a new TodoItem and returns the id of the new TodoItem. The id is also updated in the TodoItem.
	Create(*TodoItem) (id int, e error)
	// Read returns a list of TodoItems that satisfy the condition specified by the where function.
	Read(where func(TodoItem) bool) []TodoItem
	// Update updates a TodoItem with the new values specified in the todo parameter.
	Update(todo TodoItem) error
	// Delete deletes a TodoItem with the specified id.
	Delete(id int) error
}
