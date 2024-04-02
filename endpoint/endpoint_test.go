package endpoint

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"todolist/core"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"go.uber.org/mock/gomock"
)

func TestMain(m *testing.M) {
	log.SetOutput(io.Discard)
	code := m.Run()
	os.Exit(code)
}

// testEnv is a test environment that contains common test resources and implements common test functions.
type testEnv struct {
	t        *testing.T
	router   *mux.Router
	ctrl     *gomock.Controller
	mockCore *MockCore
	writer   *httptest.ResponseRecorder
}

// newTestEnv Sets up the test environment by creating a router, a http test writer, and a mock Core.
func newTestEnv(t *testing.T) *testEnv {
	ctrl := gomock.NewController(t)
	mockCore := NewMockCore(ctrl)
	SetCore(mockCore)
	return &testEnv{
		t:        t,
		router:   mux.NewRouter(),
		ctrl:     ctrl,
		mockCore: mockCore,
		writer:   httptest.NewRecorder(),
	}
}

// expectStatusCodeToBe Checks if the status code of the response is the same as the given code. If not, the test will fail as an error.
func (e *testEnv) expectStatusCodeToBe(code int) {
	if e.writer.Code != code {
		e.t.Errorf("expected status code %v, got %v", code, e.writer.Code)
	}
}

// expectUnmarshalWithoutError Unmarshals the response body into the given value. If there is an error, the test will fail as a fatal error.
func (e *testEnv) expectUnmarshalWithoutError(v any) {
	if err := json.Unmarshal(e.writer.Body.Bytes(), v); err != nil {
		e.t.Fatalf("error unmarshalling response body: %v", err)
	}
}

// expectEqual Checks if the two values are equal with reflect.DeepEqual. If not, the test will fail as an error.
func (e *testEnv) expectEqual(expected, got any) {
	if !reflect.DeepEqual(expected, got) {
		e.t.Errorf("expected %v, got %v", expected, got)
	}
}

// TestHealthz Given the Healthz handler serve at the /healthz endpoint, when a request is made to the endpoint, then the server should respond with a 200 status code and a JSON response body.
func TestHealthz(t *testing.T) {
	// arrange
	e := newTestEnv(t)
	pattern := "/healthz"
	e.router.HandleFunc(pattern, Healthz)

	// act: make a request to the /healthz endpoint
	request, _ := http.NewRequest(http.MethodGet, pattern, nil)
	e.router.ServeHTTP(e.writer, request)

	// assert
	e.expectStatusCodeToBe(http.StatusOK)
	type body struct {
		Alive bool `json:"alive"`
	}
	want := body{Alive: true}
	got := body{}
	e.expectUnmarshalWithoutError(&got)
	e.expectEqual(want, got)
}

