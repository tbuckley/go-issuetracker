package gae

import (
	"net/http"

	"github.com/gorilla/mux"
)

func init() {
	r := mux.NewRouter()

	r.HandleFunc("/api/issues/{label}", HandleGetIssues).Methods("GET")

	r.HandleFunc("/tasks/issues/reset", HandleResetIssues).Methods("GET")
	r.HandleFunc("/tasks/issues/update", HandleUpdateIssues).Methods("GET")

	r.PathPrefix("/components/").Handler(http.StripPrefix("/components/", http.FileServer(http.Dir("static/components"))))
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("static/dashboard")))

	http.Handle("/", r)
}
