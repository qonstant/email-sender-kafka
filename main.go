package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/smtp"
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
	fromEmailPassword := os.Getenv("PASSWORD")
	smtpServer := os.Getenv("SMTP_SERVER")
	smptPort := os.Getenv("SMTP_PORT")

	var auth = smtp.PlainAuth("", fromAddress, fromEmailPassword, smtpServer)
	err = smtp.SendMail(smtpServer+":"+smptPort, auth, fromAddress, []string{toAddress}, []byte(message))
	if err == nil {
		return true, nil
	}

	return false, err
}

func consume(ctx context.Context) {
	mechanism, _ := scram.Mechanism(scram.SHA256, os.Getenv("UPSTASH_KAFKA_REST_USERNAME"), os.Getenv("UPSTASH_KAFKA_REST_PASSWORD"))
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{os.Getenv("UPSTASH_KAFKA_REST_URL")},
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
