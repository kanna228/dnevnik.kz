package main

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Global MongoDB client variable
var client *mongo.Client

// Connect to MongoDB
func ConnectToMongoDB() (*mongo.Client, error) {
	// MongoDB URI - Adjust this based on your setup
	const uri = "mongodb://localhost:27017"

	// Set client options
	clientOptions := options.Client().ApplyURI(uri)

	// Connect to MongoDB
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	// Ping the database to verify the connection
	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	fmt.Println("Successfully connected to MongoDB!")
	return client, nil
}

// Disconnect from MongoDB
func DisconnectMongoDB() {
	if err := client.Disconnect(context.Background()); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Disconnected from MongoDB!")
}
