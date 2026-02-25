package repositories

import (
	"context"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
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

// CreateRating creates or updates a RATED relationship between a user and a song.
func (r *GraphRelationRepository) CreateRating(ctx context.Context, rating entities.Rating) error {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		return tx.Run(
			ctx,
			`MATCH (u:User {user_id: $user_id}), (s:Song {song_id: $song_id})
MERGE (u)-[r:RATED]->(s)
SET r.rating = $rating`,
			map[string]any{
				"user_id": rating.UserID,
				"song_id": rating.SongID,
				"rating":  rating.Value,
			},
		)
	})
	return err
}

// CreateListened creates a LISTENED relationship between a user and a song.
func (r *GraphRelationRepository) CreateListened(ctx context.Context, listened entities.Listened) error {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		return tx.Run(
			ctx,
			`MATCH (u:User {user_id: $user_id}), (s:Song {song_id: $song_id})
MERGE (u)-[:LISTENED]->(s)`,
			map[string]any{
				"user_id": listened.UserID,
				"song_id": listened.SongID,
			},
		)
	})
	return err
}

// CreateArtistSubscription creates a SUBSCRIBED relationship between a user and an artist.
func (r *GraphRelationRepository) CreateArtistSubscription(ctx context.Context, sub entities.ArtistSubscription) error {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		return tx.Run(
			ctx,
			`MATCH (u:User {user_id: $user_id}), (a:Artist {artist_id: $artist_id})
MERGE (u)-[:SUBSCRIBED]->(a)`,
			map[string]any{
				"user_id":   sub.UserID,
				"artist_id": sub.ArtistID,
			},
		)
	})
	return err
}

// CreateGenreSubscription creates a SUBSCRIBED_GENRE relationship between a user and a genre.
func (r *GraphRelationRepository) CreateGenreSubscription(ctx context.Context, sub entities.GenreSubscription) error {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		return tx.Run(
			ctx,
			`MATCH (u:User {user_id: $user_id}), (g:Genre {genre_id: $genre_id})
MERGE (u)-[:SUBSCRIBED_GENRE]->(g)`,
			map[string]any{
				"user_id":  sub.UserID,
				"genre_id": sub.GenreID,
			},
		)
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

// CreateSongHasGenre creates a HAS_GENRE relationship between a song and a genre.
func (r *GraphRelationRepository) CreateSongHasGenre(ctx context.Context, songID string, genreID string) error {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		return tx.Run(
			ctx,
			`MATCH (s:Song {song_id: $song_id}), (g:Genre {genre_id: $genre_id})
MERGE (s)-[:HAS_GENRE]->(g)`,
			map[string]any{
				"song_id":  songID,
				"genre_id": genreID,
			},
		)
	})
	return err
}

// CreateSongByArtist creates a BY relationship between a song and an artist.
func (r *GraphRelationRepository) CreateSongByArtist(ctx context.Context, songID string, artistID string) error {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		return tx.Run(
			ctx,
			`MATCH (s:Song {song_id: $song_id}), (a:Artist {artist_id: $artist_id})
MERGE (s)-[:BY]->(a)`,
			map[string]any{
				"song_id":   songID,
				"artist_id": artistID,
			},
		)
	})
	return err
}

// CreateArtistHasGenre creates a HAS_GENRE relationship between an artist and a genre.
func (r *GraphRelationRepository) CreateArtistHasGenre(ctx context.Context, artistID string, genreID string) error {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		return tx.Run(
			ctx,
			`MATCH (a:Artist {artist_id: $artist_id}), (g:Genre {genre_id: $genre_id})
MERGE (a)-[:HAS_GENRE]->(g)`,
			map[string]any{
				"artist_id": artistID,
				"genre_id":  genreID,
			},
		)
	})
	return err
}

// DeleteRating deletes a RATED relationship between a user and a song.
func (r *GraphRelationRepository) DeleteRating(ctx context.Context, userID string, songID string) error {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		return tx.Run(
			ctx,
			`MATCH (u:User {user_id: $user_id})-[r:RATED]->(s:Song {song_id: $song_id}) DELETE r`,
			map[string]any{
				"user_id": userID,
				"song_id": songID,
			},
		)
	})
	return err
}

// DeleteArtistSubscription deletes a SUBSCRIBED relationship between a user and an artist.
func (r *GraphRelationRepository) DeleteArtistSubscription(ctx context.Context, userID string, artistID string) error {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		return tx.Run(
			ctx,
			`MATCH (u:User {user_id: $user_id})-[r:SUBSCRIBED]->(a:Artist {artist_id: $artist_id}) DELETE r`,
			map[string]any{
				"user_id":   userID,
				"artist_id": artistID,
			},
		)
	})
	return err
}

// DeleteGenreSubscription deletes a SUBSCRIBED_GENRE relationship between a user and a genre.
func (r *GraphRelationRepository) DeleteGenreSubscription(ctx context.Context, userID string, genreID string) error {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		return tx.Run(
			ctx,
			`MATCH (u:User {user_id: $user_id})-[r:SUBSCRIBED_GENRE]->(g:Genre {genre_id: $genre_id}) DELETE r`,
			map[string]any{
				"user_id":  userID,
				"genre_id": genreID,
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
