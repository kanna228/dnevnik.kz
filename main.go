package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MongoDB User Model
type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name     string             `bson:"name" json:"name"`
	Email    string             `bson:"email" json:"email"`
	Password string             `bson:"password" json:"password"`
	Role     string             `bson:"role" json:"role"` // New field for role
}

// ResponseData structure
type ResponseData struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

var jwtKey = []byte("mY5uP8x/A1bC2dE3fG4hI5jK6lM7nO8pQ9rS0tU1vW2X3yZ4")

type Claims struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
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
func loginHandler(w http.ResponseWriter, r *http.Request) {
	var inputUser User
	if err := json.NewDecoder(r.Body).Decode(&inputUser); err != nil {
		log.Println("Invalid input:", err)
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	log.Printf("Login attempt for email: %s", inputUser.Email)

	// Ищем пользователя в базе данных по email
	dbUser, err := findUserByEmail(inputUser.Email)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("User not found: %s", inputUser.Email)
			http.Error(w, "User not found", http.StatusUnauthorized)
		} else {
			log.Printf("Database error: %v", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	log.Printf("User found: %+v", dbUser)

	// Проверяем пароль (без хеширования)
	if inputUser.Password != dbUser.Password {
		log.Println("Invalid password for user:", dbUser.Name)
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	log.Println("User authenticated:", dbUser.Name)

	// Создание JWT
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: dbUser.Name,
		Email:    dbUser.Email,
		Role:     dbUser.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		log.Println("Failed to generate token:", err)
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Отправка токена клиенту
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

func protectedHandler(w http.ResponseWriter, r *http.Request) {
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.Write([]byte(fmt.Sprintf("Welcome, %s!", claims.Username)))
}

// Настройка логгера
func setupLogger() *logrus.Logger {
	log := logrus.New()

	// Устанавливаем формат JSON для логов
	log.SetFormatter(&logrus.JSONFormatter{})

	// Создаем файл для логов
	file, err := os.OpenFile("logs.json", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}

	// Устанавливаем вывод логов в файл
	log.SetOutput(file)

	return log
}

// Обработчик главной страницы
func main_page(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "views/main_page.html")
}

// Обработчик страницы логина
func login_page(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "views/login.html")
}

// Обработчик страницы логина для учителей
func teacher_login_page(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "views/teacher_login.html")
}

// Обработчик страницы регистрации
func register_page(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "views/register.html")
}

// Обработчик страницы контактов
func handler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "views/contact.html")
}

// Обработчик страницы списка
func list(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "views/list.html")
}
func dashboard(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "views/dashboard.html")
}

// Original POST Handler
func postHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var requestData struct {
		Message string `json:"message"`
		Email   string `json:"email"`
	}

	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil || requestData.Message == "" || requestData.Email == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ResponseData{Status: "fail", Message: "Invalid JSON: 'message' and 'email' are required"})
		return
	}

	emailRegex := `^[^\s@]+@[^\s@]+\.[^\s@]+$`
	matched, _ := regexp.MatchString(emailRegex, requestData.Email)
	if !matched {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ResponseData{Status: "fail", Message: "Invalid email format"})
		return
	}

	collection := client.Database("your_db_name").Collection("users")
	_, err = collection.InsertOne(r.Context(), requestData)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ResponseData{Status: "fail", Message: "Failed to insert data into MongoDB"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ResponseData{Status: "success", Message: "Data successfully received"})
}

// CRUD Functions (Create, Read, Update, Delete)

// Пример создания пользователя с логированием
func generateToken() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
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

func confirmUser(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Token is required", http.StatusBadRequest)
		return
	}

	user, exists := unconfirmedUsers[token]
	if !exists {
		http.Error(w, "Invalid or expired token", http.StatusBadRequest)
		return
	}

	collection := client.Database("your_db_name").Collection("users")
	result, err := collection.InsertOne(context.Background(), user)
	if err != nil {
		http.Error(w, "Failed to insert user", http.StatusInternalServerError)
		return
	}

	user.ID = result.InsertedID.(primitive.ObjectID)

	// Удаляем пользователя из временного хранилища
	delete(unconfirmedUsers, token)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "User confirmed and created successfully"})
}

var unconfirmedUsers = make(map[string]User) // Временное хранилище для неподтвержденных пользователей

