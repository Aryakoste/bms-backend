package main

import (
	"context"
	"log"
	"time"

	"bms-backend/internal/config"
	"bms-backend/internal/database"
	"bms-backend/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	cfg := config.Load()
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Disconnect(context.Background())

	log.Println("🌱 Seeding database with MULTI-SOCIETY sample data...")

	// Clear existing data
	log.Println("🧹 Clearing existing data...")
	collections := []string{"societies", "users", "visitors", "maintenance", "amenities", "amenity_bookings", "notices"}
	for _, collName := range collections {
		db.Collection(collName).Drop(context.Background())
	}

	// Seed societies first
	societies := seedSocieties(db)

	// Seed data for each society
	for _, society := range societies {
		log.Printf("🏢 Creating data for society: %s (%s)", society.Name, society.Code)
		users := seedUsersForSociety(db, society)
		seedAmenitiesForSociety(db, society)
		seedMaintenanceForSociety(db, society, users)
		seedNoticesForSociety(db, society, users)
		seedVisitorsForSociety(db, society, users)
	}

	log.Println("✅ Multi-society database seeding completed!")
	log.Println("🔐 Demo credentials:")
	log.Println("Society: GREEN001 (Green Valley Apartments)")
	log.Println("   Secretary: rajesh@demo.com / demo123")
	log.Println("   Resident: priya@demo.com / demo123")
	log.Println("   Security: security@demo.com / demo123")
	log.Println("")
	log.Println("Society: BLUE002 (Blue Hills Society)")
	log.Println("   Secretary: admin@bluehills.com / demo123")
	log.Println("   Resident: resident@bluehills.com / demo123")
	log.Println("   Security: guard@bluehills.com / demo123")
}

