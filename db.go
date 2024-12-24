// db.go
package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Глобальная переменная для подключения к базе данных
var client *mongo.Client

// Функция для подключения к MongoDB
func ConnectToMongoDB() (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	return mongo.Connect(context.Background(), clientOptions)
}

// Функция для отключения от MongoDB
func DisconnectMongoDB() {
	if err := client.Disconnect(context.Background()); err != nil {
		log.Fatal("Failed to disconnect from MongoDB:", err)
	}
}

// Функция для чтения данных из файла users.json
func loadUsersFromJSON(filePath string) ([]User, error) {
	var users []User

	// Открыть файл
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Прочитать содержимое файла
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	// Декодировать JSON данные в структуру
	err = json.Unmarshal(data, &users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

// Функция для добавления пользователей в базу данных, если их нет
func insertUsersIfNotExist(users []User) {
	collection := client.Database("your_db_name").Collection("users")

	for _, user := range users {
		// Проверяем, существует ли уже пользователь с таким email
		var existingUser User
		err := collection.FindOne(context.Background(), bson.M{"email": user.Email}).Decode(&existingUser)

		if err != nil { // Если пользователь не найден, добавляем нового
			if err := insertUser(user); err != nil {
				log.Println("Failed to insert user:", err)
			}
		}
	}
}

// Функция для вставки пользователя в базу данных
func insertUser(user User) error {
	collection := client.Database("your_db_name").Collection("users")
	_, err := collection.InsertOne(context.Background(), user)
	if err != nil {
		return err
	}
	return nil
}
