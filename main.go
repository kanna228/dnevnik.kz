package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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
	UserID   string `json:"user_id"` // Добавлен user_id
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
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
func help(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "views/help.html")
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var inputUser User
	if err := json.NewDecoder(r.Body).Decode(&inputUser); err != nil {
		log.Println("Invalid input:", err)
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	log.Printf("Login attempt for email: %s", inputUser.Email)

	// Find user by email in MongoDB
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

	// Verify password
	if inputUser.Password != dbUser.Password {
		log.Println("Invalid password for user:", dbUser.Name)
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	log.Println("User authenticated:", dbUser.Name)

	// Convert user ID to string
	userID := dbUser.ID.Hex()

	// Generate JWT token
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID:   userID, // Добавляем user_id в токен
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

	// Send token and user_id to frontend
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token":   tokenString,
		"user_id": dbUser.ID.Hex(),
	})
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

func getUserInfoHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем токен
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

	// Возвращаем информацию о пользователе
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"id":       claims.UserID, // Теперь используем UserID из Claims
		"username": claims.Username,
		"email":    claims.Email,
		"role":     claims.Role,
	})
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

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8080") // Allow frontend
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight requests (OPTIONS method)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Pass the request to the next handler
		next(w, r)
	}
}

func main() {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Connect to MongoDB
	var err error
	client, err = ConnectToMongoDB()
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer DisconnectMongoDB()

	// Load users from JSON
	users, err := loadUsersFromJSON("users.json")
	if err != nil {
		log.Println("Error loading users from JSON:", err)
	} else {
		insertUsersIfNotExist(users)
	}

	// Apply CORS middleware to API routes
	http.Handle("/api/users/sorted", corsMiddleware(http.HandlerFunc(GetUsersSortedWithAggregation)))
	http.Handle("/api/users/create", corsMiddleware(http.HandlerFunc(createUser)))
	http.Handle("/api/users/all", corsMiddleware(http.HandlerFunc(getAllUsers)))
	http.Handle("/api/users/update", corsMiddleware(http.HandlerFunc(updateUser)))
	http.Handle("/api/users/delete", corsMiddleware(http.HandlerFunc(deleteUser)))
	http.Handle("/api/users/get", corsMiddleware(http.HandlerFunc(getUserByID)))
	http.Handle("/api/create_chat", corsMiddleware(http.HandlerFunc(CreateChatHandler)))

	// Other routes
	http.HandleFunc("/", main_page)
	http.HandleFunc("/login", login_page)
	http.HandleFunc("/teacher_login", teacher_login_page)
	http.HandleFunc("/register", register_page)
	http.HandleFunc("/contact", handler)
	http.HandleFunc("/api", postHandler)
	http.HandleFunc("/list", list)
	http.HandleFunc("/help", help)
	http.HandleFunc("/dashboard", dashboard)
	http.HandleFunc("/confirm", confirmUser)
	http.HandleFunc("/api/login", loginHandler)
	http.HandleFunc("/api/protected", protectedHandler)
	http.HandleFunc("/api/user", getUserInfoHandler)
	http.HandleFunc("/api/contact", handleSupportRequest)
	http.HandleFunc("/support", handleSupportRequest)
	http.HandleFunc("/test-email", testEmailHandler)

	RegisterScheduleRoutes()

	log := setupLogger()
	log.WithFields(logrus.Fields{
		"action": "start",
		"status": "success",
	}).Info("Application started successfully")

	go StartPaymentService()
	time.Sleep(1 * time.Second)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server running at http://localhost:%s\n", port)

	if err := http.ListenAndServe("0.0.0.0:"+port, nil); err != nil {
		log.Fatal("Server Error:", err)
	}

}
