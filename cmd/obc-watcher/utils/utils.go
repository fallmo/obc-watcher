package utils

import (
	"fmt"
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
	optionalVars := [1]string{"LABEL_KEY"}

	for i := 0; i < len(requiredVars); i++ {
		if os.Getenv(requiredVars[i]) == "" {
			log.Fatalf("ERROR: Missing environment variable %v\n", requiredVars[i])
		}
	}

	for i := 0; i < len(optionalVars); i++ {
		key := optionalVars[i]
		val := os.Getenv(key)
		if val == "" {
			fmt.Printf("Missing optional environment variable '%v'.\n", key)
		} else {
			fmt.Printf("Running with %v=%v \n", key, val)
		}
	}
}

func StartupTasks() {
	loadEnvironment()
	verifyEnvironment()
	connectRabbitMq()
	connectKubernetes()
}
