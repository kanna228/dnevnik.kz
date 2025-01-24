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
	// Read the MongoDB URI from the environment variable
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		log.Fatal("MONGO_URI environment variable is not set")
	}

	// Set MongoDB client options with the URI
	clientOptions := options.Client().ApplyURI(mongoURI)

	// Connect to MongoDB
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, err
	}

	// Ping the database to ensure connection is successful
	err = client.Ping(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	log.Println("Successfully connected to MongoDB")
	return client, nil
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
