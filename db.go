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

var client *mongo.Client

func ConnectToMongoDB() (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	return mongo.Connect(context.Background(), clientOptions)
}

func DisconnectMongoDB() {
	if err := client.Disconnect(context.Background()); err != nil {
		log.Fatal("Failed to disconnect from MongoDB:", err)
	}
}

func loadUsersFromJSON(filePath string) ([]User, error) {
	var users []User
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func insertUsersIfNotExist(users []User) {
	collection := client.Database("your_db_name").Collection("users")

	for _, user := range users {
		var existingUser User
		err := collection.FindOne(context.Background(), bson.M{"email": user.Email}).Decode(&existingUser)

		if err != nil {
			if err := insertUser(user); err != nil {
				log.Println("Failed to insert user:", err)
			}
		}
	}
}

func insertUser(user User) error {
	collection := client.Database("your_db_name").Collection("users")
	_, err := collection.InsertOne(context.Background(), user)
	if err != nil {
		return err
	}
	return nil
}
