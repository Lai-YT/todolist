package storage

import (
	"io"
	"os"
	"testing"

	"todolist/core"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestMain(m *testing.M) {
	// So that we don't see log messages during tests.
	log.SetOutput(io.Discard)
	code := m.Run()
	os.Exit(code)
}

// NOTE: Errors on the database panics because it means the test setup is incorrect.

func (dba *DatabaseAccessor) initTestDb() {
	var err error
	// NOTE: Using the in-memory SQLite database for testing purposes.
	dba.db, err = gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{
		Logger: logger.Discard,
	})
	if err != nil {
		panic(err)
	}
	err = dba.db.AutoMigrate(&TodoItemModel{})
	if err != nil {
		panic(err)
	}
}

func (dba *DatabaseAccessor) closeTestDb() {
	err := dba.db.Migrator().DropTable(&TodoItemModel{})
	if err != nil {
		panic(err)
	}
	dba.db = nil
}

// TestCreate Given a todo item, when Create is called, then the todo item should be created in the database and the id should be set and returned.
func TestCreate(t *testing.T) {
	// arrange
	dba := DatabaseAccessor{}
	dba.initTestDb()
	defer dba.closeTestDb()

	// act
	todo := core.TodoItem{Description: "Test description", Completed: false}
	id, err := dba.Create(&todo)

	// assert
	if assert.NoError(t, err) {
		assert.Equal(t, id, todo.ID, "ID not set on todo item correctly")
		want := []TodoItemModel{
			{ID: id, Description: todo.Description, Completed: todo.Completed},
		}
		todosInDb := []TodoItemModel{}
		dba.db.Find(&todosInDb)
		assert.Equal(t, want, todosInDb)
	}
}

// TestRead Given some todo items in the database, when Read is called with a where clause that matches on the description of a todo item, then the todo item should be returned.
func TestRead(t *testing.T) {
	// arrange
	dba := DatabaseAccessor{}
	dba.initTestDb()
	defer dba.closeTestDb()
	match := "Test description 1"
	dba.db.Create(&[]TodoItemModel{
		{ID: 1, Description: match, Completed: false},
		{ID: 2, Description: "Test description 2", Completed: true},
	})

	// act
	want := core.TodoItem{ID: 1, Description: "Test description 1", Completed: false}
	got := dba.Read(func(item core.TodoItem) bool { return item.Description == match })

	// assert
	if assert.Len(t, got, 1) {
		assert.Equal(t, want, got[0])
	}
}

// TestUpdate Given some todo items in the database, when Update is called with the id of a todo item, then the todo item should be updated.
func TestUpdate(t *testing.T) {
	// arrange
	dba := DatabaseAccessor{}
	dba.initTestDb()
	defer dba.closeTestDb()
	targetID := 2
	dba.db.Create(&[]TodoItemModel{
		{ID: 1, Description: "Test description 1", Completed: false},
		{ID: targetID, Description: "Test description 2", Completed: true},
	})

	// act
	updatedTodo := core.TodoItem{ID: targetID, Description: "Updated description", Completed: false}
	err := dba.Update(updatedTodo)

	// assert
	if assert.NoError(t, err) {
		want := []TodoItemModel{
			{ID: 1, Description: "Test description 1", Completed: false},
			{ID: targetID, Description: updatedTodo.Description, Completed: updatedTodo.Completed},
		}
		todosInDb := []TodoItemModel{}
		dba.db.Find(&todosInDb)
		assert.Equal(t, want, todosInDb)
	}
}

// TestUpdateNotFound Given some todo items in the database, when Update is called with an id that does not exist, then an error should be returned.
func TestUpdateNotFound(t *testing.T) {
	// arrange
	dba := DatabaseAccessor{}
	dba.initTestDb()
	defer dba.closeTestDb()
	dba.db.Create(&[]TodoItemModel{
		{ID: 1, Description: "Test description 1", Completed: false},
		{ID: 2, Description: "Test description 2", Completed: true},
	})

	// act
	nonExistentTodo := core.TodoItem{ID: 3, Description: "Updated description", Completed: false}
	err := dba.Update(nonExistentTodo)

	// assert
	assert.Error(t, err)
}

// TestDelete Given some todo items in the database, when Delete is called with the id of a todo item, then the todo item should be deleted.
func TestDelete(t *testing.T) {
	// arrange
	dba := DatabaseAccessor{}
	dba.initTestDb()
	defer dba.closeTestDb()
	dba.db.Create(&[]TodoItemModel{
		{ID: 1, Description: "Test description 1", Completed: false},
		{ID: 2, Description: "Test description 2", Completed: true},
	})

	// act
	err := dba.Delete(1)

	// assert
	if assert.NoError(t, err) {
		want := []TodoItemModel{
			{ID: 2, Description: "Test description 2", Completed: true},
		}
		todosInDb := []TodoItemModel{}
		dba.db.Find(&todosInDb)
		assert.Equal(t, want, todosInDb)
	}
}

// TestDeleteNotFound Given some todo items in the database, when Delete is called with an id that does not exist, then an error should be returned.
func TestDeleteNotFound(t *testing.T) {
	// arrange
	dba := DatabaseAccessor{}
	dba.initTestDb()
	defer dba.closeTestDb()
	items := []TodoItemModel{
		{ID: 1, Description: "Test description 1", Completed: false},
		{ID: 2, Description: "Test description 2", Completed: true},
	}
	dba.db.Create(&items)

	// act
	err := dba.Delete(3)

	// assert
	assert.Error(t, err)
}
