package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Chat структура чата
type Chat struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"` // Уникальный ID чата
	StudentID primitive.ObjectID `bson:"student_id" json:"student_id"`      // ID студента, который создал чат
	Title     string             `bson:"title" json:"title"`                // Название чата, заданное пользователем
	Messages  []Message          `bson:"messages" json:"messages"`          // История сообщений
	Status    string             `bson:"status" json:"status"`              // Статус чата (active, closed)
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`      // Время создания чата
}

// Message структура сообщения
type Message struct {
	SenderID  primitive.ObjectID `bson:"sender_id" json:"sender_id"` // ID отправителя (учитель или студент)
	Content   string             `bson:"content" json:"content"`     // Текст сообщения
	Timestamp time.Time          `bson:"timestamp" json:"timestamp"` // Время отправки
}

// CreateChatHandler обрабатывает запрос на создание нового чата
func CreateChatHandler(w http.ResponseWriter, r *http.Request) {
	// Разрешаем CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	// Проверяем, передан ли токен
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		log.Println("❌ Unauthorized: no token")
		http.Error(w, "Unauthorized: no token", http.StatusUnauthorized)
		return
	}

	// Извлекаем токен
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// Декодируем токен
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		log.Println("❌ Unauthorized: invalid token")
		http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
		return
	}

	// Логируем ID пользователя из токена
	log.Println("✅ Extracted User ID from token:", claims.UserID)

	// Конвертируем `student_id` в `ObjectID`
	studentObjectID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		log.Println("❌ Invalid user ID:", claims.UserID)
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Подключаемся к коллекции чатов
	collection := GetChatsCollection()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Проверяем, есть ли уже активный чат у студента
	var existingChat Chat
	err = collection.FindOne(ctx, bson.M{"student_id": studentObjectID, "status": "active"}).Decode(&existingChat)
	if err == nil {
		log.Println("❌ У пользователя уже есть активный чат:", existingChat.ID.Hex())
		http.Error(w, "У вас уже есть активный чат", http.StatusConflict)
		return
	}

	// Декодируем тело запроса
	var requestData struct {
		Title string `json:"title"`
	}
	body, _ := io.ReadAll(r.Body) // Читаем тело запроса
	log.Println("📥 Полученное тело запроса:", string(body))

	if err := json.Unmarshal(body, &requestData); err != nil {
		log.Println("❌ Ошибка разбора JSON:", err)
		http.Error(w, "Неверный формат данных", http.StatusBadRequest)
		return
	}

	log.Println("✅ Создание чата с названием:", requestData.Title)

	// Создаем новый чат
	newChat := Chat{
		ID:        primitive.NewObjectID(),
		StudentID: studentObjectID,
		Title:     requestData.Title,
		Messages:  []Message{},
		Status:    "active",
		CreatedAt: time.Now(),
	}

	_, err = collection.InsertOne(ctx, newChat)
	if err != nil {
		log.Println("❌ Ошибка вставки в MongoDB:", err)
		http.Error(w, "Ошибка создания чата", http.StatusInternalServerError)
		return
	}

	log.Println("✅ Чат успешно создан с ID:", newChat.ID.Hex())

	// Отправляем ответ
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Чат успешно создан",
		"chat_id": newChat.ID.Hex(),
	})
}
func GetChatsHandler(w http.ResponseWriter, r *http.Request) {
	// Разрешаем CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	// Проверяем, передан ли токен
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Unauthorized: no token", http.StatusUnauthorized)
		return
	}

	// Извлекаем токен
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// Декодируем токен
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
		return
	}

	// Конвертируем `UserID` в `ObjectID`
	userID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Подключаемся к коллекции чатов
	collection := GetChatsCollection()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Если пользователь студент, он видит только свой чат, иначе учитель видит все
	var filter bson.M
	if claims.Role == "student" {
		filter = bson.M{"student_id": userID}
	} else {
		filter = bson.M{} // Учитель видит все чаты
	}

	// Выполняем запрос к MongoDB
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		http.Error(w, "Ошибка получения чатов", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var chats []Chat
	if err = cursor.All(ctx, &chats); err != nil {
		http.Error(w, "Ошибка обработки данных", http.StatusInternalServerError)
		return
	}

	// Отправляем JSON с чатами
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(chats)
}
