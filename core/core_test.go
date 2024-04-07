package core

import (
	"errors"
	"io"
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestMain(m *testing.M) {
	// So that we don't see log messages during tests.
	log.SetOutput(io.Discard)
	code := m.Run()
	os.Exit(code)
}

// testEnv is a test environment that contains common test test resources and implements common test functions.
type testEnv struct {
	t            *testing.T
	ctrl         *gomock.Controller
	mockAccessor *MockStorageAccessor
	core         *TheCore
}

func newTestEnv(t *testing.T) *testEnv {
	ctrl := gomock.NewController(t)
	mockAccessor := NewMockStorageAccessor(ctrl)
	theCore := NewCore(mockAccessor)
	return &testEnv{t, ctrl, mockAccessor, theCore}
}

// TestCreateItem Given a description and the storage accessor returns an id, when CreateItem is called, then the item is created and returned with the id set.
func TestCreateItem(t *testing.T) {
	// arrange
	e := newTestEnv(t)
	e.mockAccessor.EXPECT().
		Create(gomock.Any()).
		DoAndReturn(func(item *TodoItem) (int, error) {
			id := 1
			item.ID = id
			return id, nil
		})

	// act
	want := TodoItem{ID: 1, Description: "some description", Completed: false}
	got := e.core.CreateItem(want.Description)

	// assert
	assert.Equal(t, want, got)
}

// TestUpdateItem Given an item of a specific id is returned by the storage accessor, when UpdateItem is called, then the item is updated and returned with the new completed status.
func TestUpdateItem(t *testing.T) {
	// arrange
	e := newTestEnv(t)
	e.mockAccessor.EXPECT().
		Read(gomock.Any()).
		DoAndReturn(func(func(TodoItem) bool) []TodoItem {
			return []TodoItem{
				{ID: 1, Description: "some description", Completed: false},
			}
		})
	e.mockAccessor.EXPECT().
		Update(gomock.Any()).
		Return(nil)

	// act
	want := TodoItem{ID: 1, Description: "some description", Completed: true}
	got, err := e.core.UpdateItem(want.ID, want.Completed)

	// assert: the item should be updated and returned without error
	if assert.NoError(t, err) {
		assert.Equal(t, want, got)
	}
}

// TestUpdateItemNotFound Given an item of a specific id is not returned by the storage accessor, when UpdateItem is called, then an ItemNotFoundError is returned.
func TestUpdateItemNotFound(t *testing.T) {
	// arrange
	e := newTestEnv(t)
	e.mockAccessor.EXPECT().
		Read(gomock.Any()).
		DoAndReturn(func(func(TodoItem) bool) []TodoItem {
			return []TodoItem{}
		})

	// act
	id := 1
	completed := true
	_, err := e.core.UpdateItem(id, completed)

	// assert: an error should be returned
	assert.IsType(t, TodoItemNotFoundError{}, err)
}

// TestDeleteItem Given an id and the storage accessor returns no error, when DeleteItem is called, then no error is returned.
func TestDeleteItem(t *testing.T) {
	// arrange
	e := newTestEnv(t)
	e.mockAccessor.EXPECT().
		Delete(gomock.Any()).
		Return(nil)

	// act
	id := 1
	err := e.core.DeleteItem(id)

	// assert
	assert.NoError(t, err)
}

// TestDeleteItemError Given an id and the storage accessor returns an error, when DeleteItem is called, then the error is returned.
func TestDeleteItemError(t *testing.T) {
	// arrange
	e := newTestEnv(t)
	e.mockAccessor.EXPECT().
		Delete(gomock.Any()).
		Return(errors.New("error"))

	// act
	id := 1
	err := e.core.DeleteItem(id)

	// assert
	assert.Error(t, err)
	// NOTE: There's no guarantee that the error is the same error that was returned by the storage accessor.
}

// TestGetItems Given items are returned by the storage accessor, when GetItems is called, then the items are returned.
func TestGetItems(t *testing.T) {
	// arrange
	e := newTestEnv(t)
	mockItems := [2]TodoItem{
		{ID: 1, Description: "some description", Completed: false},
		{ID: 2, Description: "another description", Completed: true},
	}
	e.mockAccessor.EXPECT().
		Read(gomock.Any()).
		DoAndReturn(func(func(TodoItem) bool) []TodoItem {
			// With completed = false.
			return []TodoItem{mockItems[0]}
		})

	// act
	completed := false
	want := []TodoItem{mockItems[0]}
	got := e.core.GetItems(completed)

	// assert
	assert.Equal(t, want, got)
}
