package utils

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const queueName = "events"

var rabbitMqChannel **amqp.Channel

func connectRabbitMq() {
	RABBITMQ_URI := os.Getenv("RABBITMQ_URI")

	if RABBITMQ_URI == "" {
		os.Exit(1)
	}

	log.Println("Connecting to RabbitMQ...")
	conn, err := amqp.Dial(RABBITMQ_URI)

	if err != nil {
		log.Fatalln(err, "\nFailed to create amqp connection")
	}

	log.Println("AMQP connection established...")

	channel, err := conn.Channel()

	if err != nil {
		log.Fatalln(err, "\nFailed to get amqp channel")
	}

	rabbitMqChannel = &channel

	_, err = channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)

	if err != nil {
		log.Fatalln(err, "\nFailed declare amqp queue")
	}

	log.Printf("AMQP queue '%v' opened", queueName)

}

type amqpMessage struct {
	Kind string            `json:"kind"`
	Data map[string]string `json:"data"`
}

func PushAMQPMessage(data amqpMessage) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if rabbitMqChannel == nil {
		log.Fatal("No AMQP channel")
	}

	channel := *rabbitMqChannel

	body, err := json.Marshal(data)

	if err != nil {
		log.Panic(err, "\nFailed to parse json")
	}

	err = channel.PublishWithContext(ctx,
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(body),
		})

	if err != nil {
		log.Println(err, "\nError: Failed to Push AMQP Message")
		return
	}

	log.Printf("Successfully Pushed Message to AMQP Queue '%v'", queueName)
}
