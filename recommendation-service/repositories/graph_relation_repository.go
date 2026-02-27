package repositories

import (
	"context"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/vanjmali/spotlite/common-lib/events"
	"github.com/vanjmali/spotlite/recommendation-service/entities"
)

// GraphRelationRepository provides data access for graph relationships.
type GraphRelationRepository struct {
	Driver neo4j.DriverWithContext
}

// NewGraphRelationRepository constructs a GraphRelationRepository.
func NewGraphRelationRepository(driver neo4j.DriverWithContext) *GraphRelationRepository {
	return &GraphRelationRepository{Driver: driver}
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

// CreateSongBelongsTo creates a BELONGS_TO relationship between a song and an album.
func (r *GraphRelationRepository) CreateSongBelongsTo(ctx context.Context, songID string, albumID string) error {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		return tx.Run(
			ctx,
			`MATCH (s:Song {song_id: $song_id}), (a:Album {album_id: $album_id})
MERGE (s)-[:BELONGS_TO]->(a)`,
			map[string]any{
				"song_id":  songID,
				"album_id": albumID,
			},
		)
	})
	return err
}

// GetUserRatings retrieves all songs a user has rated with their ratings.
func (r *GraphRelationRepository) GetUserRatings(ctx context.Context, userID string) ([]entities.Rating, error) {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		res, err := tx.Run(
			ctx,
			`MATCH (u:User {user_id: $user_id})-[r:RATED]->(s:Song)
RETURN u.user_id, s.song_id, r.rating`,
			map[string]any{"user_id": userID},
		)
		if err != nil {
			return nil, err
		}

		var ratings []entities.Rating
		for res.Next(ctx) {
			record := res.Record()
			ratings = append(ratings, entities.Rating{
				UserID: record.Values[0].(string),
				SongID: record.Values[1].(string),
				Value:  int(record.Values[2].(int64)),
			})
		}

		return ratings, nil
	})
	if err != nil {
		return nil, err
	}

	return result.([]entities.Rating), nil
}

// GetUserListenHistory retrieves all songs a user has listened to.
func (r *GraphRelationRepository) GetUserListenHistory(ctx context.Context, userID string, limit int) ([]string, error) {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		res, err := tx.Run(
			ctx,
			`MATCH (u:User {user_id: $user_id})-[:LISTENED]->(s:Song)
RETURN s.song_id
LIMIT $limit`,
			map[string]any{
				"user_id": userID,
				"limit":   limit,
			},
		)
		if err != nil {
			return nil, err
		}

		var songIDs []string
		for res.Next(ctx) {
			record := res.Record()
			songIDs = append(songIDs, record.Values[0].(string))
		}

		return songIDs, nil
	})
	if err != nil {
		return nil, err
	}

	return result.([]string), nil
}

// GetUserArtistSubscriptions retrieves all artists a user is subscribed to.
func (r *GraphRelationRepository) GetUserArtistSubscriptions(ctx context.Context, userID string) ([]string, error) {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		res, err := tx.Run(
			ctx,
			`MATCH (u:User {user_id: $user_id})-[:SUBSCRIBED]->(a:Artist)
RETURN a.artist_id`,
			map[string]any{"user_id": userID},
		)
		if err != nil {
			return nil, err
		}

		var artistIDs []string
		for res.Next(ctx) {
			record := res.Record()
			artistIDs = append(artistIDs, record.Values[0].(string))
		}

		return artistIDs, nil
	})
	if err != nil {
		return nil, err
	}

	return result.([]string), nil
}

// GetUserGenreSubscriptions retrieves all genres a user is subscribed to.
func (r *GraphRelationRepository) GetUserGenreSubscriptions(ctx context.Context, userID string) ([]string, error) {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		res, err := tx.Run(
			ctx,
			`MATCH (u:User {user_id: $user_id})-[:SUBSCRIBED_GENRE]->(g:Genre)
RETURN g.genre_id`,
			map[string]any{"user_id": userID},
		)
		if err != nil {
			return nil, err
		}

		var genreIDs []string
		for res.Next(ctx) {
			record := res.Record()
			genreIDs = append(genreIDs, record.Values[0].(string))
		}

		return genreIDs, nil
	})
	if err != nil {
		return nil, err
	}

	return result.([]string), nil
}

