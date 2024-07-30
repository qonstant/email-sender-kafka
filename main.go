package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/joho/godotenv"
)

type User struct {
	ID               int64     `json:"id"`
	FullName         string    `json:"full_name"`
	Email            string    `json:"email"`
	Address          string    `json:"address"`
	RegistrationDate time.Time `json:"registration_date"`
	Role             string    `json:"role"`
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func sendEmail(message string, toAddress string) (response bool, err error) {
	fromAddress := os.Getenv("EMAIL")
	fromEmailPassword := os.Getenv("EMAIL_PASSWORD")
	smtpServer := os.Getenv("SMTP_SERVER")
	smtpPort := os.Getenv("SMTP_PORT")

	auth := smtp.PlainAuth("", fromAddress, fromEmailPassword, smtpServer)
	err = smtp.SendMail(smtpServer+":"+smtpPort, auth, fromAddress, []string{toAddress}, []byte(message))
	if err != nil {
		log.Printf("SMTP error: %v", err)
		return false, err
	}

	return true, nil
}

func Consumer(topic string, groupId string) {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": os.Getenv("UPSTASH_KAFKA_REST_URL"),
		"sasl.mechanism":    "SCRAM-SHA-256",
		"security.protocol": "SASL_SSL",
		"sasl.username":     os.Getenv("UPSTASH_KAFKA_REST_USERNAME"),
		"sasl.password":     os.Getenv("UPSTASH_KAFKA_REST_PASSWORD"),
		"group.id":          groupId,
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}

	err = c.Subscribe(topic, nil)
	if err != nil {
		log.Fatalf("Failed to subscribe to topic: %v", err)
	}

	for {
		ev := c.Poll(100)
		if ev == nil {
			continue
		}

		switch e := ev.(type) {
		case *kafka.Message:
			var user User
			err := json.Unmarshal(e.Value, &user)
			if err != nil {
				log.Printf("Failed to deserialize message: %v", err)
				continue
			}

			subject := "Subject: Account created!\n\n"
			body := fmt.Sprintf("Dear %s,\nYour account is now active and your ID is %d and your role is %s. Congrats!", user.FullName, user.ID, user.Role)
			message := strings.Join([]string{subject, body}, " ")

			_, err = sendEmail(message, user.Email)
			if err != nil {
				log.Printf("Failed to send email: %v", err)
			} else {
				fmt.Printf("Received message key %s value %+v and sent email successfully\n", string(e.Key), user)
			}
		}
	}
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Works fine"))
}

func main() {
	topic := "new-user"
	groupId := "email-new-users"

	// Start Kafka consumer in a goroutine
	go Consumer(topic, groupId)

	// Set up HTTP server for health check
	http.HandleFunc("/health", healthCheckHandler)
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080" // default port if not specified
	}
	log.Printf("Starting HTTP server on port %s", httpPort)
	if err := http.ListenAndServe(":"+httpPort, nil); err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}
