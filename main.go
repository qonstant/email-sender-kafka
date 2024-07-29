package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/smtp"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
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

	var auth = smtp.PlainAuth("", fromAddress, fromEmailPassword, smtpServer)
	err = smtp.SendMail(smtpServer+":"+smtpPort, auth, fromAddress, []string{toAddress}, []byte(message))
	if err == nil {
		return true, nil
	}

	return false, err
}

func getKafkaBrokerFromEnv() (string, error) {
	urlStr := os.Getenv("UPSTASH_KAFKA_REST_URL")
	if urlStr == "" {
		return "", fmt.Errorf("UPSTASH_KAFKA_REST_URL is not set")
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse Kafka URL: %v", err)
	}

	hostname := parsedURL.Hostname()
	port := parsedURL.Port()
	if port == "" {
		port = "9092"
	}

	return fmt.Sprintf("%s:%s", hostname, port), nil
}

func consume(ctx context.Context) {
	mechanism, _ := scram.Mechanism(scram.SHA256, os.Getenv("UPSTASH_KAFKA_REST_USERNAME"), os.Getenv("UPSTASH_KAFKA_REST_PASSWORD"))

	broker, err := getKafkaBrokerFromEnv()
	if err != nil {
		log.Fatalf("failed to get Kafka broker: %v", err)
	}

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{broker}, // Use the extracted broker address
		Topic:   "new-user",
		GroupID: "email-new-users",
		Dialer: &kafka.Dialer{
			SASLMechanism: mechanism,
			TLS:           &tls.Config{},
		},
	})

	defer r.Close()

	for {
		msg, err := r.ReadMessage(ctx)
		if err != nil {
			log.Printf("could not read message: %v", err)
			continue
		}
		userData := msg.Value

		var user User

		err = json.Unmarshal(userData, &user)
		if err != nil {
			log.Printf("could not parse user data: %v", err)
			continue
		}

		subject := "Subject: Account created!\n\n"
		body := fmt.Sprintf("Dear %s,\nYour account is now active and your ID is %s and your role is %s. Congrats!", user.FullName, strconv.Itoa(int(user.ID)), user.Role)
		message := strings.Join([]string{subject, body}, " ")

		_, err = sendEmail(message, user.Email)
		if err != nil {
			log.Printf("failed to send email: %v", err)
		}
	}
}

func main() {
	ctx := context.Background()
	consume(ctx)
}
