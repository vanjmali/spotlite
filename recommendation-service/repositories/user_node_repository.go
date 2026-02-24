package repositories

import (
	"context"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/vanjmali/spotlite/recommendation-service/entities"
)

// UserNodeRepository provides data access for user nodes in the graph.
type UserNodeRepository struct {
	Driver neo4j.DriverWithContext
}

// NewUserNodeRepository constructs a UserNodeRepository.
func NewUserNodeRepository(driver neo4j.DriverWithContext) *UserNodeRepository {
	return &UserNodeRepository{Driver: driver}
}

// Create creates a user node in the graph.
func (r *UserNodeRepository) Create(ctx context.Context, user entities.UserNode) error {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		return tx.Run(
			ctx,
			`MERGE (u:User {user_id: $user_id}) SET u.username = $username`,
			map[string]interface{}{
				"user_id":  user.UserID,
				"username": user.Username,
			},
		)
	})
	return err
}

// Get retrieves a user node by user ID.
func (r *UserNodeRepository) Get(ctx context.Context, userID string) (*entities.UserNode, error) {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		res, err := tx.Run(
			ctx,
			`MATCH (u:User {user_id: $user_id}) RETURN u.user_id, u.username`,
			map[string]interface{}{"user_id": userID},
		)
		if err != nil {
			return nil, err
		}

		if res.Next(ctx) {
			record := res.Record()
			return &entities.UserNode{
				UserID:   record.Values[0].(string),
				Username: record.Values[1].(string),
			}, nil
		}

		return nil, nil
	})

	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}

	return result.(*entities.UserNode), nil
}

// Exists checks if a user node exists.
func (r *UserNodeRepository) Exists(ctx context.Context, userID string) (bool, error) {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		res, err := tx.Run(
			ctx,
			`MATCH (u:User {user_id: $user_id}) RETURN count(u) > 0 AS exists`,
			map[string]interface{}{"user_id": userID},
		)
		if err != nil {
			return false, err
		}

		if res.Next(ctx) {
			record := res.Record()
			return record.Values[0].(bool), nil
		}

		return false, nil
	})

	if err != nil {
		return false, err
	}

	return result.(bool), nil
}
