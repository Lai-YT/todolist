package core

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	log "github.com/sirupsen/logrus"
)

var accessor StorageAccessor

func SetAccessor(sa StorageAccessor) {
	accessor = sa
}

type TodoItem struct {
	Id          int
	Description string
	Completed   bool
}

func CreateItem(description string) TodoItem {
	log.WithFields(log.Fields{"description": description}).Info("CORE: Adding new TodoItem.")
	todo := TodoItem{Description: description, Completed: false}
	_, err := accessor.Create(&todo)
	if err != nil {
		log.Fatal("CORE: ", err)
	}
	return todo
}

type TodoItemNotFoundError struct {
	Id int
}

func (e TodoItemNotFoundError) Error() string {
	return fmt.Sprintf("TodoItem with id %d not found", e.Id)
}

func UpdateItem(id int, completed bool) (TodoItem, error) {
	todos := accessor.Read(func(todo TodoItem) bool {
		return todo.Id == id
	})
	if len(todos) == 0 {
		err := TodoItemNotFoundError{Id: id}
		log.Warn("CORE: ", err)
		return TodoItem{}, err
	}
	if len(todos) > 1 {
		log.Fatal("CORE: Multiple TodoItems with the same id.")
	}
	todo := todos[0]
	todo.Completed = completed

	log.WithFields(log.Fields{"id": id, "completed": completed}).Info("CORE: Updating TodoItem.")
	err := accessor.Update(todo)
	if err != nil {
		log.Warn("CORE: ", err)
		return TodoItem{}, err
	}
	return todo, nil
}

func DeleteItem(id int) error {
	log.WithFields(log.Fields{"id": id}).Info("CORE: Deleting TodoItem.")
	err := accessor.Delete(id)
	if err != nil {
		log.Warn("CORE: ", err)
		return err
	}
	return nil
}

func GetItems(completed bool) []TodoItem {
	log.Info("CORE: Getting TodoItems. completed=", completed)
	todos := accessor.Read(func(todo TodoItem) bool {
		return todo.Completed == completed
	})
	return todos
}
