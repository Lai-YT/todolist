package core

import (
	"errors"
	"io"
	"os"
	"reflect"
	"testing"

	log "github.com/sirupsen/logrus"
	"go.uber.org/mock/gomock"
)

func TestMain(m *testing.M) {
	// So that we don't see log messages during tests.
	log.SetOutput(io.Discard)
	code := m.Run()
	teardown()
	os.Exit(code)
}

func teardown() {
	SetAccessor(nil)
}

// TestCreateItem Given a description and the storage accessor returns an id, when CreateItem is called, then the item is created and returned with the id set.
func TestCreateItem(t *testing.T) {
	// arrange: mock storage accessor
	ctrl := gomock.NewController(t)
	mockAccessor := NewMockStorageAccessor(ctrl)
	SetAccessor(mockAccessor)
	mockAccessor.EXPECT().
		Create(gomock.Any()).
		DoAndReturn(func(item *TodoItem) (int, error) {
			id := 1
			item.Id = id
			return id, nil
		})

	// act
	want := TodoItem{Id: 1, Description: "some description", Completed: false}
	got := CreateItem(want.Description)

	// assert
	if !reflect.DeepEqual(want, got) {
		t.Errorf("want: %v, got: %v", want, got)
	}
}

// TestUpdateItem Given an item of a specific id is returned by the storage accessor, when UpdateItem is called, then the item is updated and returned with the new completed status.
func TestUpdateItem(t *testing.T) {
	// arrange: mock storage accessor
	ctrl := gomock.NewController(t)
	mockAccessor := NewMockStorageAccessor(ctrl)
	SetAccessor(mockAccessor)
	mockAccessor.EXPECT().
		Read(gomock.Any()).
		DoAndReturn(func(func(TodoItem) bool) []TodoItem {
			return []TodoItem{
				{Id: 1, Description: "some description", Completed: false},
			}
		})
	mockAccessor.EXPECT().
		Update(gomock.Any()).
		Return(nil)

	// act:
	want := TodoItem{Id: 1, Description: "some description", Completed: true}
	got, err := UpdateItem(want.Id, want.Completed)

	// assert: the item should be updated and returned without error
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(want, got) {
		t.Errorf("want: %v, got: %v", want, got)
	}
}

// TestUpdateItemNotFound Given an item of a specific id is not returned by the storage accessor, when UpdateItem is called, then an ItemNotFoundError is returned.
func TestUpdateItemNotFound(t *testing.T) {
	// arrange: mock storage accessor
	ctrl := gomock.NewController(t)
	mockAccessor := NewMockStorageAccessor(ctrl)
	SetAccessor(mockAccessor)
	mockAccessor.EXPECT().
		Read(gomock.Any()).
		DoAndReturn(func(func(TodoItem) bool) []TodoItem {
			return []TodoItem{}
		})

	// act:
	id := 1
	completed := true
	_, err := UpdateItem(id, completed)

	// assert: an error should be returned
	if err == nil {
		t.Errorf("expected error")
	}
	if reflect.TypeOf(err).String() != "core.TodoItemNotFoundError" {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestDeleteItem Given an id and the storage accessor returns no error, when DeleteItem is called, then no error is returned.
func TestDeleteItem(t *testing.T) {
	// arrange: mock storage accessor
	ctrl := gomock.NewController(t)
	mockAccessor := NewMockStorageAccessor(ctrl)
	SetAccessor(mockAccessor)
	mockAccessor.EXPECT().
		Delete(gomock.Any()).
		Return(nil)

	// act
	id := 1
	err := DeleteItem(id)

	// assert
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestDeleteItemError Given an id and the storage accessor returns an error, when DeleteItem is called, then the error is returned.
func TestDeleteItemError(t *testing.T) {
	// arrange: mock storage accessor
	ctrl := gomock.NewController(t)
	mockAccessor := NewMockStorageAccessor(ctrl)
	SetAccessor(mockAccessor)
	mockAccessor.EXPECT().
		Delete(gomock.Any()).
		Return(errors.New("error"))

	// act
	id := 1
	err := DeleteItem(id)

	// assert
	if err == nil {
		t.Errorf("expected error")
	}
	// NOTE: There's no guarantee that the error is the same error that was returned by the storage accessor.
}

// TestGetItems Given items are returned by the storage accessor, when GetItems is called, then the items are returned.
func TestGetItems(t *testing.T) {
	// arrange: mock storage accessor
	ctrl := gomock.NewController(t)
	mockAccessor := NewMockStorageAccessor(ctrl)
	SetAccessor(mockAccessor)
	mockItems := [2]TodoItem{
		{Id: 1, Description: "some description", Completed: false},
		{Id: 2, Description: "another description", Completed: true},
	}
	mockAccessor.EXPECT().
		Read(gomock.Any()).
		DoAndReturn(func(func(TodoItem) bool) []TodoItem {
			// With completed = false.
			return []TodoItem{mockItems[0]}
		})

	// act
	completed := false
	want := []TodoItem{mockItems[0]}
	got := GetItems(completed)

	// assert
	if !reflect.DeepEqual(want, got) {
		t.Errorf("want: %v, got: %v", want, got)
	}
}
