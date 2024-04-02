package core

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

// Core is the interface that declares the core functionality of the application.
type Core interface {
	CreateItem(description string) TodoItem
	UpdateItem(id int, completed bool) (TodoItem, error)
	DeleteItem(id int) error
	GetItems(completed bool) []TodoItem
}

// NOTE: TheCore is meant to be used as the only implementation of the Core interface. Defining the functionalities as methods allows for being replaced by a mock core in the tests.

// TheCore is the implementation of the Core interface.
type TheCore struct {
	accessor StorageAccessor
}

func NewCore(accessor StorageAccessor) *TheCore {
	return &TheCore{accessor: accessor}
}

type TodoItem struct {
	ID          int
	Description string
	Completed   bool
}

type TodoItemNotFoundError struct {
	ID int
}

func (e TodoItemNotFoundError) Error() string {
	return fmt.Sprintf("TodoItem with id %d not found", e.ID)
}

func (c *TheCore) CreateItem(description string) TodoItem {
	log.WithFields(log.Fields{"description": description}).Info("CORE: Adding new TodoItem.")
	todo := TodoItem{Description: description, Completed: false}
	_, err := c.accessor.Create(&todo)
	if err != nil {
		log.Fatal("CORE: ", err)
	}
	return todo
}

func (c *TheCore) UpdateItem(id int, completed bool) (TodoItem, error) {
	todos := c.accessor.Read(func(todo TodoItem) bool {
		return todo.ID == id
	})
	if len(todos) == 0 {
		err := TodoItemNotFoundError{ID: id}
		log.Warn("CORE: ", err)
		return TodoItem{}, err
	}
	if len(todos) > 1 {
		log.Fatal("CORE: Multiple TodoItems with the same id.")
	}
	todo := todos[0]
	todo.Completed = completed

	log.WithFields(log.Fields{"id": id, "completed": completed}).Info("CORE: Updating TodoItem.")
	err := c.accessor.Update(todo)
	if err != nil {
		log.Warn("CORE: ", err)
		return TodoItem{}, err
	}
	return todo, nil
}

func (c *TheCore) DeleteItem(id int) error {
	log.WithFields(log.Fields{"id": id}).Info("CORE: Deleting TodoItem.")
	err := c.accessor.Delete(id)
	if err != nil {
		log.Warn("CORE: ", err)
		return err
	}
	return nil
}

func (c *TheCore) GetItems(completed bool) []TodoItem {
	log.Info("CORE: Getting TodoItems. completed=", completed)
	todos := c.accessor.Read(func(todo TodoItem) bool {
		return todo.Completed == completed
	})
	return todos
}
