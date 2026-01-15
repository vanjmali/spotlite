package services

import (
	"context"
	"errors"
	"log"

	"github.com/vanjmali/spotlite/content/dtos"
	"github.com/vanjmali/spotlite/content/mappers"
	"github.com/vanjmali/spotlite/content/repositories"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var (
	ErrObjectIdCastFailed = errors.New("failed to convert hex to objectId")
	ErrArtistNotFound     = errors.New("artist not found")
)

type ArtistService struct {
	r  *repositories.ArtistRepository
	tr trace.Tracer
}

// NewArtistService builds a ArtistService with repository.
func NewArtistService(r repositories.ArtistRepository) *ArtistService {
	tr := otel.Tracer("artist-service/artist-service")
	s := ArtistService{r: &r, tr: tr}

	return &s
}

// Create creates a new artist with the provided data.
func (s *ArtistService) Create(ctx context.Context, reqDto *dtos.ArtistDto) error {
	ctx, span := s.tr.Start(ctx, "artist.create")
	defer span.End()

	// Converts ArtistDto to Artist entity.
	// No uniqueness check for artist name is done here.
	createCtx, createSpan := s.tr.Start(ctx, "artist.create.create_artist")
	artistEntity, err := mappers.ToArtistEntity(reqDto)
	if err != nil {
		createSpan.RecordError(err)
		createSpan.End()
		log.Printf("Error converting to artist entity: %v", err)
		return err
	}

	// Insert the artist in database.
	err = s.r.Create(createCtx, *artistEntity)
	if err != nil {
		createSpan.RecordError(err)
		createSpan.End()
		log.Printf("Error creating artist in database: %v", err)
		return err
	}

	createSpan.End()

	return nil
}

// FindArtistByID retrieves a single artist by its ID.
func (s *ArtistService) FindArtistByID(ctx context.Context, idStr string) (*dtos.ArtistDto, error) {
	ctx, span := s.tr.Start(ctx, "artist.find_by_id")
	defer span.End()

	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		span.RecordError(err)
		return nil, ErrObjectIdCastFailed
	}

	artist, err := s.r.FindByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		return nil, ErrArtistNotFound
	}
	return artist, nil
}

// UpdateArtist updates an existing artist with the provided partial data.
func (s *ArtistService) UpdateArtist(ctx context.Context, idStr string, dto dtos.UpdateArtistDto) (*dtos.ArtistDto, error) {
	ctx, span := s.tr.Start(ctx, "artist.update_artist")
	defer span.End()

	_, parseSpan := s.tr.Start(ctx, "artist.update_artist.parse_id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		parseSpan.RecordError(err)
		parseSpan.End()
		return nil, ErrObjectIdCastFailed
	}
	parseSpan.End()

	_, buildSpan := s.tr.Start(ctx, "artist.update.build_update_doc")

	update := make(map[string]interface{})

	if dto.Name != nil {
		update["name"] = *dto.Name
	}
	if dto.Genres != nil {
		update["genres"] = *dto.Genres
	}
	if dto.Description != nil {
		update["description"] = *dto.Description
	}

	if len(update) == 0 {
		err := errors.New("no fields to update")
		buildSpan.RecordError(err)
		buildSpan.End()
		return nil, err
	}
	buildSpan.End()

	repoCtx, repoSpan := s.tr.Start(ctx, "artist.update.repository_update")

	updatedArtist, err := s.r.UpdateByID(repoCtx, id, update)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			repoSpan.RecordError(err)
			repoSpan.End()
			return nil, ErrArtistNotFound
		}
		repoSpan.RecordError(err)
		repoSpan.End()
		return nil, err
	}
	repoSpan.End()

	return updatedArtist, nil
}

// DeleteArtist deletes an artist by its ID.
func (s *ArtistService) DeleteArtist(ctx context.Context, idStr string) error {
	ctx, span := s.tr.Start(ctx, "artist.delete_artist")
	defer span.End()

	_, parseSpan := s.tr.Start(ctx, "artist.update_artist.parse_id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		parseSpan.RecordError(err)
		parseSpan.End()
		return ErrObjectIdCastFailed
	}
	parseSpan.End()

	repoCtx, repoSpan := s.tr.Start(ctx, "artist.delete.repository_delete")
	res, err := s.r.DeleteByID(repoCtx, id)
	if err != nil {
		repoSpan.RecordError(err)
		repoSpan.End()
		return err
	}

	if res.DeletedCount == 0 {
		err = ErrArtistNotFound
		repoSpan.RecordError(err)
		repoSpan.End()
		return err
	}
	repoSpan.End()
	return nil
}

// GetArtists retrieves a paginated list of artists with optional filtering by name or genre.
func (s *ArtistService) GetArtists(ctx context.Context, q dtos.ArtistQueryDto) (*dtos.ArtistListResponseDto, error) {
	ctx, span := s.tr.Start(ctx, "artists.get_all")
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

	if q.Name != "" {
		filter["name"] = bson.M{
			"$regex":   q.Name,
			"$options": "i",
		}
	}

	if q.Genre != "" {
		filter["genres"] = q.Genre
	}

	items, total, err := s.r.FindAll(ctx, filter, skip, limit)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return &dtos.ArtistListResponseDto{
		Items: items,
		Page:  q.Page,
		Size:  q.Size,
		Total: total,
	}, nil
}
