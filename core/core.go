package core

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	log "github.com/sirupsen/logrus"
)

var db *gorm.DB

// InitDb initializes the database connection and creates the TodoItemModel table.
func InitDb() {
	var err error
	db, err = gorm.Open("mysql", "root:root@/todolist?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		log.Fatal(err)
	}
	// TODO: Keep the table.
	db.Debug().DropTableIfExists(&TodoItemModel{})
	db.Debug().AutoMigrate(&TodoItemModel{})
}

// CloseDb closes the database connection.
func CloseDb() {
	db.Close()
}

// NOTE: We separate the core data structure with the database model.

type TodoItem struct {
	Id          int
	Description string
	Completed   bool
}

type TodoItemModel struct {
	Id          int `gorm:"primary_key"`
	Description string
	Completed   bool
}

func CreateItem(description string) TodoItem {
	log.WithFields(log.Fields{"description": description}).Info("Adding new TodoItem. Saving to database.")

	db.Create(&TodoItemModel{Description: description, Completed: false})

	// We access it from the database to get the Id.
	var todoModel TodoItemModel
	db.Last(&todoModel)
	return TodoItem{Id: todoModel.Id, Description: todoModel.Description, Completed: todoModel.Completed}
}

type TodoItemNotFoundError struct {
	Id int
}

func (e TodoItemNotFoundError) Error() string {
	return fmt.Sprintf("TodoItem with id %d not found", e.Id)
}

func UpdateItem(id int, completed bool) (TodoItem, error) {
	var todoModel TodoItemModel
	result := db.First(&todoModel, id)
	if result.Error != nil {
		log.Warn("TodoItem not found in database")
		return TodoItem{}, TodoItemNotFoundError{Id: id}
	}

	log.WithFields(log.Fields{"id": id, "completed": completed}).Info("Updating TodoItem.")
	todoModel.Completed = completed
	db.Save(&todoModel)
	return TodoItem{Id: todoModel.Id, Description: todoModel.Description, Completed: todoModel.Completed}, nil
}

func DeleteItem(id int) error {
	var todoModel TodoItemModel
	result := db.First(&todoModel, id)
	if result.Error != nil {
		log.Warn("TodoItem not found in database")
		return TodoItemNotFoundError{Id: id}
	}

	log.WithFields(log.Fields{"id": id}).Info("Deleting TodoItem.")
	db.Delete(&todoModel)
	return nil
}

func GetItems(completed bool) []TodoItem {
	log.Info("Get TodoItems, completed=", completed)
	var todoModels []TodoItemModel
	db.Where("completed = ?", completed).Find(&todoModels)
	todoItems := make([]TodoItem, len(todoModels))
	for i, todoModel := range todoModels {
		todoItems[i] = TodoItem{Id: todoModel.Id, Description: todoModel.Description, Completed: todoModel.Completed}
	}
	return todoItems
}
