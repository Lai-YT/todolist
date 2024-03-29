package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
	"todolist/core"
	"todolist/endpoint"
	"todolist/storage"
)

// init is executed when the program first begins (before main).
func init() {
	// Set up our logger settings.
	log.SetFormatter(&log.TextFormatter{})
	log.SetReportCaller(true)
}

func main() {
	accessor := &storage.DatabaseAccessor{}
	accessor.InitDb()
	defer accessor.CloseDb()
	core.SetAccessor(accessor)

	log.Info("Starting Todolist API server")
	router := mux.NewRouter()
	// NOTE: The endpoint are not entirely the same as the blog post.
	router.HandleFunc("/healthz", endpoint.Healthz).Methods("GET")
	router.HandleFunc("/todo", endpoint.CreateItem).Methods("POST")
	router.HandleFunc("/todo", endpoint.GetItems).Methods("GET")
	router.HandleFunc("/todo/{id}", endpoint.UpdateItem).Methods("POST")
	router.HandleFunc("/todo/{id}", endpoint.DeleteItem).Methods("DELETE")

	handler := cors.New(cors.Options{
		// NOTE: "OPTIONS" is not included in comparison with the blog post since it's not necessary.
		// See https://stackoverflow.com/questions/66926518/should-access-control-allow-methods-include-options.
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
	}).Handler(router)
	err := http.ListenAndServe(":8000", handler)
	if err != nil {
		log.Fatal(err)
	}
}
