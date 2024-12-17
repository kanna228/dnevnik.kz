package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MongoDB User Model
type User struct {
	ID      primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name    string             `bson:"name" json:"name"`
	Email   string             `bson:"email" json:"email"`
	Message string             `bson:"message" json:"message"`
}

// ResponseData structure
type ResponseData struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func main_page(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "main_page.html")
}

func login_page(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "login.html")
}

func teacher_login_page(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "teacher_login.html")
}

func register_page(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "register.html")
}

func handler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "contact.html")
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

// Create a user
func createUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil || user.Name == "" || user.Email == "" || user.Message == "" {
		http.Error(w, "Invalid input data", http.StatusBadRequest)
		return
	}

	collection := client.Database("your_db_name").Collection("users")
	result, err := collection.InsertOne(context.Background(), user)
	if err != nil {
		http.Error(w, "Failed to insert user", http.StatusInternalServerError)
		return
	}

	user.ID = result.InsertedID.(primitive.ObjectID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// Get all users
func getAllUsers(w http.ResponseWriter, r *http.Request) {
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// Update a user
func updateUser(w http.ResponseWriter, r *http.Request) {
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
		"name":    updateData.Name,
		"email":   updateData.Email,
		"message": updateData.Message,
	}}

	collection := client.Database("your_db_name").Collection("users")
	_, err = collection.UpdateOne(context.Background(), bson.M{"_id": objID}, update)
	if err != nil {
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(ResponseData{Status: "success", Message: "User updated"})
}

// Delete a user
func deleteUser(w http.ResponseWriter, r *http.Request) {
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

	json.NewEncoder(w).Encode(ResponseData{Status: "success", Message: "User deleted"})
}

func main() {
	// Connect to MongoDB
	var err error
	client, err = ConnectToMongoDB()
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer DisconnectMongoDB()

	// Existing Routes
	http.HandleFunc("/", main_page)
	http.HandleFunc("/login", login_page)
	http.HandleFunc("/teacher_login", teacher_login_page)
	http.HandleFunc("/register", register_page)
	http.HandleFunc("/contact", handler)
	http.HandleFunc("/api", postHandler)

	// CRUD Routes
	http.HandleFunc("/api/users/create", createUser)
	http.HandleFunc("/api/users/all", getAllUsers)
	http.HandleFunc("/api/users/update", updateUser)
	http.HandleFunc("/api/users/delete", deleteUser)

	// Start Server
	fmt.Println("Server running at http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Server Error:", err)
	}
}
