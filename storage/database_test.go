package storage

import (
	"io"
	"os"
	"reflect"
	"testing"

	"todolist/core"

	log "github.com/sirupsen/logrus"
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

// expectNoError Fails the test with a fatal error if the given error is not nil.
func expectNoError(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// expectError Fails the test with a fatal error if the given error is nil.
func expectError(t *testing.T, err error) {
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

// expectEqual Fails the test with an error if the given values are not equal using the reflect.DeepEqual function.
func expectEqual(t *testing.T, want, got any) {
	if !reflect.DeepEqual(want, got) {
		t.Errorf("want: %v, got: %v", want, got)
	}
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
	expectNoError(t, err)
	if todo.ID != id {
		t.Errorf("ID not set on todo item correctly, expected %v, got %v", id, todo.ID)
	}
	want := []TodoItemModel{
		{ID: id, Description: "Test description", Completed: false},
	}
	todosInDb := []TodoItemModel{}
	dba.db.Find(&todosInDb)
	expectEqual(t, want, todosInDb)
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
	if len(got) != 1 {
		t.Fatalf("expected 1 item, got %v", len(got))
	}
	expectEqual(t, want, got[0])
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
	expectNoError(t, err)
	want := []TodoItemModel{
		{ID: 1, Description: "Test description 1", Completed: false},
		{ID: targetID, Description: updatedTodo.Description, Completed: updatedTodo.Completed},
	}
	todosInDb := []TodoItemModel{}
	dba.db.Find(&todosInDb)
	expectEqual(t, want, todosInDb)
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
	expectError(t, err)
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
	expectNoError(t, err)
	want := []TodoItemModel{
		{ID: 2, Description: "Test description 2", Completed: true},
	}
	todosInDb := []TodoItemModel{}
	dba.db.Find(&todosInDb)
	expectEqual(t, want, todosInDb)
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
	expectError(t, err)
}
