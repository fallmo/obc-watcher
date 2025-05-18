package api

import (
	"log"
	"net/http"
	"os"
)

func StartServer() {
	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = "8080"
	}

	server := &http.Server{
		Addr:    ":" + PORT,
		Handler: getRouter(),
	}

	log.Printf("Server listening on port %v", PORT)
	server.ListenAndServe()

}
