package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/vanjmali/spotlite/common-lib/logging"
	"github.com/vanjmali/spotlite/recommendation-service/entities"
)

var (
	ErrNoData       = errors.New("no data found")
	ErrSongNotFound = errors.New("song not found")
)

// GraphRelationRepository provides data access for graph relationships.
type GraphRelationRepository struct {
	Driver neo4j.DriverWithContext
}

// NewGraphRelationRepository constructs a GraphRelationRepository.
func NewGraphRelationRepository(driver neo4j.DriverWithContext) *GraphRelationRepository {
	return &GraphRelationRepository{Driver: driver}
}

func (r *GraphRelationRepository) UpdateSongWithGenres(ctx context.Context, sn entities.SongNode) error {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	query := `
		MATCH (s:Song {song_id: $songId})
		SET s.title = $title, 
		    s.duration = $duration,
			s.artist_names = $artistNames
		
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
		"songId":      sn.SongID,
		"title":       sn.Title,
		"duration":    sn.Duration,
		"genreIds":    sn.GenreIDs,
		"artistNames": sn.Artists,
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

// SaveSongWithGenres saves a song and links it to its genres.
func (r *GraphRelationRepository) SaveSongWithGenres(ctx context.Context, sn entities.SongNode) error {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	query := `
		MERGE (s:Song {song_id: $songId})
		SET s.title = $title, 
			s.duration = $duration,
			s.artist_names = $artistNames
		WITH s
		UNWIND $genreIds AS genreId
		MATCH (g:Genre {genre_id: genreId})
		MERGE (s)-[:BELONGS_TO]->(g)
	`

	params := map[string]any{
		"songId":      sn.SongID,
		"title":       sn.Title,
		"duration":    sn.Duration,
		"genreIds":    sn.GenreIDs,
		"artistNames": sn.Artists,
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

func (r *GraphRelationRepository) UpdateGenre(ctx context.Context, gn entities.GenreNode) error {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(
			ctx,
			`MATCH (g:Genre {genre_id: $genreId})
             SET g.name = $name`,
			map[string]any{
				"genreId": gn.GenreID,
				"name":    gn.Name,
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
			logging.Errorf(ctx, "genre update failed: genre_id %s not found", gn.Name)
			return nil, err
		}
		return summary, nil
	})

	return err
}

// CreateGenreSubscription creates a SUBSCRIBED_GENRE relationship between a user and a genre.
func (r *GraphRelationRepository) CreateGenreSubscription(ctx context.Context, gs entities.GenreSubscription) error {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(
			ctx,
			`MATCH (u:User {user_id: $userId}), (g:Genre {genre_id: $genreId})
             MERGE (u)-[:SUBSCRIBED_TO]->(g)`,
			map[string]any{
				"userId":  gs.UserID,
				"genreId": gs.GenreID,
			},
		)
		if err != nil {
			return nil, err
		}

		return result.Consume(ctx)
	})

	return err
}

// CreateArtistSubscription creates a SUBSCRIBED_TO relationship between a user and an artist.
func (r *GraphRelationRepository) CreateArtistSubscription(ctx context.Context, as entities.ArtistSubscription) error {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(
			ctx,
			`MATCH (u:User {user_id: $userId}), (a:Artist {artist_id: $artistId})
             MERGE (u)-[:SUBSCRIBED_TO]->(a)`,
			map[string]any{
				"userId":   as.UserID,
				"artistId": as.ArtistID,
			},
		)
		if err != nil {
			return nil, err
		}

		return result.Consume(ctx)
	})

	return err
}

func (r *GraphRelationRepository) CreateRating(ctx context.Context, sr entities.SongRating) error {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(
			ctx,
			`MATCH (u:User {user_id: $userId}), (s:Song {song_id: $songId})
             MERGE (u)-[r:RATED]->(s)
			 SET r.value = $value`,
			map[string]any{
				"userId": sr.UserID,
				"songId": sr.SongID,
				"value":  sr.Value,
			},
		)
		if err != nil {
			return nil, err
		}

		return result.Consume(ctx)
	})

	return err
}

func (r *GraphRelationRepository) FindSubscriptionBasedRecommendations(ctx context.Context, userID string) ([]*entities.SongRecommendation, error) {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	query := `
		MATCH (u:User {user_id: $userId})-[:SUBSCRIBED_TO]->(g:Genre)<-[:BELONGS_TO]-(s:Song)
		WHERE NOT EXISTS {
			MATCH (u)-[r:RATED]->(s)
			WHERE r.value < 4
		}
		WITH DISTINCT s
		OPTIONAL MATCH (:User)-[all_r:RATED]->(s)
		RETURN 
			s.song_id AS songId, 
			s.title AS title, 
			s.duration AS duration, 
			s.artist_names AS artists, 
			COALESCE(avg(all_r.value), 0.0) AS avgRating
		LIMIT 5
	`

	params := map[string]any{
		"userId": userID,
	}

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		records, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		var songs []*entities.SongRecommendation
		for records.Next(ctx) {
			record := records.Record()

			songIdVal, _ := record.Get("songId")
			titleVal, _ := record.Get("title")
			durationVal, _ := record.Get("duration")
			artistsVal, _ := record.Get("artists")
			avgRatingVal, _ := record.Get("avgRating")

			var songId, title string
			if songIdVal != nil {
				songId = songIdVal.(string)
			}
			if titleVal != nil {
				title = titleVal.(string)
			}

			var duration int
			if durationVal != nil {
				switch v := durationVal.(type) {
				case int64:
					duration = int(v)
				case int:
					duration = v
				}
			}

			var artists []string
			if artistsVal != nil {
				if genericArray, ok := artistsVal.([]any); ok {
					for _, item := range genericArray {
						if strItem, isStr := item.(string); isStr {
							artists = append(artists, strItem)
						}
					}
				}
			}

			var average float64
			if avgRatingVal != nil {
				switch v := avgRatingVal.(type) {
				case float64:
					average = v
				case int64:
					average = float64(v)
				}
			}

			songs = append(songs, &entities.SongRecommendation{
				SongID:   songId,
				Title:    title,
				Duration: duration,
				Rating:   average,
				Artists:  artists,
			})
		}
		return songs, records.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get subscribed songs with average ratings: %w", err)
	}

	return result.([]*entities.SongRecommendation), nil
}

func (r *GraphRelationRepository) FindLikeBasedRecommendation(ctx context.Context, userID string) ([]*entities.SongRecommendation, error) {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	query := `
		MATCH (u:User)-[r:RATED]->(s:Song)
		WHERE r.value = 5 AND u.user_id <> $userId
		  AND NOT EXISTS {
			MATCH (:User {user_id: $userId})-[:SUBSCRIBED_TO]->(:Genre)<-[:BELONGS_TO]-(s)
		  }
		WITH s, count(r) AS fives
		ORDER BY fives DESC
		LIMIT 10
		
		OPTIONAL MATCH (:User)-[all_r:RATED]->(s)
		RETURN 
			s.song_id AS songId, 
			s.title AS title, 
			s.duration AS duration, 
			s.artist_names AS artists, 
			COALESCE(avg(all_r.value), 0.0) AS avgRating
	`

	params := map[string]any{
		"userId": userID,
	}

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		records, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		recommendations := make([]*entities.SongRecommendation, 0)

		for records.Next(ctx) {
			record := records.Record()

			songIdVal, _ := record.Get("songId")
			titleVal, _ := record.Get("title")
			durationVal, _ := record.Get("duration")
			artistsVal, _ := record.Get("artists")
			avgRatingVal, _ := record.Get("avgRating")

			var songId, title string
			if songIdVal != nil {
				songId = songIdVal.(string)
			}
			if titleVal != nil {
				title = titleVal.(string)
			}

			var duration int
			if durationVal != nil {
				switch v := durationVal.(type) {
				case int64:
					duration = int(v)
				case int:
					duration = v
				}
			}

			var artists []string
			if artistsVal != nil {
				if genericArray, ok := artistsVal.([]any); ok {
					for _, item := range genericArray {
						if strItem, isStr := item.(string); isStr {
							artists = append(artists, strItem)
						}
					}
				}
			}

			var average float64
			if avgRatingVal != nil {
				switch v := avgRatingVal.(type) {
				case float64:
					average = v
				case int64:
					average = float64(v)
				}
			}

			recommendations = append(recommendations, &entities.SongRecommendation{
				SongID:   songId,
				Title:    title,
				Duration: duration,
				Rating:   average,
				Artists:  artists,
			})
		}

		return recommendations, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get top rated unsubscribed songs: %w", err)
	}

	if result == nil {
		return []*entities.SongRecommendation{}, nil
	}

	return result.([]*entities.SongRecommendation), nil
}

func (r *GraphRelationRepository) DeleteSong(ctx context.Context, songID string) error {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	// ignores delete count, it is returned to avoid nilnil
	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			MATCH (s:Song {song_id: $songID})
			DETACH DELETE s
		`
		params := map[string]interface{}{
			"songID": songID,
		}

		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		summary, err := result.Consume(ctx)
		if err != nil {
			return nil, err
		}

		deletedCount := summary.Counters().NodesDeleted()
		if deletedCount == 0 {
			return nil, ErrSongNotFound
		}

		return deletedCount, nil
	})

	return err
}

func (r *GraphRelationRepository) UpdateRating(ctx context.Context, songID string, userID string, rating int) error {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	logging.Infof(ctx, "Updating rating: song=%s, user=%s, val=%d", songID, userID, rating)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(
			ctx,
			`MATCH (u:User {user_id: $userId})
             MATCH (s:Song {song_id: $songId})
			 MERGE (u)-[r:RATED]->(s)
			 SET r.value = $value`,
			map[string]any{
				"userId": userID,
				"songId": songID,
				"value":  rating,
			},
		)
		if err != nil {
			return nil, err
		}

		summary, err := result.Consume(ctx)
		if err != nil {
			return nil, err
		}

		if summary.Counters().PropertiesSet() == 0 && summary.Counters().RelationshipsCreated() == 0 {
			logging.Warnf(ctx, "Neo4j update had no effect. User %s or Song %s might be missing.", userID, songID)
		}

		return summary, nil
	})

	return err
}
