package endpoint

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"todolist/core"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// Healthz responds with a simple health check message to the client every time it's invoked.
func Healthz(writer http.ResponseWriter, request *http.Request) {
	log.Info("API Health is OK")
	writer.Header().Set("Content-Type", "application/json")
	io.WriteString(writer, `{"alive": true}`)
}

// CreateItem creates a new TodoItem in the database and returns the newly created item to the client to ensure that the operation was successful.
//
// The description of the TodoItem is passed as a form parameter named "description".
//
//	{ "description": "string" }
//
// The response will be the newly created TodoItem.
func CreateItem(writer http.ResponseWriter, request *http.Request) {
	description := request.FormValue("description")
	todo := core.CreateItem(description)
	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(todo)
}

// UpdateItem updates the completed status of a TodoItem in the database.
//
// The completed status is passed as a form parameter named "completed".
//
//	{ "completed": bool }
//
// If the operation was successful:
//
//	{"updated": true}
//
// If the TodoItem was not found in the database:
//
//	{"updated": false, "error": "some error message"}
func UpdateItem(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	id, _ := strconv.Atoi(vars["id"])
	completed, _ := strconv.ParseBool(request.FormValue("completed"))

	_, err := core.UpdateItem(id, completed)

	var response string
	if err != nil {
		response = `{"updated": false, "error": "` + err.Error() + `"}`
	} else {
		response = `{"updated": true}`
	}
	writer.Header().Set("Content-Type", "application/json")
	io.WriteString(writer, response)
}

// DeleteItem deletes a TodoItem from the database.
// If the operation was successful:
//
//	{"deleted": true}
//
// If the TodoItem was not found in the database:
//
//	{"deleted": false, "error": "some error message"}
func DeleteItem(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	id, _ := strconv.Atoi(vars["id"])

	err := core.DeleteItem(id)

	var response string
	if err != nil {
		response = `{"deleted": false, "error": "` + err.Error() + `"}`
	} else {
		response = `{"deleted": true}`
	}
	writer.Header().Set("Content-Type", "application/json")
	io.WriteString(writer, response)
}

// GetItems returns all TodoItems from the database.
// The completed status of the TodoItems can be filtered by passing a query parameter named "completed".
// If the query parameter "completed" is not passed, all TodoItems are returned.
func GetItems(writer http.ResponseWriter, request *http.Request) {
	completed, unspecified := strconv.ParseBool(request.FormValue("completed"))

	var todos []core.TodoItem
	// If the query parameter "completed" is not passed, all TodoItems are returned.
	if unspecified != nil {
		todos = core.GetItems(true)
		todos = append(todos, core.GetItems(false)...)
	} else {
		todos = core.GetItems(completed)
	}

	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(todos)
}
