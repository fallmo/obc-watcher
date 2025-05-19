package utils

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func loadEnvironment() {
	err := godotenv.Load()
	if err != nil && !strings.Contains(err.Error(), "no such file or directory") {
		log.Println("Error loading .env file ", err)
	}
}

func verifyEnvironment() {
	requiredVars := [1]string{"RABBITMQ_URI"}

	for i := 0; i < len(requiredVars); i++ {
		if os.Getenv(requiredVars[i]) == "" {
			log.Fatalf("ERROR: Missing environment variable %v\n", requiredVars[i])
		}
	}
}

func StartupTasks() {
	loadEnvironment()
	verifyEnvironment()
	connectRabbitMq()
	connectKubernetes()
}
