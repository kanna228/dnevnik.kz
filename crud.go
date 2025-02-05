package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"

	"net/http"

	"github.com/sirupsen/logrus"

	"go.mongodb.org/mongo-driver/mongo/options"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

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
func generateToken() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
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
