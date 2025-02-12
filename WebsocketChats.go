package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
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
	// Получаем ID пользователя из токена
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

	// Декодируем тело запроса
	var requestData struct {
		Title string `json:"title"`
	}
	body, _ := io.ReadAll(r.Body) // Читаем тело запроса
	log.Println("Полученное тело запроса:", string(body))

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Неверный формат данных", http.StatusBadRequest)
		return
	}
	log.Println("Заголовок чата:", requestData.Title)

	// Конвертируем строковый `UserID` в `primitive.ObjectID`
	userID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		http.Error(w, "Invalid User ID", http.StatusBadRequest)
		return
	}

	// Подключаемся к коллекции чатов
	collection := GetChatsCollection()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Создаем новый чат
	newChat := Chat{
		ID:        primitive.NewObjectID(),
		StudentID: userID, // Используем корректный ObjectID
		Title:     requestData.Title,
		Messages:  []Message{},
		Status:    "active",
		CreatedAt: time.Now(),
	}

	_, err = collection.InsertOne(ctx, newChat)
	if err != nil {
		http.Error(w, "Ошибка создания чата", http.StatusInternalServerError)
		return
	}

	// Отправляем ответ
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Чат успешно создан",
		"chat_id": newChat.ID.Hex(),
	})
}
