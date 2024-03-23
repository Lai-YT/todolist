package main

import (
	"encoding/json"
	_ "github.com/go-sql-driver/mysql" // We do not intend to use the variable.
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
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
	router.HandleFunc("/healthz", Healthz).Methods("GET")
	router.HandleFunc("/todo", CreateItem).Methods("POST")
	http.ListenAndServe(":8000", router)
}
