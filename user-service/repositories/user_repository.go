package repositories

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/vanjmali/spotlite/common-lib/account"
	"github.com/vanjmali/spotlite/user-service/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ErrTokenExpired signals that the verification token was not found or already used.
var ErrTokenExpired = errors.New("invalid token")

// UserRepository provides data access helpers for user documents.
type UserRepository struct {
	DbName   string
	CollName string
	Client   *mongo.Client
}

// NewRepository constructs a UserRepository for the given database and collection.
func NewRepository(dbName string, collName string, c *mongo.Client) *UserRepository {
	r := UserRepository{Client: c, DbName: dbName, CollName: collName}
	return &r
}

// Create func, inserts a new user into the database.
func (r *UserRepository) Create(ctx context.Context, user entities.User) error {
	c := r.Client.Database(r.DbName).Collection(r.CollName)

	_, err := c.InsertOne(ctx, user)
	if err != nil {
		return err
	}

	return nil
}

// ActiveAndRevokeToken func, that activates the account and revokes the token in one database trip.
func (r *UserRepository) ActiveAndRevokeToken(ctx context.Context, token string) error {
	c := r.Client.Database(r.DbName).Collection(r.CollName)

	// define filtering parameters,
	filter := bson.M{
		"account_status":           account.StatusInactive,
		"email_verification.type":  entities.AccountVerification,
		"email_verification.token": token,
	}

	// define set (set a field value to a new one) and unset (fully remove a field) operations,
	update := bson.M{
		"$set":   bson.M{"account_status": account.StatusActive},
		"$unset": bson.M{"email_verification": ""},
	}

	// calls the update one method which will atomically (all or nothing) update the document
	res, err := c.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	// matched count refers to a number of document that have been found while modified count the
	// number of documents that were modified, the result for both should always be one because
	// there is only one account which is inactive, has the given account verification token and
	// has to be updated
	if res.MatchedCount == 1 && res.ModifiedCount == 1 {
		return nil
	}

	return ErrTokenExpired
}

// SetLoginOtp stores the hashed OTP and expiry for a user.
func (r *UserRepository) SetLoginOtp(
	ctx context.Context,
	userId primitive.ObjectID,
	hash string,
	expiry time.Time,
) error {
	c := r.Client.Database(r.DbName).Collection(r.CollName)

	_, err := c.UpdateOne(ctx,
		bson.M{"_id": userId},
		bson.M{"$set": bson.M{
			"otp_code.hash":   hash,
			"otp_code.expiry": expiry,
			"updated_at":      time.Now(),
		}},
	)
	return err
}

// ClearLoginOtp removes the stored OTP data for a user.
func (r *UserRepository) ClearLoginOtp(ctx context.Context, userId primitive.ObjectID) error {
	c := r.Client.Database(r.DbName).Collection(r.CollName)

	_, err := c.UpdateOne(ctx,
		bson.M{"_id": userId},
		bson.M{"$unset": bson.M{"otp_code": ""}},
	)
	return err
}

// FindUserByEmail fetches a user document by email.
func (r *UserRepository) FindUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	var user entities.User
	c := r.Client.Database(r.DbName).Collection(r.CollName)

	filter := bson.M{"email": email}

	err := c.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindUsersForExpiryNotification finds users which password expiry date is less than (today + daysBeforeExpiry), but
// also greater than today. The user also has to have an active account.
func (r *UserRepository) FindUsersForExpiryNotification(
	ctx context.Context,
	daysUntilExpiry int,
	batchSize int,
	lastID string,
) ([]*entities.User, string, error) {
	log.Printf(
		"DEBUG: FindUsersForExpiryNotification repository function has been called with the following params, %d, %d, %s",
		daysUntilExpiry,
		batchSize,
		lastID,
	)
	c := r.Client.Database(r.DbName).Collection(r.CollName)

	now := time.Now()
	expiryThreshold := time.Now().AddDate(0, 0, daysUntilExpiry)
	notificationWindow := time.Now().Add(-23 * time.Hour)

	filter := bson.M{
		"$and": []bson.M{
			{"password_expires_at": bson.M{"$lte": expiryThreshold}},
			{"password_expires_at": bson.M{"$gt": now}},
		},
		"$or": []bson.M{
			{"last_expiry_notification_sent_at": bson.M{"$exists": false}},
			{"last_expiry_notification_sent_at": bson.M{"$lt": notificationWindow}},
		},
		"account_status": account.StatusActive,
	}

	if lastID != "" {
		objID, _ := primitive.ObjectIDFromHex(lastID)
		filter["_id"] = bson.M{"$gt": objID}
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "_id", Value: 1}}).
		SetLimit(int64(batchSize))

	cursor, err := c.Find(ctx, filter, opts)
	if err != nil {
		log.Printf("ERROR: (Find) An error has occurred while finding users: %s", err)
		return nil, "", err
	}
	defer cursor.Close(ctx)

	var users []*entities.User
	if err := cursor.All(ctx, &users); err != nil {
		log.Printf("ERROR:(Cursor.All) An error has occurred while finding users: %s", err)
		return nil, "", err
	}

	var nextID string
	if len(users) > 0 {
		nextID = users[len(users)-1].ID.Hex()
	}

	return users, nextID, nil
}

// UpdateExpiryNotificationsSentDateBulk function is used to update the "last_expiry_notification_sent" field for
// a batch of users.
func (r *UserRepository) UpdateExpiryNotificationsSentDateBulk(ctx context.Context, ids []primitive.ObjectID) error {
	c := r.Client.Database(r.DbName).Collection(r.CollName)

	filter := bson.M{"_id": bson.M{"$in": ids}}
	update := bson.M{
		"$set": bson.M{
			"last_expiry_notification_sent_at": time.Now(),
		},
	}

	_, err := c.UpdateMany(ctx, filter, update)
	return err
}

// UpdateExpiryNotificationSentDate function is used to update the "last_expiry_notification_sent" field for
// a user that is processed.
func (r *UserRepository) UpdateExpiryNotificationSentDate(ctx context.Context, userID primitive.ObjectID) error {
	log.Printf("DEBUG: UpdateExpiryNotificationSentDate repository function has been called!")
	c := r.Client.Database(r.DbName).Collection(r.CollName)

	filter := bson.M{"_id": userID}
	update := bson.M{
		"$set": bson.M{
			"last_expiry_notification_sent_at": time.Now(),
		},
	}

	_, err := c.UpdateOne(ctx, filter, update)
	return err
}

// FindUserByID finds users by ID,.
func (r *UserRepository) FindUserByID(ctx context.Context, id primitive.ObjectID) (*entities.User, error) {
	var user entities.User
	c := r.Client.Database(r.DbName).Collection(r.CollName)
	if err := c.FindOne(ctx, bson.M{"_id": id}).Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

// ExistsByUsername reports whether a username already exists.
func (r *UserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var user entities.User
	c := r.Client.Database(r.DbName).Collection(r.CollName)

	filter := bson.M{"username": username}
	err := c.FindOne(ctx, filter).Decode(&user)

	if err == nil {
		return true, nil
	}

	if errors.Is(err, mongo.ErrNoDocuments) {
		return false, nil
	}

	return false, err
}

// ExistsByEmail reports whether an email already exists.
func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var user entities.User
	c := r.Client.Database(r.DbName).Collection(r.CollName)

	filter := bson.M{"email": email}
	err := c.FindOne(ctx, filter).Decode(&user)

	if err == nil {
		return true, nil
	}
	if errors.Is(err, mongo.ErrNoDocuments) {
		return false, nil
	}

	return false, err
}
