package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AnalyticsHandler struct {
	userCollection        *mongo.Collection
	visitorCollection     *mongo.Collection
	maintenanceCollection *mongo.Collection
	amenityCollection     *mongo.Collection
	noticeCollection      *mongo.Collection
}

func NewAnalyticsHandler(db *mongo.Database) *AnalyticsHandler {
	return &AnalyticsHandler{
		userCollection:        db.Collection("users"),
		visitorCollection:     db.Collection("visitors"),
		maintenanceCollection: db.Collection("maintenance"),
		amenityCollection:     db.Collection("amenity_bookings"),
		noticeCollection:      db.Collection("notices"),
	}
}

type DashboardStats struct {
	TotalUsers                      int     `json:"total_users"`
	TotalResidents                  int     `json:"total_residents"`
	PendingVisitors                 int     `json:"pending_visitors"`
	ApprovedVisitorsToday           int     `json:"approved_visitors_today"`
	OverdueMaintenance              int     `json:"overdue_maintenance"`
	MaintenanceCollectionPercentage float64 `json:"maintenance_collection_percentage"`
	ActiveAmenityBookings           int     `json:"active_amenity_bookings"`
	UnreadNotices                   int     `json:"unread_notices"`
	SocietyCode                     string  `json:"society_code"`
	LastUpdated                     string  `json:"last_updated"`

	MyPendingAmount *float64 `json:"my_pending_amount,omitempty"`
	MyPaidAmount    *float64 `json:"my_paid_amount,omitempty"`

	CheckedInNow *int `json:"checked_in_now,omitempty"`
}

func (h *AnalyticsHandler) GetStats(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userID := c.GetString("user_id")
	role := c.GetString("user_role")
	societyCode := c.GetString("society_code")
	unit := c.GetString("unit")

	stats := DashboardStats{
		SocietyCode: societyCode,
		LastUpdated: time.Now().Format(time.RFC3339),
	}

	switch role {
	case "secretary":
		stats = h.getSecretaryStats(ctx, societyCode)
	case "resident":
		stats = h.getResidentStats(ctx, societyCode, userID, unit)
	case "security":
		stats = h.getSecurityStats(ctx, societyCode)
	default:
		c.JSON(http.StatusForbidden, gin.H{"error": "Invalid role"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h *AnalyticsHandler) getSecretaryStats(ctx context.Context, societyCode string) DashboardStats {
	stats := DashboardStats{
		SocietyCode: societyCode,
		LastUpdated: time.Now().Format(time.RFC3339),
	}

	totalUsers, _ := h.userCollection.CountDocuments(ctx, bson.M{"society_code": societyCode})
	stats.TotalUsers = int(totalUsers)

	totalResidents, _ := h.userCollection.CountDocuments(ctx, bson.M{
		"society_code": societyCode,
		"role":         "resident",
	})
	stats.TotalResidents = int(totalResidents)

	pendingVisitors, _ := h.visitorCollection.CountDocuments(ctx, bson.M{
		"society_code": societyCode,
		"status":       "pending",
	})
	stats.PendingVisitors = int(pendingVisitors)

	todayStart := time.Now().Truncate(24 * time.Hour)
	approvedToday, _ := h.visitorCollection.CountDocuments(ctx, bson.M{
		"society_code": societyCode,
		"status":       bson.M{"$in": []string{"approved", "checked_in"}},
		"created_at":   bson.M{"$gte": todayStart},
	})
	stats.ApprovedVisitorsToday = int(approvedToday)

	overdueMaintenance, _ := h.maintenanceCollection.CountDocuments(ctx, bson.M{
		"society_code": societyCode,
		"status":       "pending",
		"due_date":     bson.M{"$lt": time.Now()},
	})
	stats.OverdueMaintenance = int(overdueMaintenance)

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"society_code": societyCode}}},
		{{Key: "$group", Value: bson.M{
			"_id":   "$status",
			"total": bson.M{"$sum": "$amount"},
		}}},
	}
	cursor, _ := h.maintenanceCollection.Aggregate(ctx, pipeline)
	var results []bson.M
	cursor.All(ctx, &results)

	var totalPaid, totalDue float64
	for _, result := range results {
		status := result["_id"].(string)
		amount := result["total"].(float64)
		if status == "paid" {
			totalPaid = amount
		} else if status == "pending" || status == "overdue" {
			totalDue += amount
		}
	}
	if totalPaid+totalDue > 0 {
		stats.MaintenanceCollectionPercentage = (totalPaid / (totalPaid + totalDue)) * 100
	}

	activeBookings, _ := h.amenityCollection.CountDocuments(ctx, bson.M{
		"society_code": societyCode,
		"status":       "confirmed",
	})
	stats.ActiveAmenityBookings = int(activeBookings)

	unreadNotices, _ := h.noticeCollection.CountDocuments(ctx, bson.M{
		"society_code": societyCode,
		"is_active":    true,
	})
	stats.UnreadNotices = int(unreadNotices)

	return stats
}

