package services

import (
	"context"
	"errors"
	"log"

	"github.com/vanjmali/spotlite/content/dtos"
	"github.com/vanjmali/spotlite/content/entities"
	"github.com/vanjmali/spotlite/content/mappers"
	"github.com/vanjmali/spotlite/content/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var (
	ErrSongNotFound = errors.New("song not found")
)

type SongService struct {
	songRepo   *repositories.SongRepository
	artistRepo *repositories.ArtistRepository
	tr         trace.Tracer
}

func NewSongService(songRepo repositories.SongRepository, artistRepo repositories.ArtistRepository) *SongService {
	tr := otel.Tracer("song-service/song-service")
	s := SongService{songRepo: &songRepo, artistRepo: &artistRepo, tr: tr}

	return &s
}

func (s *SongService) Create(ctx context.Context, songDto *dtos.SongDto) error {
	ctx, span := s.tr.Start(ctx, "song.create")
	defer span.End()

	resolveCtx, resolveSpan := s.tr.Start(ctx, "song.create.resolve_artists")
	embeddedArtists := make([]entities.Artist, 0)

	for _, artistIdStr := range songDto.ArtistIds {
		artistId, err := primitive.ObjectIDFromHex(artistIdStr)
		if err != nil {
			resolveSpan.RecordError(err)
			resolveSpan.End()
			return ErrObjectIdCastFailed
		}

		artist, err := s.artistRepo.FindByID(resolveCtx, artistId)
		if err != nil {
			resolveSpan.RecordError(err)
			resolveSpan.End()
			return err
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

// func (s *SongService) DeleteSong(ctx context.Context, idStr string) error {
// 	ctx, span := s.tr.Start(ctx, "song-service.delete")
// 	defer span.End()

// 	_, parseSpan
// }
