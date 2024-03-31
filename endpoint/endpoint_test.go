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

// TestHealthz Given the Healthz handler serve at the /healthz endpoint, when a request is made to the endpoint, then the server should respond with a 200 status code and a JSON response body.
func TestHealthz(t *testing.T) {
	// arrange
	router := mux.NewRouter()
	pattern := "/healthz"
	router.HandleFunc(pattern, Healthz)

	// act: make a request to the /healthz endpoint
	request, _ := http.NewRequest(http.MethodGet, pattern, nil)
	writer := httptest.NewRecorder()
	router.ServeHTTP(writer, request)

	// assert
	if writer.Code != http.StatusOK {
		t.Errorf("expected status code %v, got %v", http.StatusOK, writer.Code)
	}
	type body struct {
		Alive bool `json:"alive"`
	}
	want := body{Alive: true}
	got := body{}
	err := json.Unmarshal(writer.Body.Bytes(), &got)
	if err != nil {
		t.Fatalf("error unmarshalling response body: %v", err)
	}
	if got != want {
		t.Errorf("expected body %v, got %v", want, got)
	}
}

// TestCreateItem Give the CreateItem handler serve at the /todo endpoint, when a request is made to the endpoint with a description form parameter, then the server should respond with a 200 status code and a JSON response body describing the newly created TodoItem.
func TestCreateItem(t *testing.T) {
	// arrange
	router := mux.NewRouter()
	pattern := "/todo"
	router.HandleFunc(pattern, CreateItem)
	ctrl := gomock.NewController(t)
	mockCore := NewMockCore(ctrl)
	SetCore(mockCore)
	testDescription := "test"
	mockCore.EXPECT().
		CreateItem(testDescription).
		Return(core.TodoItem{Id: 1, Description: testDescription, Completed: false})

	// act
	params := url.Values{
		"description": []string{testDescription},
	}
	request, _ := http.NewRequest(http.MethodPost, pattern, strings.NewReader(params.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	writer := httptest.NewRecorder()
	router.ServeHTTP(writer, request)

	// assert
	if writer.Code != http.StatusOK {
		t.Errorf("expected status code %v, got %v", http.StatusOK, writer.Code)
	}
	want := core.TodoItem{Id: 1, Description: testDescription, Completed: false}
	got := core.TodoItem{}
	err := json.Unmarshal(writer.Body.Bytes(), &got)
	if err != nil {
		t.Fatalf("error unmarshalling response body: %v", err)
	}
	if got != want {
		t.Errorf("expected body %v, got %v", want, got)
	}
}

// TestUpdateItem Given the UpdateItem handler serve at the /todo/{id} endpoint and the core returns without error, when a request is made to the endpoint with a completed form parameter, then the server should respond with a 200 status code and a JSON response body indicating that the update was successful.
func TestUpdateItem(t *testing.T) {
	// arrange
	router := mux.NewRouter()
	pattern := "/todo/{id}"
	router.HandleFunc(pattern, UpdateItem)
	ctrl := gomock.NewController(t)
	mockCore := NewMockCore(ctrl)
	SetCore(mockCore)
	testId := 1
	testCompleted := true
	mockCore.EXPECT().
		UpdateItem(testId, testCompleted).
		Return(core.TodoItem{Id: testId} /* dummy */, nil)

	// act
	params := url.Values{
		"completed": []string{`true`},
	}
	request, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/todo/%d", testId), strings.NewReader(params.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	writer := httptest.NewRecorder()
	router.ServeHTTP(writer, request)

	// assert
	if writer.Code != http.StatusOK {
		t.Errorf("expected status code %v, got %v", http.StatusOK, writer.Code)
	}
	want := map[string]json.RawMessage{"updated": []byte(`true`)}
	got := map[string]json.RawMessage{}
	err := json.Unmarshal(writer.Body.Bytes(), &got)
	if err != nil {
		t.Fatalf("error unmarshalling response body: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("expected body %v, got %v", want, got)
	}
}

// TestUpdateItemError Given the UpdateItem handler serve at the /todo/{id} endpoint and the core returns an error, when a request is made to the endpoint with a completed form parameter, then the server should respond with a 200 status code and a JSON response body indicating that the update was not successful.
func TestUpdateItemError(t *testing.T) {
	// arrange
	router := mux.NewRouter()
	pattern := "/todo/{id}"
	router.HandleFunc(pattern, UpdateItem)
	ctrl := gomock.NewController(t)
	mockCore := NewMockCore(ctrl)
	SetCore(mockCore)
	mockCore.EXPECT().
		UpdateItem(gomock.Any(), gomock.Any()).
		Return(core.TodoItem{} /* dummy */, errors.New("test error"))

	// act
	params := url.Values{
		"completed": []string{`true`},
	}
	request, _ := http.NewRequest(http.MethodPost, "/todo/1", strings.NewReader(params.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	writer := httptest.NewRecorder()
	router.ServeHTTP(writer, request)

	// assert
	if writer.Code != http.StatusOK {
		t.Errorf("expected status code %v, got %v", http.StatusOK, writer.Code)
	}
	// NOTE: We do not check the error message because it is not guaranteed to be the same as the one returned by the Core.
	type body struct {
		Updated bool `json:"updated"`
	}
	want := body{Updated: false}
	got := body{}
	err := json.Unmarshal(writer.Body.Bytes(), &got)
	if err != nil {
		t.Fatalf("error unmarshalling response body: %v", err)
	}
	if got != want {
		t.Errorf("expected body %v, got %v", want, got)
	}
}

// TestGetItem Given the GetItem handler serve at the /todo/{id} endpoint and the core returns without error, when a request is made to the endpoint, then the server should respond with a 200 status code and a JSON response body indicating that the deletion was successful.
func TestDeleteItem(t *testing.T) {
	// arrange
	router := mux.NewRouter()
	pattern := "/todo/{id}"
	router.HandleFunc(pattern, DeleteItem)
	ctrl := gomock.NewController(t)
	mockCore := NewMockCore(ctrl)
	SetCore(mockCore)
	testId := 1
	mockCore.EXPECT().
		DeleteItem(testId).
		Return(nil)

	// act
	request, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/todo/%d", testId), strings.NewReader(""))
	writer := httptest.NewRecorder()
	router.ServeHTTP(writer, request)

	// assert
	if writer.Code != http.StatusOK {
		t.Errorf("expected status code %v, got %v", http.StatusOK, writer.Code)
	}
	want := map[string]json.RawMessage{"deleted": []byte(`true`)}
	got := map[string]json.RawMessage{}
	err := json.Unmarshal(writer.Body.Bytes(), &got)
	if err != nil {
		t.Fatalf("error unmarshalling response body: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("expected body %v, got %v", want, got)
	}
}

// TestDeleteItemError Given the DeleteItem handler serve at the /todo/{id} endpoint and the core returns an error, when a request is made to the endpoint, then the server should respond with a 200 status code and a JSON response body indicating that the deletion was not successful.
func TestDeleteItemError(t *testing.T) {
	// arrange
	router := mux.NewRouter()
	pattern := "/todo/{id}"
	router.HandleFunc(pattern, DeleteItem)
	ctrl := gomock.NewController(t)
	mockCore := NewMockCore(ctrl)
	SetCore(mockCore)
	mockCore.EXPECT().
		DeleteItem(gomock.Any()).
		Return(errors.New("test error"))

	// act
	request, _ := http.NewRequest(http.MethodDelete, "/todo/1", strings.NewReader(""))
	writer := httptest.NewRecorder()
	router.ServeHTTP(writer, request)

	// assert
	if writer.Code != http.StatusOK {
		t.Errorf("expected status code %v, got %v", http.StatusOK, writer.Code)
	}
	// NOTE: We do not check the error message because it is not guaranteed to be the same as the one returned by the Core.
	type body struct {
		Deleted bool `json:"deleted"`
	}
	want := body{Deleted: false}
	got := body{}
	err := json.Unmarshal(writer.Body.Bytes(), &got)
	if err != nil {
		t.Fatalf("error unmarshalling response body: %v", err)
	}
	if got != want {
		t.Errorf("expected body %v, got %v", want, got)
	}
}

func TestGetItemsCompleted(t *testing.T) {
	// arrange
	router := mux.NewRouter()
	pattern := "/todo"
	router.HandleFunc(pattern, GetItems)
	ctrl := gomock.NewController(t)
	mockCore := NewMockCore(ctrl)
	SetCore(mockCore)
	todoItems := []core.TodoItem{
		{Id: 1, Description: "test1", Completed: true},
		{Id: 3, Description: "test3", Completed: true},
	}
	mockCore.EXPECT().
		GetItems(true).
		Return(todoItems)

	// act
	completed := true
	request, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/todo?completed=%t", completed), strings.NewReader(""))
	writer := httptest.NewRecorder()
	router.ServeHTTP(writer, request)

	// assert
	if writer.Code != http.StatusOK {
		t.Errorf("expected status code %v, got %v", http.StatusOK, writer.Code)
	}
	want := todoItems
	got := []core.TodoItem{}
	err := json.Unmarshal(writer.Body.Bytes(), &got)
	if err != nil {
		t.Fatalf("error unmarshalling response body: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("expected body %v, got %v", want, got)
	}
}

func TestGetItemIncomplete(t *testing.T) {
	// arrange
	router := mux.NewRouter()
	pattern := "/todo"
	router.HandleFunc(pattern, GetItems)
	ctrl := gomock.NewController(t)
	mockCore := NewMockCore(ctrl)
	SetCore(mockCore)
	todoItems := []core.TodoItem{
		{Id: 2, Description: "test2", Completed: false},
		{Id: 4, Description: "test4", Completed: false},
	}
	mockCore.EXPECT().
		GetItems(false).
		Return(todoItems)

	// act
	completed := false
	request, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/todo?completed=%t", completed), strings.NewReader(""))
	writer := httptest.NewRecorder()
	router.ServeHTTP(writer, request)

	// assert
	if writer.Code != http.StatusOK {
		t.Errorf("expected status code %v, got %v", http.StatusOK, writer.Code)
	}
	want := todoItems
	got := []core.TodoItem{}
	err := json.Unmarshal(writer.Body.Bytes(), &got)
	if err != nil {
		t.Fatalf("error unmarshalling response body: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("expected body %v, got %v", want, got)
	}
}

func TestGetItemsAll(t *testing.T) {
	// arrange
	router := mux.NewRouter()
	pattern := "/todo"
	router.HandleFunc(pattern, GetItems)
	ctrl := gomock.NewController(t)
	mockCore := NewMockCore(ctrl)
	SetCore(mockCore)
	todoItems := []core.TodoItem{
		{Id: 1, Description: "test1", Completed: true},
		{Id: 2, Description: "test2", Completed: false},
		{Id: 3, Description: "test3", Completed: true},
		{Id: 4, Description: "test4", Completed: false},
	}
	mockCore.EXPECT().
		GetItems(true).
		Return([]core.TodoItem{todoItems[0], todoItems[2]}).
		MaxTimes(1)
	mockCore.EXPECT().
		GetItems(false).
		Return([]core.TodoItem{todoItems[1], todoItems[3]}).
		MaxTimes(1)

	// act
	request, _ := http.NewRequest(http.MethodGet, "/todo", strings.NewReader(""))
	writer := httptest.NewRecorder()
	router.ServeHTTP(writer, request)

	// assert
	if writer.Code != http.StatusOK {
		t.Errorf("expected status code %v, got %v", http.StatusOK, writer.Code)
	}
	want := todoItems
	got := []core.TodoItem{}
	err := json.Unmarshal(writer.Body.Bytes(), &got)
	if err != nil {
		t.Fatalf("error unmarshalling response body: %v", err)
	}
	// NOTE: Sort the slices before comparing them because the order of the items is not guaranteed.
	sort.Slice(got, func(i, j int) bool { return got[i].Id < got[j].Id })
	if !reflect.DeepEqual(got, want) {
		t.Errorf("expected body %v, got %v", want, got)
	}
}
