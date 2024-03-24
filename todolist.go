package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql" // We do not intend to use the variable.
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
)

var db, _ = gorm.Open("mysql", "root:root@/todolist?charset=utf8&parseTime=True&loc=Local")

type TodoItemModel struct {
	Id          int `gorm:"primary_key"`
	Description string
	Completed   bool // Whether the todo item is done or not.
}

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
func CreateItem(writer http.ResponseWriter, request *http.Request) {
	description := request.FormValue("description")
	log.WithFields(log.Fields{"description": description}).Info("Add new TodoItem. Saving to database.")
	todo := &TodoItemModel{Description: description, Completed: false}
	db.Create(&todo)
	result := db.Last(&todo)
	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(result.Value)
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
//	{"updated": false, "error": "TodoItem not found"}
func UpdateItem(writer http.ResponseWriter, request *http.Request) {
	// Get the ID from the URL.
	vars := mux.Vars(request)
	id, _ := strconv.Atoi(vars["id"])

	todo := &TodoItemModel{}
	result := db.First(&todo, id)
	writer.Header().Set("Content-Type", "application/json")
	if result.Error != nil {
		log.Warn("TodoItem not found in database")
		io.WriteString(writer, `{"updated": false, "error": "TodoItem not found"}`)
	} else {
		completed, _ := strconv.ParseBool(request.FormValue("completed"))
		log.WithFields(log.Fields{"Id": id, "Completed": completed}).Info("Updating TodoItem")
		todo.Completed = completed
		db.Save(&todo)
		io.WriteString(writer, `{"updated": true}`)
	}
}

// DeleteItem deletes a TodoItem from the database.
// If the operation was successful:
//
//	{"deleted": true}
//
// If the TodoItem was not found in the database:
//
//	{"deleted": false, "error": "TodoItem not found"}
func DeleteItem(writer http.ResponseWriter, request *http.Request) {
	// Get the ID from the URL.
	vars := mux.Vars(request)
	id, _ := strconv.Atoi(vars["id"])

	todo := &TodoItemModel{}
	result := db.First(&todo, id)
	writer.Header().Set("Content-Type", "application/json")
	if result.Error != nil {
		log.Warn("TodoItem not found in database")
		io.WriteString(writer, `{"deleted": false, "error": "TodoItem not found"}`)
	} else {
		log.WithFields(log.Fields{"Id": id}).Info("Deleting TodoItem")
		db.Delete(&todo)
		io.WriteString(writer, `{"deleted": true}`)
	}
}

// GetItems returns all TodoItems from the database.
// The completed status of the TodoItems can be filtered by passing a query parameter named "completed".
// If the query parameter "completed" is not passed, all TodoItems are returned.
func GetItems(writer http.ResponseWriter, request *http.Request) {
	completed, err := strconv.ParseBool(request.URL.Query().Get("completed"))
	var completedTodoItems []TodoItemModel
	if err == nil {
		log.Info("Get TodoItems, completed=", completed)
		db.Where("completed = ?", completed).Find(&completedTodoItems)
	} else {
		log.Info("Get TodoItems")
		db.Find(&completedTodoItems)
	}
	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(completedTodoItems)
}

// init is executed when the program first begins (before main).
func init() {
	// Set up our logger settings.
	log.SetFormatter(&log.TextFormatter{})
	log.SetReportCaller(true)
}

func main() {
	// Close the database connection when the main function ends.
	defer db.Close()
	db.Debug().DropTableIfExists(&TodoItemModel{})
	db.Debug().AutoMigrate(&TodoItemModel{})

	log.Info("Starting Todolist API server")
	router := mux.NewRouter()
	// NOTE: The endpoint are not entirely the same as the blog post.
	router.HandleFunc("/healthz", Healthz).Methods("GET")
	router.HandleFunc("/todo", CreateItem).Methods("POST")
	router.HandleFunc("/todo", GetItems).Methods("GET")
	router.HandleFunc("/todo/{id}", UpdateItem).Methods("POST")
	router.HandleFunc("/todo/{id}", DeleteItem).Methods("DELETE")

	handler := cors.New(cors.Options{
		// NOTE: "OPTIONS" is not included in comparison with the blog post since it's not necessary.
		// See https://stackoverflow.com/questions/66926518/should-access-control-allow-methods-include-options.
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
	}).Handler(router)
	http.ListenAndServe(":8000", handler)
}
