package repositories

import (
	"context"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/vanjmali/spotlite/recommendation-service/entities"
)

// ArtistNodeRepository provides data access for artist nodes in the graph.
type ArtistNodeRepository struct {
	Driver neo4j.DriverWithContext
}

// NewArtistNodeRepository constructs an ArtistNodeRepository.
func NewArtistNodeRepository(driver neo4j.DriverWithContext) *ArtistNodeRepository {
	return &ArtistNodeRepository{Driver: driver}
}

// Create creates an artist node in the graph.
func (r *ArtistNodeRepository) Create(ctx context.Context, artist entities.ArtistNode) error {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		return tx.Run(
			ctx,
			`MERGE (a:Artist {artist_id: $artist_id}) SET a.name = $name`,
			map[string]any{
				"artist_id": artist.ArtistID,
				"name":      artist.Name,
			},
		)
	})
	return err
}
