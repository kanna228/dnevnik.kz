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
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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
	SenderID  primitive.ObjectID `bson:"sender_id" json:"sender_id"`
	Content   string             `bson:"content" json:"content"`
	Timestamp time.Time          `bson:"timestamp" json:"timestamp"`
	ChatID    primitive.ObjectID `bson:"chat_id" json:"chat_id"`
	Username  string             `bson:"username" json:"username"` // Добавлено для имени пользователя
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

	// Фильтруем только активные чаты
	var filter bson.M
	if claims.Role == "student" {
		// Студент видит только свои активные чаты
		filter = bson.M{"student_id": userID, "status": "active"}
	} else {
		// Учитель видит все активные чаты
		filter = bson.M{"status": "active"}
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

// Функция для добавления сообщения в чат
func addMessageToChat(chatID primitive.ObjectID, message Message) (*mongo.UpdateResult, error) {
	// Подключаемся к коллекции чатов
	collection := GetChatsCollection()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Обновляем чат: добавляем новое сообщение в массив Messages
	update := bson.M{
		"$push": bson.M{
			"messages": message,
		},
	}

	// Выполняем обновление
	return collection.UpdateOne(ctx, bson.M{"_id": chatID}, update)
}

// Хранение всех соединений для каждого чата
var chatConnections = make(map[primitive.ObjectID][]*websocket.Conn)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Разрешаем подключение с любых источников
		return true
	},
}

// Обработчик соединений WebSocket
func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Ошибка WebSocket:", err)
		return
	}
	defer ws.Close()

	// Извлекаем chat_id из query параметра
	chatId := r.URL.Query().Get("chat_id")
	if chatId == "" {
		log.Println("Ошибка: chat_id не передан")
		return
	}

	// Конвертируем chat_id в ObjectID
	chatObjectID, err := primitive.ObjectIDFromHex(chatId)
	if err != nil {
		log.Println("Ошибка: неверный формат chat_id")
		return
	}

	log.Printf("Подключен к чату %s", chatId)

	// Добавляем текущее соединение в список для этого чата
	chatConnections[chatObjectID] = append(chatConnections[chatObjectID], ws)

	for {
		var message Message
		err := ws.ReadJSON(&message)
		if err != nil {
			log.Println("Ошибка при получении сообщения:", err)
			break
		}

		// Заполняем поля сообщения
		message.Timestamp = time.Now()
		message.ChatID = chatObjectID

		// Сохраняем сообщение в базе данных
		updateResult, err := addMessageToChat(chatObjectID, message)
		if err != nil {
			log.Println("Ошибка при добавлении сообщения в чат:", err)
			continue
		}

		// Отправляем сообщение всем подключенным клиентам в этом чате
		for _, conn := range chatConnections[chatObjectID] {
			err := conn.WriteJSON(message)
			if err != nil {
				log.Println("Ошибка при отправке сообщения:", err)
				break
			}
		}

		log.Printf("Сообщение успешно добавлено в чат. Обновлено %v чатов.", updateResult.ModifiedCount)
	}
}

func getChatHistoryHandler(w http.ResponseWriter, r *http.Request) {
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

	// Получаем chat_id из параметров запроса
	chatId := r.URL.Query().Get("chat_id")
	if chatId == "" {
		http.Error(w, "chat_id is required", http.StatusBadRequest)
		return
	}

	chatObjectID, err := primitive.ObjectIDFromHex(chatId)
	if err != nil {
		http.Error(w, "Invalid chat_id format", http.StatusBadRequest)
		return
	}

	// Подключаемся к коллекции чатов
	collection := GetChatsCollection()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var chat Chat
	err = collection.FindOne(ctx, bson.M{"_id": chatObjectID}).Decode(&chat)
	if err != nil {
		http.Error(w, "Chat not found", http.StatusNotFound)
		return
	}

	// Отправляем историю сообщений чата
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chat.Messages)
}

// CloseChatHandler переводит чат в состояние "closed"
func CloseChatHandler(w http.ResponseWriter, r *http.Request) {
	// Разрешаем CORS и устанавливаем тип ответа
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	// Проверяем наличие токена
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		log.Println("❌ Unauthorized: no token")
		http.Error(w, "Unauthorized: no token", http.StatusUnauthorized)
		return
	}
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

	// Получаем chat_id из query-параметра
	chatId := r.URL.Query().Get("chat_id")
	if chatId == "" {
		http.Error(w, "chat_id is required", http.StatusBadRequest)
		return
	}
	chatObjectID, err := primitive.ObjectIDFromHex(chatId)
	if err != nil {
		http.Error(w, "Invalid chat_id format", http.StatusBadRequest)
		return
	}

	// Подключаемся к коллекции чатов
	collection := GetChatsCollection()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Опционально: можно добавить проверку, что чат принадлежит текущему пользователю

	// Обновляем статус чата на "closed"
	update := bson.M{
		"$set": bson.M{
			"status": "closed",
		},
	}

	updateResult, err := collection.UpdateOne(ctx, bson.M{"_id": chatObjectID}, update)
	if err != nil || updateResult.ModifiedCount == 0 {
		log.Println("❌ Не удалось закрыть чат:", err)
		http.Error(w, "Не удалось закрыть чат", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Чат %s успешно закрыт", chatId)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Чат успешно закрыт",
		"chat_id": chatId,
	})
}
