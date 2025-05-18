package api

import (
	"net/http"
)

func getRouter() *http.ServeMux {
	router := http.NewServeMux()

	router.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	return router
}
