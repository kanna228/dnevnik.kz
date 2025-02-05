package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"

	"go.mongodb.org/mongo-driver/bson"
	"gopkg.in/gomail.v2"
)

// SupportRequest structure to handle support form submissions
type SupportRequest struct {
	UserEmail string `json:"user_email"`
	Subject   string `json:"subject"`
	Message   string `json:"message"`
}

// sendEmail function to send an email using gomail
func sendEmail(to string, subject string, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", "dildahanz@mail.ru") // Replace with your email
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	// Replace SMTP settings with your provider's information
	d := gomail.NewDialer("smtp.mail.ru", 587, "dildahanz@mail.ru", "NmwPuFt4svU9eiDa0Bu0") // Update these details
	return d.DialAndSend(m)
}

func sendEmailWithAttachment(to, subject, body, attachmentPath string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", "dildahanz@mail.ru") // Укажите ваш email
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)
	m.Attach(attachmentPath) // Прикрепляем файл

	d := gomail.NewDialer("smtp.mail.ru", 587, "dildahanz@mail.ru", "NmwPuFt4svU9eiDa0Bu0")
	return d.DialAndSend(m)
}

// handleSupportRequest function to process support form submissions
func handleSupportRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Установить максимальный размер файла (например, 10 MB)

	// Получение текстовых данных
	email := r.FormValue("email")
	message := r.FormValue("message")

	// Валидация email
	emailRegex := `^[^\s@]+@[^\s@]+\.[^\s@]+$`
	matched, _ := regexp.MatchString(emailRegex, email)
	if !matched {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	// Получение файла
	file, handler, err := r.FormFile("file")
	if err != nil {
		if err == http.ErrMissingFile {
			fmt.Println("Файл не прикреплен, продолжаем без него.")
		} else {
			http.Error(w, "Error reading file", http.StatusBadRequest)
			return
		}
	}
	defer file.Close()

	// Если файл был загружен, сохраните его
	var filePath string
	if handler != nil {
		filePath = fmt.Sprintf("./uploads/%s", handler.Filename)
		outFile, err := os.Create(filePath)
		if err != nil {
			http.Error(w, "Failed to save file", http.StatusInternalServerError)
			return
		}
		defer outFile.Close()
		io.Copy(outFile, file) // Сохраняем файл на сервере
	}

	// Отправка email
	subject := "Support Request"
	body := fmt.Sprintf("Message: %s\n\nEmail: %s", message, email)

	if filePath != "" {
		// Если файл загружен, прикрепляем его к письму
		err = sendEmailWithAttachment(email, subject, body, filePath)
	} else {
		err = sendEmail(email, subject, body)
	}

	if err != nil {
		http.Error(w, "Failed to send support email", http.StatusInternalServerError)
		return
	}

	// Ответ пользователю
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ResponseData{Status: "success", Message: "Ваш запрос отправлен"})
}

func testEmailHandler(w http.ResponseWriter, r *http.Request) {
	err := sendEmail("amigo553@mail.ru", "Test Subject", "This is a test email.")
	if err != nil {
		log.Println("Failed to send test email:", err)
		http.Error(w, "Failed to send test email", http.StatusInternalServerError)
		return
	}
	w.Write([]byte("Test email sent successfully!"))
}

func findUserByEmail(email string) (*User, error) {
	collection := client.Database("your_db_name").Collection("users")
	var user User
	filter := bson.M{"email": email}
	err := collection.FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func sendConfirmationEmail(email, token string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", "dildahanz@mail.ru")
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Confirm your registration")
	m.SetBody("text/plain", "Please confirm your registration by clicking the link: http://localhost:8080/confirm?token="+token)

	d := gomail.NewDialer("smtp.mail.ru", 587, "dildahanz@mail.ru", "NmwPuFt4svU9eiDa0Bu0")

	return d.DialAndSend(m)
}