func (h *AnalyticsHandler) getResidentStats(ctx context.Context, societyCode, userID, unit string) DashboardStats {
	stats := DashboardStats{
		SocietyCode: societyCode,
		LastUpdated: time.Now().Format(time.RFC3339),
	}

	userObjectID, _ := primitive.ObjectIDFromHex(userID)

	myPendingVisitors, _ := h.visitorCollection.CountDocuments(ctx, bson.M{
		"society_code": societyCode,
		"host_id":      userObjectID,
		"status":       "pending",
	})
	stats.PendingVisitors = int(myPendingVisitors)

	todayStart := time.Now().Truncate(24 * time.Hour)
	myVisitorsToday, _ := h.visitorCollection.CountDocuments(ctx, bson.M{
		"society_code": societyCode,
		"host_id":      userObjectID,
		"created_at":   bson.M{"$gte": todayStart},
	})
	stats.ApprovedVisitorsToday = int(myVisitorsToday)

	myOverdue, _ := h.maintenanceCollection.CountDocuments(ctx, bson.M{
		"society_code": societyCode,
		"unit_number":  unit,
		"status":       "pending",
		"due_date":     bson.M{"$lt": time.Now()},
	})
	stats.OverdueMaintenance = int(myOverdue)

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"society_code": societyCode,
			"unit_number":  unit,
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":   "$status",
			"total": bson.M{"$sum": "$amount"},
		}}},
	}
	cursor, _ := h.maintenanceCollection.Aggregate(ctx, pipeline)
	var results []bson.M
	cursor.All(ctx, &results)

	var myPending, myPaid float64
	for _, result := range results {
		status := result["_id"].(string)
		amount := result["total"].(float64)
		if status == "paid" {
			myPaid = amount
		} else if status == "pending" || status == "overdue" {
			myPending += amount
		}
	}
	stats.MyPendingAmount = &myPending
	stats.MyPaidAmount = &myPaid

	myBookings, _ := h.amenityCollection.CountDocuments(ctx, bson.M{
		"society_code": societyCode,
		"user_id":      userObjectID,
		"status":       "confirmed",
	})
	stats.ActiveAmenityBookings = int(myBookings)

	unreadNotices, _ := h.noticeCollection.CountDocuments(ctx, bson.M{
		"society_code": societyCode,
		"is_active":    true,
	})
	stats.UnreadNotices = int(unreadNotices)

	return stats
}

func (h *AnalyticsHandler) getSecurityStats(ctx context.Context, societyCode string) DashboardStats {
	stats := DashboardStats{
		SocietyCode: societyCode,
		LastUpdated: time.Now().Format(time.RFC3339),
	}

	pendingVisitors, _ := h.visitorCollection.CountDocuments(ctx, bson.M{
		"society_code": societyCode,
		"status":       "pending",
	})
	stats.PendingVisitors = int(pendingVisitors)

	todayStart := time.Now().Truncate(24 * time.Hour)
	approvedToday, _ := h.visitorCollection.CountDocuments(ctx, bson.M{
		"society_code": societyCode,
		"status":       bson.M{"$in": []string{"approved", "checked_in"}},
		"created_at":   bson.M{"$gte": todayStart},
	})
	stats.ApprovedVisitorsToday = int(approvedToday)

	checkedInNow, _ := h.visitorCollection.CountDocuments(ctx, bson.M{
		"society_code": societyCode,
		"status":       "checked_in",
	})
	checkedIn := int(checkedInNow)
	stats.CheckedInNow = &checkedIn

	return stats
}
