package main

import (
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
)

// Healthz responds with a simple health check message to the client every time it's invoked.
func Healthz(writer http.ResponseWriter, request *http.Request) {
	log.Info("API Health is OK")
	writer.Header().Set("Content-Type", "application/json")
	io.WriteString(writer, `{"alive": true}`)
}

// init is executed when the program first begins (before main).
func init() {
	// Set up our logger settings.
	log.SetFormatter(&log.TextFormatter{})
	log.SetReportCaller(true)
}

func main() {
	log.Info("Starting Todolist API server")
	router := mux.NewRouter()
	router.HandleFunc("/healthz", Healthz).Methods("GET")
	http.ListenAndServe(":8000", router)
}
