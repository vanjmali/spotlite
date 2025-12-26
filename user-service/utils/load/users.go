package load

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/vanjmali/spotlite/user-service/entities"
	infraMongo "github.com/vanjmali/spotlite/user-service/infrastructure/mongo"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestLoadSeed(mc *mongo.Client) {
	const (
		totalUsers  = 600
		insertBatch = 500
	)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	c := mc.Database(infraMongo.DatabaseName()).Collection(`users`)

	now := time.Now().UTC()
	expiryDate := now.AddDate(0, 0, 5)
	lastSent := now.AddDate(0, 0, -2)
	passwordHash := "$2a$12$cREglmMvY.5rWeAg1Fy.u.WANxzgD5B4kn6MOQmw07Ao9VAInA2Rm"

	fmt.Printf("Starting insertion of %d users...\n", totalUsers)

	for i := 0; i < totalUsers; i += insertBatch {
		var batch []interface{}

		currentBatchSize := insertBatch
		if i+insertBatch > totalUsers {
			currentBatchSize = totalUsers - i
		}

		for j := 1; j <= currentBatchSize; j++ {
			userID := i + j
			batch = append(batch, entities.User{
				ID:                           primitive.NewObjectID(),
				Username:                     fmt.Sprintf("loadtest_%d", userID),
				FirstName:                    "Test",
				LastName:                     fmt.Sprintf("User_%d", userID),
				Email:                        fmt.Sprintf("test%d@example.com", userID),
				Password:                     passwordHash,
				Role:                         "MEMBER",
				PasswordLastChanged:          now,
				PasswordExpiresAt:            expiryDate,
				AccountStatus:                "ACTIVE",
				CreatedAt:                    now,
				UpdatedAt:                    now,
				LastExpiryNotificationSentAt: lastSent,
			})
		}

		// Execute InsertMany for the batch
		_, err := c.InsertMany(ctx, batch)
		if err != nil {
			log.Fatalf("Failed to insert batch at index %d: %v", i, err)
		}
	}
}