func seedSocieties(db *mongo.Database) []models.Society {
	collection := db.Collection("societies")

	societies := []models.Society{
		{
			ID:           primitive.NewObjectID(),
			Name:         "Green Valley Apartments",
			Code:         "GREEN001",
			Address:      "123 Green Valley Road",
			City:         "Mumbai",
			State:        "Maharashtra",
			PinCode:      "400001",
			ContactEmail: "admin@greenvalley.com",
			ContactPhone: "+91 22 1234 5678",
			Buildings: []models.Building{
				{
					ID:            primitive.NewObjectID(),
					Name:          "Tower A",
					Floors:        20,
					UnitsPerFloor: 4,
				},
				{
					ID:            primitive.NewObjectID(),
					Name:          "Tower B",
					Floors:        15,
					UnitsPerFloor: 6,
				},
			},
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:           primitive.NewObjectID(),
			Name:         "Blue Hills Society",
			Code:         "BLUE002",
			Address:      "456 Blue Hills Avenue",
			City:         "Pune",
			State:        "Maharashtra",
			PinCode:      "411001",
			ContactEmail: "contact@bluehills.com",
			ContactPhone: "+91 20 9876 5432",
			Buildings: []models.Building{
				{
					ID:            primitive.NewObjectID(),
					Name:          "Block A",
					Floors:        10,
					UnitsPerFloor: 8,
				},
				{
					ID:            primitive.NewObjectID(),
					Name:          "Block B",
					Floors:        12,
					UnitsPerFloor: 6,
				},
			},
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:           primitive.NewObjectID(),
			Name:         "Sunrise Residency",
			Code:         "SUN003",
			Address:      "789 Sunrise Street",
			City:         "Bangalore",
			State:        "Karnataka",
			PinCode:      "560001",
			ContactEmail: "info@sunriseresidency.com",
			ContactPhone: "+91 80 5555 6666",
			Buildings: []models.Building{
				{
					ID:            primitive.NewObjectID(),
					Name:          "Wing A",
					Floors:        8,
					UnitsPerFloor: 10,
				},
			},
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	var createdSocieties []models.Society
	for _, society := range societies {
		_, err := collection.InsertOne(context.Background(), society)
		if err != nil {
			log.Printf("Error inserting society %s: %v", society.Name, err)
		} else {
			log.Printf("✓ Created society: %s (Code: %s)", society.Name, society.Code)
			createdSocieties = append(createdSocieties, society)
		}
	}

	return createdSocieties
}

func seedUsersForSociety(db *mongo.Database, society models.Society) map[string]models.User {
	collection := db.Collection("users")

	var users []models.User
	if society.Code == "GREEN001" {
		users = []models.User{
			{
				ID:          primitive.NewObjectID(),
				Name:        "Rajesh Kumar",
				Email:       "rajesh@demo.com",
				Password:    hashPassword("demo123"),
				Role:        "secretary",
				Unit:        "A-501",
				Building:    "Tower A",
				Phone:       "+91 9876543210",
				SocietyID:   society.ID,
				SocietyCode: society.Code,
				IsActive:    true,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			{
				ID:          primitive.NewObjectID(),
				Name:        "Priya Sharma",
				Email:       "priya@demo.com",
				Password:    hashPassword("demo123"),
				Role:        "resident",
				Unit:        "B-302",
				Building:    "Tower B",
				Phone:       "+91 9876543211",
				SocietyID:   society.ID,
				SocietyCode: society.Code,
				IsActive:    true,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			{
				ID:          primitive.NewObjectID(),
				Name:        "Security Guard",
				Email:       "security@demo.com",
				Password:    hashPassword("demo123"),
				Role:        "security",
				Unit:        "Security Cabin",
				Building:    "Main Gate",
				Phone:       "+91 9876543212",
				SocietyID:   society.ID,
				SocietyCode: society.Code,
				IsActive:    true,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		}
	} else if society.Code == "BLUE002" {
		users = []models.User{
			{
				ID:          primitive.NewObjectID(),
				Name:        "Admin Blue Hills",
				Email:       "admin@bluehills.com",
				Password:    hashPassword("demo123"),
				Role:        "secretary",
				Unit:        "A-101",
				Building:    "Block A",
				Phone:       "+91 9876543220",
				SocietyID:   society.ID,
				SocietyCode: society.Code,
				IsActive:    true,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			{
				ID:          primitive.NewObjectID(),
				Name:        "Resident Blue Hills",
				Email:       "resident@bluehills.com",
				Password:    hashPassword("demo123"),
				Role:        "resident",
				Unit:        "B-205",
				Building:    "Block B",
				Phone:       "+91 9876543221",
				SocietyID:   society.ID,
				SocietyCode: society.Code,
				IsActive:    true,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			{
				ID:          primitive.NewObjectID(),
				Name:        "Guard Blue Hills",
				Email:       "guard@bluehills.com",
				Password:    hashPassword("demo123"),
				Role:        "security",
				Unit:        "Gate Office",
				Building:    "Main Gate",
				Phone:       "+91 9876543222",
				SocietyID:   society.ID,
				SocietyCode: society.Code,
				IsActive:    true,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		}
	} else {
		// Default users for other societies
		users = []models.User{
			{
				ID:          primitive.NewObjectID(),
				Name:        "Society Admin",
				Email:       "admin@" + society.Code + ".com",
				Password:    hashPassword("demo123"),
				Role:        "secretary",
				Unit:        "A-101",
				Building:    "Block A",
				Phone:       "+91 9876543230",
				SocietyID:   society.ID,
				SocietyCode: society.Code,
				IsActive:    true,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		}
	}

	userMap := make(map[string]models.User)
	for _, user := range users {
		_, err := collection.InsertOne(context.Background(), user)
		if err != nil {
			log.Printf("Error inserting user %s: %v", user.Email, err)
		} else {
			log.Printf("✓ Created user: %s (%s) for %s", user.Name, user.Role, society.Code)
			userMap[user.Role] = user
		}
	}

	return userMap
}

func seedAmenitiesForSociety(db *mongo.Database, society models.Society) {
	collection := db.Collection("amenities")

	amenities := []models.Amenity{
		{
			ID:             primitive.NewObjectID(),
			Name:           "Community Hall",
			Description:    "Large hall for events and gatherings",
			BookingFee:     1000.0,
			Capacity:       100,
			Facilities:     []string{"AC", "Sound System", "Projector"},
			AvailableHours: "6:00 AM - 11:00 PM",
			SocietyID:      society.ID,
			SocietyCode:    society.Code,
			IsActive:       true,
		},
		{
			ID:             primitive.NewObjectID(),
			Name:           "Swimming Pool",
			Description:    "Olympic size swimming pool",
			BookingFee:     500.0,
			Capacity:       20,
			Facilities:     []string{"Changing Rooms", "Towels"},
			AvailableHours: "5:00 AM - 10:00 PM",
			SocietyID:      society.ID,
			SocietyCode:    society.Code,
			IsActive:       true,
		},
	}

	for _, amenity := range amenities {
		_, err := collection.InsertOne(context.Background(), amenity)
		if err != nil {
			log.Printf("Error inserting amenity %s: %v", amenity.Name, err)
		} else {
			log.Printf("✓ Created amenity: %s for %s", amenity.Name, society.Code)
		}
	}
}

func seedMaintenanceForSociety(db *mongo.Database, society models.Society, users map[string]models.User) {
	collection := db.Collection("maintenance")

	resident, exists := users["resident"]
	if !exists {
		log.Printf("No resident found for maintenance records in %s", society.Code)
		return
	}

	records := []models.MaintenanceRecord{
		{
			ID:          primitive.NewObjectID(),
			UnitID:      resident.ID,
			UnitNumber:  resident.Unit,
			Amount:      2500.0,
			Month:       "October 2025",
			DueDate:     time.Date(2025, 10, 31, 0, 0, 0, 0, time.UTC),
			Status:      "pending",
			Description: "Monthly maintenance charges",
			SocietyID:   society.ID,
			SocietyCode: society.Code,
			CreatedAt:   time.Now(),
		},
	}

	for _, record := range records {
		_, err := collection.InsertOne(context.Background(), record)
		if err != nil {
			log.Printf("Error inserting maintenance record: %v", err)
		} else {
			log.Printf("✓ Created maintenance record for %s in %s", record.UnitNumber, society.Code)
		}
	}
}

func seedNoticesForSociety(db *mongo.Database, society models.Society, users map[string]models.User) {
	collection := db.Collection("notices")

	secretary, exists := users["secretary"]
	if !exists {
		log.Printf("No secretary found for notices in %s", society.Code)
		return
	}

	notices := []models.Notice{
		{
			ID:          primitive.NewObjectID(),
			Title:       "Society Meeting",
			Content:     "Monthly society meeting scheduled for next Sunday at 10 AM in the community hall.",
			Type:        "announcement",
			AuthorID:    secretary.ID,
			AuthorName:  secretary.Name,
			SocietyID:   society.ID,
			SocietyCode: society.Code,
			IsActive:    true,
			CreatedAt:   time.Now(),
		},
	}

	for _, notice := range notices {
		_, err := collection.InsertOne(context.Background(), notice)
		if err != nil {
			log.Printf("Error inserting notice: %v", err)
		} else {
			log.Printf("✓ Created notice: %s for %s", notice.Title, society.Code)
		}
	}
}

func seedVisitorsForSociety(db *mongo.Database, society models.Society, users map[string]models.User) {
	collection := db.Collection("visitors")

	resident, exists := users["resident"]
	if !exists {
		log.Printf("No resident found for visitors in %s", society.Code)
		return
	}

	visitors := []models.Visitor{
		{
			ID:           primitive.NewObjectID(),
			Name:         "John Doe",
			Phone:        "+91 9876543299",
			Purpose:      "Family Visit",
			HostID:       resident.ID,
			HostName:     resident.Name,
			HostUnit:     resident.Unit,
			ExpectedTime: time.Now().Add(2 * time.Hour),
			QRCode:       "BMS-" + society.Code + "-visitor-001",
			Status:       "pending",
			SocietyID:    society.ID,
			SocietyCode:  society.Code,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}

	for _, visitor := range visitors {
		_, err := collection.InsertOne(context.Background(), visitor)
		if err != nil {
			log.Printf("Error inserting visitor: %v", err)
		} else {
			log.Printf("✓ Created visitor: %s for %s", visitor.Name, society.Code)
		}
	}
}

func hashPassword(password string) string {
	bytes, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes)
}
