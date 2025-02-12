package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Структуры данных для расписания
type Schedule struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	StudentID primitive.ObjectID `bson:"student_id" json:"student_id"`
	Schedule  []DaySchedule      `bson:"schedule" json:"schedule"`
}

type DaySchedule struct {
	Day     string   `bson:"day" json:"day"`
	Lessons []Lesson `bson:"lessons" json:"lessons"`
}

type Lesson struct {
	Time    string `bson:"time" json:"time"`
	Teacher string `bson:"teacher" json:"teacher"`
	Subject string `bson:"subject" json:"subject"`
	Room    string `bson:"room" json:"room"`
}

// Функция для получения расписания ученика
func getScheduleHandler(w http.ResponseWriter, r *http.Request) {
	// Разрешаем CORS (чтобы работало с фронтендом)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	// Проверяем, передан ли токен
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		log.Println("❌ Unauthorized: no token")
		http.Error(w, "Unauthorized: no token", http.StatusUnauthorized)
		return
	}

	// Извлекаем токен (может быть с "Bearer" или просто токен)
	tokenString := authHeader
	if strings.HasPrefix(authHeader, "Bearer ") {
		tokenString = strings.TrimPrefix(authHeader, "Bearer ")
	}

	// Логируем токен перед парсингом
	log.Println("✅ Received Token:", tokenString)

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

	// Подключаемся к коллекции расписаний
	schedulesCollection := GetScheduleCollection()

	// Ищем расписание по `student_id`
	var schedule Schedule
	err = schedulesCollection.FindOne(context.Background(), bson.M{"student_id": studentObjectID}).Decode(&schedule)
	if err != nil {
		log.Println("❌ Schedule not found for user ID:", studentObjectID.Hex()) // Логируем ID, по которому искали
		http.Error(w, "Schedule not found", http.StatusNotFound)
		return
	}

	// Логируем найденное расписание
	log.Println("✅ Schedule found for user ID:", studentObjectID.Hex())

	// Отправляем JSON с расписанием
	json.NewEncoder(w).Encode(schedule)
}

// Регистрация маршрута
func RegisterScheduleRoutes() {
	http.HandleFunc("/api/schedule", getScheduleHandler)
}
