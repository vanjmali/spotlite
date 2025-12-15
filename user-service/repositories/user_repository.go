package repositories

import (
	"context"
	"errors"
	"time"
	"time"

	"github.com/vanjmali/spotlite/user-service/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	TokenExpiredErr = errors.New("invalid token")
)

type UserRepository struct {
	DbName   string
	CollName string
	Client   *mongo.Client
}

func NewUserRepository(dbName string, collName string, c *mongo.Client) *UserRepository {
	r := UserRepository{Client: c, DbName: dbName, CollName: collName}
	return &r
}

// Create func, inserts a new user into the database,
func (r *UserRepository) Create(ctx context.Context, user entities.User) error {
	c := r.Client.Database(r.DbName).Collection(r.CollName)

	_, err := c.InsertOne(ctx, user)
	if err != nil {
		return err
	}

	return nil
}

// ActiveAndRevokeToken func, that activates the account and revokes the token in one database trip,
func (r *UserRepository) ActiveAndRevokeToken(ctx context.Context, token string) error {
	c := r.Client.Database(r.DbName).Collection(r.CollName)

	// define filtering parameters,
	filter := bson.M{
		"account_status":           entities.StatusInactive,
		"email_verification.type":  entities.AccountVerification,
		"email_verification.token": token,
	}

	// define set (set a field value to a new one) and unset (fully remove a field) operations,
	update := bson.M{
		"$set":   bson.M{"account_status": entities.StatusActive},
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

	return TokenExpiredErr
}

func (r *UserRepository) SetLoginOtp(ctx context.Context, userId primitive.ObjectID, hash string, expiry time.Time) error {
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

func (r *UserRepository) ClearLoginOtp(ctx context.Context, userId primitive.ObjectID) error {
	c := r.Client.Database(r.DbName).Collection(r.CollName)

	_, err := c.UpdateOne(ctx,
		bson.M{"_id": userId},
		bson.M{"$unset": bson.M{"otp_code": ""}},
	)
	return err
}

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
func (r *UserRepository) FindUserByID(ctx context.Context, id primitive.ObjectID) (*entities.User, error) {
	var user entities.User
	c := r.Client.Database(r.DbName).Collection(r.CollName)
	if err := c.FindOne(ctx, bson.M{"_id": id}).Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

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
