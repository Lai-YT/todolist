package storage

import (
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"todolist/core"
)

type DatabaseAccessor struct {
	db *gorm.DB
}

type TodoItemModel struct {
	ID          int `gorm:"primary_key"`
	Description string
	Completed   bool
}

// InitDb initializes the database connection and creates the TodoItemModel table.
func (dba *DatabaseAccessor) InitDb() {
	var err error
	dba.db, err = gorm.Open(mysql.Open("root:root@/todolist?charset=utf8&parseTime=True&loc=Local"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	err = dba.db.Debug().AutoMigrate(&TodoItemModel{})
	if err != nil {
		panic(err)
	}
}

// CloseDb closes the database connection.
func (dba *DatabaseAccessor) CloseDb() {
	// NOTE: Starting from GORM v2, the db.Close() method is not available as it supports connection pooling.
	dba.db = nil
}

func (dba *DatabaseAccessor) Create(todo *core.TodoItem) (id int, e error) {
	log.WithFields(log.Fields{"description": todo.Description}).Info("DB: Adding new TodoItemModel to database.")

	result := dba.db.Create(&TodoItemModel{Description: todo.Description, Completed: false})
	if result.Error != nil {
		log.Warn("DB: ", result.Error)
		return 0, result.Error
	}

	// We access it from the database to get the Id.
	var todoModel TodoItemModel
	dba.db.Last(&todoModel)
	todo.ID = todoModel.ID
	return todoModel.ID, nil
}

func (dba *DatabaseAccessor) Read(where func(core.TodoItem) bool) []core.TodoItem {
	log.Info("DB: Reading all TodoItemModels from database.")
	// TODO: Reading all items may not be efficient.
	var todoModels []TodoItemModel
	dba.db.Find(&todoModels)

	log.Info("DB: Filtering TodoItemModels.")
	var todoItems []core.TodoItem
	for _, todoModel := range todoModels {
		if item := (core.TodoItem{ID: todoModel.ID, Description: todoModel.Description, Completed: todoModel.Completed}); where(item) {
			todoItems = append(todoItems, item)
		}
	}
	return todoItems
}

func (dba *DatabaseAccessor) Update(todo core.TodoItem) error {
	var todoModel TodoItemModel
	result := dba.db.First(&todoModel, todo.ID)
	if result.Error != nil {
		log.Warn("DB: ", result.Error)
		return result.Error
	}

	log.WithFields(log.Fields{"id": todo.ID}).Info("DB: Updating TodoItemModel.")
	todoModel.Description = todo.Description
	todoModel.Completed = todo.Completed
	dba.db.Save(&todoModel)
	return nil
}

func (dba *DatabaseAccessor) Delete(id int) error {
	var todoModel TodoItemModel
	result := dba.db.First(&todoModel, id)
	if result.Error != nil {
		log.Warn("DB: ", result.Error)
		return result.Error
	}

	log.WithFields(log.Fields{"id": id}).Info("DB: Deleting TodoItemModel.")
	dba.db.Delete(&todoModel)
	return nil
}
