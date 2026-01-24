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

var ErrGenreNotFound = errors.New("genre not found")

type GenreService struct {
	r  *repositories.GenreRepository
	tr trace.Tracer
}

// NewArtistService builds a ArtistService with repository.
func NewGenreService(r repositories.GenreRepository) *GenreService {
	tr := otel.Tracer("content-service/genre-service")
	s := GenreService{r: &r, tr: tr}

	return &s
}

func (s *GenreService) Create(ctx context.Context, reqDto *dtos.GenreDto) error {
	ctx, span := s.tr.Start(ctx, "genre.create")
	defer span.End()

	createCtx, createSpan := s.tr.Start(ctx, "genre.create.create_genre")
	genreEntity, err := mappers.ToGenreEntity(reqDto)
	if err != nil {
		createSpan.RecordError(err)
		createSpan.End()
		log.Printf("Error converting to genre entity: %v", err)
		return err
	}

	err = s.r.Create(createCtx, *genreEntity)
	if err != nil {
		createSpan.RecordError(err)
		createSpan.End()
		log.Printf("Error creating genre in database: %v", err)
		return err
	}

	createSpan.End()

	return nil
}

func (s *GenreService) FindGenreByID(ctx context.Context, idStr string) (*entities.Genre, error) {
	ctx, span := s.tr.Start(ctx, "genre.find_by_id")
	defer span.End()

	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		span.RecordError(err)
		span.End()
		return nil, ErrObjectIdCastFailed
	}

	genre, err := s.r.FindByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.End()
		return nil, ErrGenreNotFound
	}

	return genre, nil
}
