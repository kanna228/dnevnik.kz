package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"

	"github.com/sirupsen/logrus"

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
	http.ServeFile(w, r, "main_page.html")
}

// Обработчик страницы логина
func login_page(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "login.html")
}

// Обработчик страницы логина для учителей
func teacher_login_page(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "teacher_login.html")
}

// Обработчик страницы регистрации
func register_page(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "register.html")
}

// Обработчик страницы контактов
func handler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "contact.html")
}

// Обработчик страницы списка
func list(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "list.html")
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

	collection := client.Database("your_db_name").Collection("users")
	result, err := collection.InsertOne(context.Background(), user)
	if err != nil {
		http.Error(w, "Failed to insert user", http.StatusInternalServerError)
		return
	}

	user.ID = result.InsertedID.(primitive.ObjectID)

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

	var updateData User
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	update := bson.M{"$set": bson.M{
		"name":     updateData.Name,
		"email":    updateData.Email,
		"password": updateData.Password,
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

	// Existing Routes
	http.HandleFunc("/", main_page)
	http.HandleFunc("/login", login_page)
	http.HandleFunc("/teacher_login", teacher_login_page)
	http.HandleFunc("/register", register_page)
	http.HandleFunc("/contact", handler)
	http.HandleFunc("/api", postHandler)
	http.HandleFunc("/list", list)

	// CRUD Routes
	http.HandleFunc("/api/users/create", createUser)
	http.HandleFunc("/api/users/all", getAllUsers)
	http.HandleFunc("/api/users/update", updateUser)
	http.HandleFunc("/api/users/delete", deleteUser)
	http.HandleFunc("/api/users/get", getUserByID)

	log := setupLogger()

	log.WithFields(logrus.Fields{
		"action": "start",
		"status": "success",
	}).Info("Application started successfully")

	// Start Server
	fmt.Println("Server running at http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Server Error:", err)
	}
}