func createUser(w http.ResponseWriter, r *http.Request) {
	log := setupLogger()

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil || user.Name == "" || user.Email == "" || user.Password == "" || user.Role == "" {
		http.Error(w, "Invalid input data", http.StatusBadRequest)
		return
	}

	if user.Role != "teacher" && user.Role != "student" {
		http.Error(w, "Invalid role, must be 'teacher' or 'student'", http.StatusBadRequest)
		return
	}

	token, err := generateToken()
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Сохраняем пользователя во временное хранилище
	unconfirmedUsers[token] = user

	// Отправляем email с токеном
	err = sendConfirmationEmail(user.Email, token)
	if err != nil {
		http.Error(w, "Failed to send confirmation email", http.StatusInternalServerError)
		return
	}

	// Логирование
	log.WithFields(logrus.Fields{
		"action": "create_user",
		"method": r.Method,
		"user":   user.Name,
	}).Info("User created successfully")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// Пример получения всех пользователей с логированием
func getAllUsers(w http.ResponseWriter, r *http.Request) {
	log := setupLogger()

	collection := client.Database("your_db_name").Collection("users")
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		http.Error(w, "Failed to retrieve users", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	var users []User
	for cursor.Next(context.Background()) {
		var user User
		if err = cursor.Decode(&user); err != nil {
			continue
		}
		users = append(users, user)
	}
	// Проверяем, если пользователей нет
	if len(users) == 0 {
		http.Error(w, "No users found in the database", http.StatusNotFound)
		return
	}

	// Логирование
	log.WithFields(logrus.Fields{
		"action": "get_all_users",
		"method": r.Method,
		"count":  len(users),
	}).Info("Fetched all users")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// Пример обновления пользователя с логированием
func updateUser(w http.ResponseWriter, r *http.Request) {
	log := setupLogger()

	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idParam := r.URL.Query().Get("id")
	objID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var updateData struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"` // Добавляем поле role
	}

	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Проверяем, что роль либо "student", либо "teacher"
	if updateData.Role != "student" && updateData.Role != "teacher" {
		http.Error(w, "Invalid role, must be 'student' or 'teacher'", http.StatusBadRequest)
		return
	}

	update := bson.M{"$set": bson.M{
		"name":     updateData.Name,
		"email":    updateData.Email,
		"password": updateData.Password,
		"role":     updateData.Role, // Обновляем роль
	}}

	collection := client.Database("your_db_name").Collection("users")
	_, err = collection.UpdateOne(context.Background(), bson.M{"_id": objID}, update)
	if err != nil {
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	// Логирование
	log.WithFields(logrus.Fields{
		"action":  "update_user",
		"method":  r.Method,
		"user_id": objID,
	}).Info("User updated successfully")

	json.NewEncoder(w).Encode(ResponseData{Status: "success", Message: "User updated"})
}

// Пример удаления пользователя с логированием
func deleteUser(w http.ResponseWriter, r *http.Request) {
	log := setupLogger()

	idParam := r.URL.Query().Get("id")
	objID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	collection := client.Database("your_db_name").Collection("users")
	_, err = collection.DeleteOne(context.Background(), bson.M{"_id": objID})
	if err != nil {
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	// Логирование
	log.WithFields(logrus.Fields{
		"action":  "delete_user",
		"method":  r.Method,
		"user_id": objID,
	}).Info("User deleted successfully")

	json.NewEncoder(w).Encode(ResponseData{Status: "success", Message: "User deleted"})
}

// Пример получения пользователя по ID с логированием
func getUserByID(w http.ResponseWriter, r *http.Request) {
	log := setupLogger()

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idParam := r.URL.Query().Get("id")
	if idParam == "" {
		http.Error(w, "ID parameter is required", http.StatusBadRequest)
		return
	}

	objID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	collection := client.Database("your_db_name").Collection("users")
	var user User
	err = collection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&user)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Логирование
	log.WithFields(logrus.Fields{
		"action":  "get_user_by_id",
		"method":  r.Method,
		"user_id": objID,
	}).Info("Fetched user by ID")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)

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

func GetUsersSorted(w http.ResponseWriter, r *http.Request) {
	role := r.URL.Query().Get("role")       // Получаем фильтр роли
	sortOrder := r.URL.Query().Get("order") // Получаем сортировку (asc или desc)

	sort := 1 // По умолчанию сортировка A-Z
	if sortOrder == "desc" {
		sort = -1 // Z-A
	}

	// Фильтр для MongoDB
	filter := bson.M{}
	if role != "" {
		filter["role"] = role
	}

	collection := client.Database("your_db_name").Collection("users")
	cursor, err := collection.Find(
		context.Background(),
		filter,
		options.Find().SetSort(bson.D{{Key: "name", Value: sort}}),
	)
	if err != nil {
		http.Error(w, "Failed to retrieve users", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	var users []User
	for cursor.Next(context.Background()) {
		var user User
		if err := cursor.Decode(&user); err != nil {
			continue
		}
		users = append(users, user)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func main() {
	fs := http.FileServer(http.Dir("./static")) // Указываем папку "static"
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Connect to MongoDB
	var err error
	client, err = ConnectToMongoDB()
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer DisconnectMongoDB()

	// Загрузить пользователей из файла users.json
	users, err := loadUsersFromJSON("users.json")
	if err != nil {
		log.Println("Error loading users from JSON:", err)
	} else {
		insertUsersIfNotExist(users) // Вставить пользователей, если их нет в базе данных
	}

	http.HandleFunc("/api/users/sorted", GetUsersSorted)

	// Existing Routes
	http.HandleFunc("/", main_page)
	http.HandleFunc("/login", login_page)
	http.HandleFunc("/teacher_login", teacher_login_page)
	http.HandleFunc("/register", register_page)
	http.HandleFunc("/contact", handler)
	http.HandleFunc("/api", postHandler)
	http.HandleFunc("/list", list)
	http.HandleFunc("/dashboard", dashboard)

	// CRUD Routes
	http.HandleFunc("/api/users/create", createUser)
	http.HandleFunc("/api/users/all", getAllUsers)
	http.HandleFunc("/api/users/update", updateUser)
	http.HandleFunc("/api/users/delete", deleteUser)
	http.HandleFunc("/api/users/get", getUserByID)

	http.HandleFunc("/confirm", confirmUser)

	http.HandleFunc("/api/login", loginHandler)
	http.HandleFunc("/api/protected", protectedHandler)

	log := setupLogger()

	log.WithFields(logrus.Fields{
		"action": "start",
		"status": "success",
	}).Info("Application started successfully")

	http.HandleFunc("/api/contact", handleSupportRequest)
	// New route for handling support requests
	http.HandleFunc("/support", handleSupportRequest)
	http.HandleFunc("/test-email", testEmailHandler)
	// Start Server
	fmt.Println("Server running at http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Server Error:", err)
	}
}
