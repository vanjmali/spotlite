package repositories

import (
	"context"
	"errors"

	"github.com/vanjmali/spotlite/user-service/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository struct {
	DbName   string
	CollName string
	Client   *mongo.Client
}

func NewRepository(dbName string, collName string, c *mongo.Client) *UserRepository {
	r := UserRepository{Client: c, DbName: dbName, CollName: collName}
	return &r
}

func (r *UserRepository) Create(ctx context.Context, user entities.User) error {
	c := r.Client.Database(r.DbName).Collection(r.CollName)

	_, err := c.InsertOne(ctx, user)
	if err != nil {
		return err
	}

	return nil
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
