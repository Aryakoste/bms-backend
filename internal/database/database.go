package database

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

func Connect(uri string) (*mongo.Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// MongoDB connection options
	clientOptions := options.Client().ApplyURI(uri)
	clientOptions.SetMaxPoolSize(10)

	var err error
	client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Ping the database
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	log.Println("✅ Connected to MongoDB successfully")

	db := client.Database("building_management_society")

	// Create indexes in background
	go createIndexes(db)

	return db, nil
}

func Disconnect(ctx context.Context) error {
	if client != nil {
		log.Println("📴 Disconnecting from MongoDB...")
		return client.Disconnect(ctx)
	}
	return nil
}

func createIndexes(db *mongo.Database) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Println("🔧 Creating database indexes for multi-society system...")

	// Societies collection unique code index
	societiesCollection := db.Collection("societies")
	societiesCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    map[string]interface{}{"code": 1},
		Options: options.Index().SetUnique(true),
	})

	// Users collection compound unique index (email + society_code)
	usersCollection := db.Collection("users")
	usersCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: map[string]interface{}{
			"email":        1,
			"society_code": 1,
		},
		Options: options.Index().SetUnique(true),
	})

	// Visitors collection QR code index
	visitorsCollection := db.Collection("visitors")
	visitorsCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    map[string]interface{}{"qr_code": 1},
		Options: options.Index().SetUnique(true),
	})

	// Society code indexes for all collections
	collections := []string{"users", "visitors", "maintenance", "amenities", "amenity_bookings", "notices"}
	for _, collName := range collections {
		collection := db.Collection(collName)
		collection.Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys: map[string]interface{}{"society_code": 1},
		})
	}

	log.Println("✅ Multi-society database indexes created")
}