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

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		return tx.Run(
			ctx,
			`MERGE (u:User {user_id: $user_id}) SET u.username = $username`,
			map[string]any{
				"user_id":  user.UserID,
				"username": user.Username,
			},
		)
	})
	return err
}
