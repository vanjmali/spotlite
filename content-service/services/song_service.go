package services

import (
	"context"
	"errors"
	"log"

	"github.com/vanjmali/spotlite/content/dtos"
	"github.com/vanjmali/spotlite/content/entities"
	"github.com/vanjmali/spotlite/content/mappers"
	"github.com/vanjmali/spotlite/content/repositories"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var (
	ErrSongNotFound = errors.New("song not found")
)

type SongService struct {
	songRepo      *repositories.SongRepository
	artistService *ArtistService
	tr            trace.Tracer
}

// NewSongService creates and returns a new SongService with the provided repository and artist service.
func NewSongService(songRepo repositories.SongRepository, artistService ArtistService) *SongService {
	tr := otel.Tracer("song-service/song-service")
	s := SongService{songRepo: &songRepo, artistService: &artistService, tr: tr}

	return &s
}

// Create creates a new song with the provided data, resolving associated artists.
func (s *SongService) Create(ctx context.Context, songDto *dtos.SongDto) error {
	ctx, span := s.tr.Start(ctx, "song.create")
	defer span.End()

	resolveCtx, resolveSpan := s.tr.Start(ctx, "song.create.resolve_artists")
	embeddedArtists := make([]entities.Artist, 0)

	for _, artistIdStr := range songDto.ArtistIds {
		artist, err := s.artistService.FindArtistByID(resolveCtx, artistIdStr)
		if err != nil {
			resolveSpan.RecordError(err)
			resolveSpan.End()

			switch {
			case errors.Is(err, ErrObjectIdCastFailed):
				return ErrObjectIdCastFailed
			case errors.Is(err, ErrArtistNotFound):
				return ErrArtistNotFound
			default:
				return err
			}
		}

		embeddedArtists = append(embeddedArtists, entities.Artist{
			ID:          artist.ID,
			Name:        artist.Name,
			Genres:      artist.Genres,
			Description: artist.Description,
		})
	}

	resolveSpan.End()

	_, mapSpan := s.tr.Start(ctx, "song.create.map_entity")
	songEntity, err := mappers.ToSongEntity(songDto, embeddedArtists)
	if err != nil {
		mapSpan.RecordError(err)
		mapSpan.End()
		log.Printf("Error converting to song entity: %v", err)
		return err
	}
	mapSpan.End()

	createCtx, createSpan := s.tr.Start(ctx, "song.create.create_song")
	err = s.songRepo.Create(createCtx, *songEntity)
	if err != nil {
		createSpan.RecordError(err)
		createSpan.End()
		log.Printf("Error creating song in database: %v", err)
		return err
	}

	createSpan.End()
	return nil

}

// FindSongById retrieves a single song by its ID.
func (s *SongService) FindSongById(ctx context.Context, idStr string) (*entities.Song, error) {
	ctx, span := s.tr.Start(ctx, "song.find_by_id")
	defer span.End()

	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		span.RecordError(err)
		span.End()
		return nil, ErrObjectIdCastFailed
	}

	song, err := s.songRepo.FindByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.End()
		return nil, ErrSongNotFound
	}

	return song, nil
}

// GetSongs retrieves a paginated list of songs with optional filtering by title, genre, or artist ID.
func (s *SongService) GetSongs(ctx context.Context, q dtos.SongQueryDto) (*dtos.SongListResponseDto, error) {
	ctx, span := s.tr.Start(ctx, "song.get_all")
	defer span.End()

	if q.Page <= 1 {
		q.Page = 1
	}
	if q.Size <= 0 || q.Size > 50 {
		q.Size = 10
	}

	skip := int64((q.Page - 1) * q.Size)
	limit := int64(q.Size)

	filter := bson.M{}

	if q.Title != "" {
		filter["title"] = bson.M{
			"$regex":   q.Title,
			"$options": "i",
		}
	}

	if q.Genre != "" {
		filter["genre"] = q.Genre
	}

	if q.ArtistID != "" {
		artistId, err := primitive.ObjectIDFromHex(q.ArtistID)
		if err != nil {
			return nil, ErrObjectIdCastFailed
		}
		filter["artists._id"] = artistId
	}

	items, total, err := s.songRepo.FindAll(ctx, filter, skip, limit)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return &dtos.SongListResponseDto{
		Items: items,
		Page:  q.Page,
		Size:  q.Size,
		Total: total,
	}, nil
}