// TestCreateItem Give the CreateItem handler serve at the /todo endpoint, when a request is made to the endpoint with a description form parameter, then the server should respond with a 200 status code and a JSON response body describing the newly created TodoItem.
func TestCreateItem(t *testing.T) {
	// arrange
	e := newTestEnv(t)
	pattern := "/todo"
	e.router.HandleFunc(pattern, CreateItem)
	testDescription := "test"
	e.mockCore.EXPECT().
		CreateItem(testDescription).
		Return(core.TodoItem{Id: 1, Description: testDescription, Completed: false})

	// act
	params := url.Values{
		"description": []string{testDescription},
	}
	request, _ := http.NewRequest(http.MethodPost, pattern, strings.NewReader(params.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	e.router.ServeHTTP(e.writer, request)

	// assert
	e.expectStatusCodeToBe(http.StatusOK)
	want := core.TodoItem{Id: 1, Description: testDescription, Completed: false}
	got := core.TodoItem{}
	e.expectUnmarshalWithoutError(&got)
	e.expectEqual(want, got)
}

// TestUpdateItem Given the UpdateItem handler serve at the /todo/{id} endpoint and the core returns without error, when a request is made to the endpoint with a completed form parameter, then the server should respond with a 200 status code and a JSON response body indicating that the update was successful.
func TestUpdateItem(t *testing.T) {
	// arrange
	e := newTestEnv(t)
	pattern := "/todo/{id}"
	e.router.HandleFunc(pattern, UpdateItem)
	testId := 1
	testCompleted := true
	e.mockCore.EXPECT().
		UpdateItem(testId, testCompleted).
		Return(core.TodoItem{Id: testId} /* dummy */, nil)

	// act
	params := url.Values{
		"completed": []string{`true`},
	}
	request, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/todo/%d", testId), strings.NewReader(params.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	e.router.ServeHTTP(e.writer, request)

	// assert
	e.expectStatusCodeToBe(http.StatusOK)
	want := map[string]json.RawMessage{"updated": []byte(`true`)}
	got := map[string]json.RawMessage{}
	e.expectUnmarshalWithoutError(&got)
	e.expectEqual(want, got)
}

// TestUpdateItemError Given the UpdateItem handler serve at the /todo/{id} endpoint and the core returns an error, when a request is made to the endpoint with a completed form parameter, then the server should respond with a 200 status code and a JSON response body indicating that the update was not successful.
func TestUpdateItemError(t *testing.T) {
	// arrange
	e := newTestEnv(t)
	pattern := "/todo/{id}"
	e.router.HandleFunc(pattern, UpdateItem)
	e.mockCore.EXPECT().
		UpdateItem(gomock.Any(), gomock.Any()).
		Return(core.TodoItem{} /* dummy */, errors.New("test error"))

	// act
	params := url.Values{
		"completed": []string{`true`},
	}
	request, _ := http.NewRequest(http.MethodPost, "/todo/1", strings.NewReader(params.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	e.router.ServeHTTP(e.writer, request)

	// assert
	e.expectStatusCodeToBe(http.StatusOK)
	// NOTE: We do not check the error message because it is not guaranteed to be the same as the one returned by the Core.
	type body struct {
		Updated bool `json:"updated"`
	}
	want := body{Updated: false}
	got := body{}
	e.expectUnmarshalWithoutError(&got)
	e.expectEqual(want, got)
}

// TestGetItem Given the GetItem handler serve at the /todo/{id} endpoint and the core returns without error, when a request is made to the endpoint, then the server should respond with a 200 status code and a JSON response body indicating that the deletion was successful.
func TestDeleteItem(t *testing.T) {
	// arrange
	e := newTestEnv(t)
	pattern := "/todo/{id}"
	e.router.HandleFunc(pattern, DeleteItem)
	testId := 1
	e.mockCore.EXPECT().
		DeleteItem(testId).
		Return(nil)

	// act
	request, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/todo/%d", testId), strings.NewReader(""))
	e.router.ServeHTTP(e.writer, request)

	// assert
	e.expectStatusCodeToBe(http.StatusOK)
	want := map[string]json.RawMessage{"deleted": []byte(`true`)}
	got := map[string]json.RawMessage{}
	e.expectUnmarshalWithoutError(&got)
	e.expectEqual(want, got)
}

// TestDeleteItemError Given the DeleteItem handler serve at the /todo/{id} endpoint and the core returns an error, when a request is made to the endpoint, then the server should respond with a 200 status code and a JSON response body indicating that the deletion was not successful.
func TestDeleteItemError(t *testing.T) {
	// arrange
	e := newTestEnv(t)
	pattern := "/todo/{id}"
	e.router.HandleFunc(pattern, DeleteItem)
	e.mockCore.EXPECT().
		DeleteItem(gomock.Any()).
		Return(errors.New("test error"))

	// act
	request, _ := http.NewRequest(http.MethodDelete, "/todo/1", strings.NewReader(""))
	e.router.ServeHTTP(e.writer, request)

	// assert
	e.expectStatusCodeToBe(http.StatusOK)
	// NOTE: We do not check the error message because it is not guaranteed to be the same as the one returned by the Core.
	type body struct {
		Deleted bool `json:"deleted"`
	}
	want := body{Deleted: false}
	got := body{}
	e.expectUnmarshalWithoutError(&got)
	e.expectEqual(want, got)
}

func TestGetItemsCompleted(t *testing.T) {
	// arrange
	e := newTestEnv(t)
	pattern := "/todo"
	e.router.HandleFunc(pattern, GetItems)
	todoItems := []core.TodoItem{
		{Id: 1, Description: "test1", Completed: true},
		{Id: 3, Description: "test3", Completed: true},
	}
	e.mockCore.EXPECT().
		GetItems(true).
		Return(todoItems)

	// act
	completed := true
	request, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/todo?completed=%t", completed), strings.NewReader(""))
	e.router.ServeHTTP(e.writer, request)

	// assert
	e.expectStatusCodeToBe(http.StatusOK)
	want := todoItems
	got := []core.TodoItem{}
	e.expectUnmarshalWithoutError(&got)
	e.expectEqual(want, got)
}

func TestGetItemIncomplete(t *testing.T) {
	// arrange
	e := newTestEnv(t)
	pattern := "/todo"
	e.router.HandleFunc(pattern, GetItems)
	todoItems := []core.TodoItem{
		{Id: 2, Description: "test2", Completed: false},
		{Id: 4, Description: "test4", Completed: false},
	}
	e.mockCore.EXPECT().
		GetItems(false).
		Return(todoItems)

	// act
	completed := false
	request, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/todo?completed=%t", completed), strings.NewReader(""))
	e.router.ServeHTTP(e.writer, request)

	// assert
	e.expectStatusCodeToBe(http.StatusOK)
	want := todoItems
	got := []core.TodoItem{}
	e.expectUnmarshalWithoutError(&got)
	e.expectEqual(want, got)
}

func TestGetItemsAll(t *testing.T) {
	// arrange
	e := newTestEnv(t)
	pattern := "/todo"
	e.router.HandleFunc(pattern, GetItems)
	todoItems := []core.TodoItem{
		{Id: 1, Description: "test1", Completed: true},
		{Id: 2, Description: "test2", Completed: false},
		{Id: 3, Description: "test3", Completed: true},
		{Id: 4, Description: "test4", Completed: false},
	}
	e.mockCore.EXPECT().
		GetItems(true).
		Return([]core.TodoItem{todoItems[0], todoItems[2]}).
		MaxTimes(1)
	e.mockCore.EXPECT().
		GetItems(false).
		Return([]core.TodoItem{todoItems[1], todoItems[3]}).
		MaxTimes(1)

	// act
	request, _ := http.NewRequest(http.MethodGet, "/todo", strings.NewReader(""))
	e.router.ServeHTTP(e.writer, request)

	// assert
	e.expectStatusCodeToBe(http.StatusOK)
	want := todoItems
	got := []core.TodoItem{}
	e.expectUnmarshalWithoutError(&got)
	// NOTE: Sort the slices before comparing them because the order of the items is not guaranteed.
	sort.Slice(got, func(i, j int) bool { return got[i].Id < got[j].Id })
	e.expectEqual(want, got)
}
