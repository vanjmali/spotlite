package load

import (
	"context"
	"fmt"
	"time"

	"github.com/vanjmali/spotlite/common-lib/logging"
	"github.com/vanjmali/spotlite/user-service/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestLoadSeed(mc *mongo.Client, dbName string) {
	const (
		totalUsers  = 600
		insertBatch = 500
	)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	c := mc.Database(dbName).Collection(`users`)

	now := time.Now().UTC()
	expiryDate := now.AddDate(0, 0, 5)
	lastSent := now.AddDate(0, 0, -2)
	passwordHash := ""

	logging.Infof(ctx, "starting insertion of %d users", totalUsers)

	for i := 0; i < totalUsers; i += insertBatch {
		var batch []any

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
			logging.Errorf(ctx, "failed to insert batch at index %d: %v", i, err)
		}
	}
}
