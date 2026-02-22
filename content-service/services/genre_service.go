package services

import (
	"context"
	"errors"
	"time"

	"github.com/avast/retry-go"
	commondtos "github.com/vanjmali/spotlite/common-lib/dtos"
	"github.com/vanjmali/spotlite/common-lib/events"
	"github.com/vanjmali/spotlite/common-lib/logging"
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
	r   *repositories.GenreRepository
	jsc *events.JetStreamClient
	tr  trace.Tracer
}

// NewArtistService builds a ArtistService with repository.
func NewGenreService(r repositories.GenreRepository, jsc events.JetStreamClient) *GenreService {
	tr := otel.Tracer("content-service/genre-service")
	s := GenreService{r: &r, jsc: &jsc, tr: tr}

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
		logging.Errorf(ctx, "error converting to genre entity: %v", err)
		return err
	}

	err = s.r.Create(createCtx, *genreEntity)
	if err != nil {
		createSpan.RecordError(err)
		createSpan.End()
		logging.Errorf(ctx, "error creating genre in database: %v", err)
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
	updateCtx, updateSpan := s.tr.Start(ctx, "genre.update")
	defer updateSpan.End()

	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		updateSpan.RecordError(err)
		return nil, ErrObjectIdCastFailed
	}

	update := make(map[string]any)
	if dto.Name != nil {
		update["name"] = *dto.Name
	}

	if len(update) == 0 {
		err := errors.New("no fields to update")
		updateSpan.RecordError(err)
		return nil, err
	}

	genre, err := s.r.UpdateByID(updateCtx, id, update)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			updateSpan.RecordError(err)
			return nil, ErrGenreNotFound
		}
		updateSpan.RecordError(err)
		return nil, err
	}

	eventCtx, eventSpan := s.tr.Start(ctx, "genre.update.update_event")
	defer eventSpan.End()

	aep := toGenreUpdatedEvent(genre.ID.Hex(), genre.Name)

	err = retry.Do(
		func() error {
			return s.jsc.Publish(eventCtx, events.SUBJECT_ENTITY_UPDATED, aep)
		},
		retry.Attempts(3),
		retry.Delay(time.Second),
		retry.DelayType(retry.BackOffDelay),
		retry.Context(eventCtx),
	)
	if err != nil {
		logging.Errorf(eventCtx, "failed to publish entity updated event: %v", err)
		eventSpan.RecordError(err)
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
	return commondtos.ListWithPagination(ctx, p, filter, s.r.FindAll)
}

func (s *GenreService) Exists(ctx context.Context, genreIDstr string) (bool, error) {
	ctx, span := s.tr.Start(ctx, "genres.exists")
	defer span.End()

	genreID, err := primitive.ObjectIDFromHex(genreIDstr)
	if err != nil {
		span.RecordError(err)
		return false, err
	}

	exists, err := s.r.Exists(ctx, genreID)
	if err != nil {
		span.RecordError(err)
		return false, err
	}

	return exists, nil
}

func toGenreUpdatedEvent(genreID string, genreName string) *events.EntityUpdatedEventPayload {
	return &events.EntityUpdatedEventPayload{
		EntityID:   genreID,
		EntityName: genreName,
	}
}