// GetSongsByGenre finds all songs belonging to a specific genre.
func (r *GraphRelationRepository) GetSongsByGenre(ctx context.Context, genreID string, limit int) ([]string, error) {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		res, err := tx.Run(
			ctx,
			`MATCH (s:Song)-[:HAS_GENRE]->(g:Genre {genre_id: $genre_id})
RETURN s.song_id
LIMIT $limit`,
			map[string]any{
				"genre_id": genreID,
				"limit":    limit,
			},
		)
		if err != nil {
			return nil, err
		}

		var songIDs []string
		for res.Next(ctx) {
			record := res.Record()
			songIDs = append(songIDs, record.Values[0].(string))
		}

		return songIDs, nil
	})
	if err != nil {
		return nil, err
	}

	return result.([]string), nil
}

// GetSongsByArtist finds all songs by a specific artist.
func (r *GraphRelationRepository) GetSongsByArtist(ctx context.Context, artistID string, limit int) ([]string, error) {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		res, err := tx.Run(
			ctx,
			`MATCH (s:Song)-[:BY]->(a:Artist {artist_id: $artist_id})
RETURN s.song_id
LIMIT $limit`,
			map[string]any{
				"artist_id": artistID,
				"limit":     limit,
			},
		)
		if err != nil {
			return nil, err
		}

		var songIDs []string
		for res.Next(ctx) {
			record := res.Record()
			songIDs = append(songIDs, record.Values[0].(string))
		}

		return songIDs, nil
	})
	if err != nil {
		return nil, err
	}

	return result.([]string), nil
}

// GetSongsByAlbum finds all songs in a specific album.
func (r *GraphRelationRepository) GetSongsByAlbum(ctx context.Context, albumID string) ([]string, error) {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		res, err := tx.Run(
			ctx,
			`MATCH (s:Song)-[:BELONGS_TO]->(a:Album {album_id: $album_id})
RETURN s.song_id`,
			map[string]any{"album_id": albumID},
		)
		if err != nil {
			return nil, err
		}

		var songIDs []string
		for res.Next(ctx) {
			record := res.Record()
			songIDs = append(songIDs, record.Values[0].(string))
		}

		return songIDs, nil
	})
	if err != nil {
		return nil, err
	}

	return result.([]string), nil
}

// GetHighlyRatedSongs retrieves songs with average rating above a threshold.
func (r *GraphRelationRepository) GetHighlyRatedSongs(ctx context.Context, minRating float64, limit int) ([]string, error) {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		res, err := tx.Run(
			ctx,
			`MATCH (s:Song)<-[r:RATED]-()
WITH s, avg(r.rating) as avg_rating, count(r) as rating_count
WHERE avg_rating >= $min_rating AND rating_count >= 3
RETURN s.song_id
ORDER BY avg_rating DESC, rating_count DESC
LIMIT $limit`,
			map[string]any{
				"min_rating": minRating,
				"limit":      limit,
			},
		)
		if err != nil {
			return nil, err
		}

		var songIDs []string
		for res.Next(ctx) {
			record := res.Record()
			songIDs = append(songIDs, record.Values[0].(string))
		}

		return songIDs, nil
	})
	if err != nil {
		return nil, err
	}

	return result.([]string), nil
}

// GetRecommendedSongsForUser generates personalized recommendations based on user's subscriptions and ratings.
func (r *GraphRelationRepository) GetRecommendedSongsForUser(ctx context.Context, userID string, limit int) ([]string, error) {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		res, err := tx.Run(
			ctx,
			`MATCH (u:User {user_id: $user_id})
WITH u,
     [(u)-[:SUBSCRIBED]->(:Artist)<-[:BY]-(s:Song) | s] AS artist_songs,
     [(u)-[:SUBSCRIBED_GENRE]->(:Genre)<-[:HAS_GENRE]-(s:Song) | s] AS genre_songs,
     [(u)-[r:RATED]->(:Song)-[:HAS_GENRE]->(:Genre)<-[:HAS_GENRE]-(s:Song) WHERE r.rating >= 4 | s] AS similar_genre_songs
WITH u, artist_songs + genre_songs + similar_genre_songs as candidate_songs
UNWIND candidate_songs as s
WITH u, s
WHERE s IS NOT NULL AND NOT (u)-[:RATED]->(s) AND NOT (u)-[:LISTENED]->(s)
OPTIONAL MATCH (s)<-[r:RATED]->()
WITH s, avg(r.rating) as avg_rating, count(r) as rating_count
RETURN s.song_id, avg_rating, rating_count
ORDER BY avg_rating DESC, rating_count DESC
LIMIT $limit`,
			map[string]any{
				"user_id": userID,
				"limit":   limit,
			},
		)
		if err != nil {
			return nil, err
		}

		var songIDs []string
		for res.Next(ctx) {
			record := res.Record()
			songIDs = append(songIDs, record.Values[0].(string))
		}

		return songIDs, nil
	})
	if err != nil {
		return nil, err
	}

	return result.([]string), nil
}

