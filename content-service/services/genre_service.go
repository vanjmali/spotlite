package services

import (
	"context"
	"errors"
	"log"

	"github.com/vanjmali/spotlite/common-lib/pagination"
	"github.com/vanjmali/spotlite/content/dtos"
	"github.com/vanjmali/spotlite/content/entities"
	"github.com/vanjmali/spotlite/content/mappers"
	"github.com/vanjmali/spotlite/content/repositories"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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

func (s *GenreService) UpdateGenre(ctx context.Context, idStr string, dto dtos.UpdateGenreDto) (*entities.Genre, error) {
	ctx, span := s.tr.Start(ctx, "genre.update")
	defer span.End()

	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		span.RecordError(err)
		return nil, ErrObjectIdCastFailed
	}

	update := make(map[string]any)
	if dto.Name != nil {
		update["name"] = *dto.Name
	}

	if len(update) == 0 {
		err := errors.New("no fields to update")
		span.RecordError(err)
		return nil, err
	}

	genre, err := s.r.UpdateByID(ctx, id, update)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			span.RecordError(err)
			return nil, ErrGenreNotFound
		}
		span.RecordError(err)
		return nil, err
	}

	return genre, nil
}

func (s *GenreService) DeleteGenre(ctx context.Context, idStr string) error {
	ctx, span := s.tr.Start(ctx, "genre.delete")
	defer span.End()

	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		span.RecordError(err)
		return ErrObjectIdCastFailed
	}

	res, err := s.r.DeleteByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		return err
	}

	if res.DeletedCount == 0 {
		err = ErrGenreNotFound
		span.RecordError(err)
		return err
	}

	return nil
}

type GenresQuery struct {
	Page int
	Size int
	Name string
}

func (s *GenreService) GetGenres(ctx context.Context, q GenresQuery) (*dtos.GenreListResponseDto, error) {
	ctx, span := s.tr.Start(ctx, "genres.get_all")
	defer span.End()

	filter := bson.M{}
	if q.Name != "" {
		filter["name"] = bson.M{
			"$regex":   q.Name,
			"$options": "i",
		}
	}

	p := pagination.NewPagination(q.Page, q.Size)
	return listWithPagination(ctx, p, filter, s.r.FindAll)
}

func (s *GenreService) Exists(ctx context.Context, genreIDstr string) (bool, error) {
	ctx, span := s.tr.Start(ctx, "genres.exists")
	defer span.End()

	genreID, err := primitive.ObjectIDFromHex(genreIDstr)
	if err != nil {
		span.RecordError(err)
		return false, err
	}

	existsCtx, existsSpan := s.tr.Start(ctx, "genres.exists.existence_check")
	defer existsSpan.End()

	exists, err := s.r.Exists(existsCtx, genreID)
	if err != nil {
		existsSpan.RecordError(err)
		return false, err
	}

	return exists, nil
}
