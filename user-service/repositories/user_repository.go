package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/vanjmali/spotlite/common-lib/account"
	"github.com/vanjmali/spotlite/common-lib/logging"
	"github.com/vanjmali/spotlite/user-service/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ErrTokenExpired signals that the verification token was not found or already used.
var (
	ErrTokenExpired         = errors.New("invalid token")
	ErrUserNotFound         = errors.New("user not found")
	ErrUsernameAlreadyTaken = errors.New("username already taken")
	ErrEmailAlreadyTaken    = errors.New("email already taken")
)

// UserRepository provides data access helpers for user documents.
type UserRepositoryMongo struct {
	DbName   string
	CollName string
	Client   *mongo.Client
}

func (r *UserRepositoryMongo) getCollection() *mongo.Collection {
	return r.Client.Database(r.DbName).Collection(r.CollName)
}

// NewUserRepositoryMongo constructs a UserRepository for the given database and collection.
func NewUserRepositoryMongo(dbName string, collName string, c *mongo.Client) *UserRepositoryMongo {
	r := UserRepositoryMongo{Client: c, DbName: dbName, CollName: collName}
	return &r
}

// EnsureUserIndexes creates unique indexes that guard against duplicate users.
func (r *UserRepositoryMongo) EnsureUserIndexes(ctx context.Context) error {
	c := r.getCollection()
	_, err := c.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "username", Value: 1}},
			Options: options.Index().
				SetName("users_username_unique").
				SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "email", Value: 1}},
			Options: options.Index().
				SetName("users_email_unique").
				SetUnique(true),
		},
	})
	return err
}

// Create func, inserts a new user into the database.
func (r *UserRepositoryMongo) Create(ctx context.Context, user entities.User) error {
	c := r.getCollection()

	_, err := c.InsertOne(ctx, user)
	if err != nil {
		// Check for duplicate key error and determine if it's related to username or email to return specific errors.
		var writeExc mongo.WriteException
		if errors.As(err, &writeExc) {
			for _, writeErr := range writeExc.WriteErrors {
				if writeErr.Code != 11000 {
					continue
				}
				msg := strings.ToLower(writeErr.Message)
				if strings.Contains(msg, "username") {
					return ErrUsernameAlreadyTaken
				}
				if strings.Contains(msg, "email") {
					return ErrEmailAlreadyTaken
				}
			}
		}

		var cmdErr mongo.CommandError
		if errors.As(err, &cmdErr) && cmdErr.Code == 11000 {
			msg := strings.ToLower(cmdErr.Message)
			if strings.Contains(msg, "username") {
				return ErrUsernameAlreadyTaken
			}
			if strings.Contains(msg, "email") {
				return ErrEmailAlreadyTaken
			}
		}

		return err
	}

	return nil
}

// ActiveAndRevokeToken func, that activates the account and revokes the token in one database trip.
func (r *UserRepositoryMongo) ActiveAndRevokeToken(ctx context.Context, token string) error {
	c := r.getCollection()

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

// UpdateVerificationToken replaces the verification token for an inactive account.
func (r *UserRepositoryMongo) UpdateVerificationToken(ctx context.Context, userID primitive.ObjectID, token string) error {
	c := r.getCollection()

	filter := bson.M{
		"_id":            userID,
		"account_status": account.StatusInactive,
	}

	update := bson.M{
		"$set": bson.M{
			"email_verification": bson.M{
				"type":  entities.AccountVerification,
				"token": token,
			},
			"updated_at": time.Now(),
		},
	}

	res, err := c.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if res.MatchedCount == 1 && res.ModifiedCount == 1 {
		return nil
	}

	return ErrUserNotFound
}

func (r *UserRepositoryMongo) SetHashPassowrd(ctx context.Context, userId primitive.ObjectID, passwordHash string, newTime, expiresAt time.Time) error {
	c := r.getCollection()

	_, err := c.UpdateByID(ctx,
		userId,
		bson.M{"$set": bson.M{
			"password":                 passwordHash,
			"password_last_changed_at": newTime,
			"password_expires_at":      expiresAt,
			"updated_at":               time.Now(),
		}},
	)
	return err
}

// SetLoginOtp stores the hashed OTP and expiry for a user.
func (r *UserRepositoryMongo) SetLoginOtp(
	ctx context.Context,
	userId primitive.ObjectID,
	hash string,
	expiry time.Time,
) error {
	c := r.getCollection()

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
func (r *UserRepositoryMongo) ClearLoginOtp(ctx context.Context, userId primitive.ObjectID) error {
	c := r.getCollection()

	_, err := c.UpdateOne(ctx,
		bson.M{"_id": userId},
		bson.M{"$unset": bson.M{"otp_code": ""}},
	)
	return err
}

// FindUserByEmail fetches a user document by email.
func (r *UserRepositoryMongo) FindUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	c := r.getCollection()

	var user entities.User
	filter := bson.M{"email": email}
	if err := c.FindOne(ctx, filter).Decode(&user); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

// FindUsersForExpiryNotification finds users which password expiry date is less than (today + daysBeforeExpiry), but
// also greater than today. The user also has to have an active account.
func (r *UserRepositoryMongo) FindUsersForExpiryNotification(
	ctx context.Context,
	daysUntilExpiry int,
	batchSize int,
	lastID string,
) ([]*entities.User, string, error) {
	c := r.getCollection()

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
		return nil, "", fmt.Errorf("ERROR: (Find) An error has occurred while finding users: %w", err)
	}

	defer cursor.Close(ctx)

	var users []*entities.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, "", fmt.Errorf("ERROR: (Cursor.All) An error has occurred while finding users: %w", err)
	}

	var nextID string
	if len(users) > 0 {
		nextID = users[len(users)-1].ID.Hex()
	}

	return users, nextID, nil
}

// UpdateExpiryNotificationSentDate function is used to update the "last_expiry_notification_sent" field for
// a user that is processed.
func (r *UserRepositoryMongo) UpdateExpiryNotificationSentDate(ctx context.Context, userID primitive.ObjectID) error {
	logging.Infof(ctx, "update_expiry_notification_sent repository called")
	c := r.getCollection()

	filter := bson.M{"_id": userID}
	update := bson.M{
		"$set": bson.M{
			"last_expiry_notification_sent_at": time.Now(),
		},
	}

	_, err := c.UpdateOne(ctx, filter, update)
	return err
}

// FindUserByID finds users by ID.
func (r *UserRepositoryMongo) FindUserByID(ctx context.Context, id primitive.ObjectID) (*entities.User, error) {
	var user entities.User
	c := r.getCollection()
	if err := c.FindOne(ctx, bson.M{"_id": id}).Decode(&user); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	return &user, nil
}

// ExistsByUsername reports whether a username already exists.
func (r *UserRepositoryMongo) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	c := r.getCollection()

	var user entities.User
	filter := bson.M{"username": username}
	if err := c.FindOne(ctx, filter).Decode(&user); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

// ExistsByEmail reports whether an email already exists.
func (r *UserRepositoryMongo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	c := r.getCollection()

	var user entities.User
	filter := bson.M{"email": email}
	if err := c.FindOne(ctx, filter).Decode(&user); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (r *UserRepositoryMongo) Delete(ctx context.Context, userID primitive.ObjectID) error {
	c := r.getCollection()

	filter := bson.M{"_id": userID}
	res, err := c.DeleteOne(ctx, filter)
	if err == nil && res.DeletedCount == 0 {
		return ErrUserNotFound
	}

	if err != nil {
		return err
	}

	return nil
}