// GetSimilarUsers finds users with similar rating patterns (collaborative filtering).
func (r *GraphRelationRepository) GetSimilarUsers(ctx context.Context, userID string, limit int) ([]string, error) {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		res, err := tx.Run(
			ctx,
			`MATCH (u:User {user_id: $user_id})-[r1:RATED]->(s:Song)<-[r2:RATED]-(other:User)
WHERE other.user_id <> $user_id AND abs(r1.rating - r2.rating) <= 1
WITH other, count(s) as common_songs, avg(abs(r1.rating - r2.rating)) as avg_diff
WHERE common_songs >= 3
RETURN other.user_id
ORDER BY common_songs DESC, avg_diff ASC
LIMIT $limit`,
			map[string]any{
				"user_id": userID,
				"limit":   limit,
			},
		)
		if err != nil {
			return nil, err
		}

		var userIDs []string
		for res.Next(ctx) {
			record := res.Record()
			userIDs = append(userIDs, record.Values[0].(string))
		}

		return userIDs, nil
	})
	if err != nil {
		return nil, err
	}

	return result.([]string), nil
}

// GetCollaborativeRecommendations generates recommendations based on similar users' preferences.
func (r *GraphRelationRepository) GetCollaborativeRecommendations(ctx context.Context, userID string, limit int) ([]string, error) {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		res, err := tx.Run(
			ctx,
			`MATCH (u:User {user_id: $user_id})-[r1:RATED]->(s1:Song)<-[r2:RATED]-(similar:User)
WHERE similar.user_id <> $user_id AND abs(r1.rating - r2.rating) <= 1
WITH u, similar, count(s1) as common_songs
WHERE common_songs >= 1
MATCH (similar)-[r3:RATED]->(s2:Song)
WHERE r3.rating >= 4 AND NOT (u)-[:RATED]->(s2) AND NOT (u)-[:LISTENED]->(s2)
RETURN s2.song_id, count(similar) as similar_user_count, avg(r3.rating) as avg_rating
ORDER BY similar_user_count DESC, avg_rating DESC
LIMIT $limit`,
			map[string]any{
				"user_id": userID,
				"limit":   limit,
			},
		)
		if err != nil {
			return nil, err
		}

		var songIDs []string
		for res.Next(ctx) {
			record := res.Record()
			songIDs = append(songIDs, record.Values[0].(string))
		}

		return songIDs, nil
	})
	if err != nil {
		return nil, err
	}

	return result.([]string), nil
}

// GetSongRatingStats retrieves average rating and count for a specific song.
func (r *GraphRelationRepository) GetSongRatingStats(ctx context.Context, songID string) (avgRating float64, ratingCount int64, err error) {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		res, err := tx.Run(
			ctx,
			`MATCH (s:Song {song_id: $song_id})<-[r:RATED]-()
WITH avg(r.rating) as avg_rating, count(r) as rating_count
RETURN COALESCE(avg_rating, 0.0) as avg_rating, rating_count`,
			map[string]interface{}{"song_id": songID},
		)
		if err != nil {
			return nil, err
		}

		if res.Next(ctx) {
			record := res.Record()
			avg := record.Values[0].(float64)
			count := record.Values[1].(int64)
			return map[string]interface{}{"avg": avg, "count": count}, nil
		}

		// No ratings found
		return map[string]interface{}{"avg": 0.0, "count": int64(0)}, nil
	})

	if err != nil {
		return 0.0, 0, err
	}

	stats := result.(map[string]interface{})
	return stats["avg"].(float64), stats["count"].(int64), nil
}

// GetSongArtists retrieves all artists for a specific song.
func (r *GraphRelationRepository) GetSongArtists(ctx context.Context, songID string) ([]string, error) {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		res, err := tx.Run(
			ctx,
			`MATCH (s:Song {song_id: $song_id})-[:BY]->(a:Artist)
RETURN a.name`,
			map[string]interface{}{"song_id": songID},
		)
		if err != nil {
			return nil, err
		}

		var artistNames []string
		for res.Next(ctx) {
			record := res.Record()
			artistNames = append(artistNames, record.Values[0].(string))
		}

		return artistNames, nil
	})

	if err != nil {
		return nil, err
	}

	return result.([]string), nil
}
