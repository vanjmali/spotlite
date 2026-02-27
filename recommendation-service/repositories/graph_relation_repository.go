package repositories

import (
	"context"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/vanjmali/spotlite/common-lib/events"
	"github.com/vanjmali/spotlite/common-lib/logging"
)

// GraphRelationRepository provides data access for graph relationships.
type GraphRelationRepository struct {
	Driver neo4j.DriverWithContext
}

// NewGraphRelationRepository constructs a GraphRelationRepository.
func NewGraphRelationRepository(driver neo4j.DriverWithContext) *GraphRelationRepository {
	return &GraphRelationRepository{Driver: driver}
}

func (r *GraphRelationRepository) UpdateSongWithGenres(ctx context.Context, e events.SongUpdatePayload) error {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	query := `
		MATCH (s:Song {song_id: $songId})
		SET s.title = $title, 
		    s.duration = $duration
		
		WITH s
		OPTIONAL MATCH (s)-[rg:BELONGS_TO]->(:Genre)
		DELETE rg

		WITH s
		CALL {
			WITH s
			UNWIND $genreIds AS genreId
			MATCH (g:Genre {genre_id: genreId})
			MERGE (s)-[:BELONGS_TO]->(g)
		}
	`

	params := map[string]any{
		"songId":   e.SongID,
		"title":    e.SongTitle,
		"duration": e.Duration,
		"genreIds": e.GenreIDs,
	}

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		return result.Consume(ctx)
	})

	return err
}

func (r *GraphRelationRepository) UpdateGenre(ctx context.Context, e events.EntityUpdatedEventPayload) error {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(
			ctx,
			`MATCH (g:Genre {genre_id: $genreId})
             SET g.name = $name`,
			map[string]any{
				"genreId": e.EntityID,
				"name":    e.EntityName,
			},
		)
		if err != nil {
			return nil, err
		}

		summary, err := result.Consume(ctx)
		if err != nil {
			return nil, err
		}

		// Catch silent failures!
		if summary.Counters().PropertiesSet() == 0 {
			logging.Errorf(ctx, "genre update failed: genre_id %s not found", e.EntityName)
			return nil, err
		}
		return summary, nil
	})

	return err
}

// SaveSongWithGenres saves a song and links it to its genres.
func (r *GraphRelationRepository) SaveSongWithGenres(ctx context.Context, e events.SongCreationPayload) error {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	query := `
		MERGE (s:Song {song_id: $songId})
		SET s.title = $title, s.duration = $duration
		WITH s
		UNWIND $genreIds AS genreId
		MATCH (g:Genre {genre_id: genreId})
		MERGE (s)-[:BELONGS_TO]->(g)
	`

	params := map[string]any{
		"songId":   e.SongID,
		"title":    e.SongTitle,
		"duration": e.Duration,
		"genreIds": e.GenreIDs,
	}

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		return result.Consume(ctx)
	})

	return err
}

// CreateGenreSubscription creates a SUBSCRIBED_GENRE relationship between a user and a genre.
func (r *GraphRelationRepository) CreateGenreSubscription(ctx context.Context, e events.GenreSubscriptionEventPayload) error {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(
			ctx,
			`MATCH (u:User {user_id: $userId}), (g:Genre {genre_id: $genreId})
             MERGE (u)-[:SUBSCRIBED_TO]->(g)`,
			map[string]any{
				"userId":  e.UserID,
				"genreId": e.GenreID,
			},
		)
		if err != nil {
			return nil, err
		}

		return result.Consume(ctx)
	})

	return err
}

func (r *GraphRelationRepository) CreateRating(ctx context.Context, e events.SongRatingPayload) error {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(
			ctx,
			`MATCH (u:User {user_id: $userId}), (s:Song {song_id: $songId})
             MERGE (u)-[r:RATED]->(s)
			 SET r.value = $value`,
			map[string]any{
				"userId": e.UserID,
				"songId": e.SongID,
				"value":  e.Value,
			},
		)
		if err != nil {
			return nil, err
		}

		return result.Consume(ctx)
	})

	return err
}
