package services

import (
	"context"
	"errors"
	"time"

	"github.com/avast/retry-go"
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

var (
	ErrObjectIdCastFailed = errors.New("failed to convert hex to objectId")
	ErrArtistNotFound     = errors.New("artist not found")
)

type ArtistService struct {
	r            *repositories.ArtistRepository
	genreService *GenreService
	jsc          *events.JetStreamClient
	tr           trace.Tracer
}

// NewArtistService builds a ArtistService with repository.
func NewArtistService(r repositories.ArtistRepository, genreService GenreService, jsc events.JetStreamClient) *ArtistService {
	tr := otel.Tracer("content-service/artist-service")
	s := ArtistService{r: &r, genreService: &genreService, tr: tr, jsc: &jsc}
	return &s
}

// Create creates a new artist with the provided data.
func (s *ArtistService) Create(ctx context.Context, reqDto *dtos.ArtistDto) error {
	ctx, span := s.tr.Start(ctx, "artist.create")
	defer span.End()

	resolveGenreCtx, resolveGenreSpan := s.tr.Start(ctx, "artist.create.resolve_genres")
	defer resolveGenreSpan.End()

	embeddedGenre := make([]entities.Genre, 0)
	genreIDs := []string{}

	for _, genresIdStr := range reqDto.GenreIds {
		genre, err := s.genreService.FindGenreByID(resolveGenreCtx, genresIdStr)
		if err != nil {
			resolveGenreSpan.RecordError(err)

			switch {
			case errors.Is(err, ErrObjectIdCastFailed):
				return ErrObjectIdCastFailed
			case errors.Is(err, ErrGenreNotFound):
				return ErrGenreNotFound
			default:
				return err
			}
		}

		embeddedGenre = append(embeddedGenre, entities.Genre{
			ID:   genre.ID,
			Name: genre.Name,
		})

		genreIDs = append(genreIDs, genre.ID.Hex())
	}

	// Converts ArtistDto to Artist entity.
	// No uniqueness check for artist name is done here.
	createCtx, createSpan := s.tr.Start(ctx, "artist.create.create_artist")
	defer createSpan.End()

	artistEntity, err := mappers.ToArtistEntity(reqDto, embeddedGenre)
	if err != nil {
		createSpan.RecordError(err)
		logging.Errorf(ctx, "error converting to artist entity: %v", err)
		return err
	}

	// Insert the artist in database.
	err = s.r.Create(createCtx, *artistEntity)
	if err != nil {
		createSpan.RecordError(err)
		logging.Errorf(ctx, "error creating artist in database: %v", err)
		return err
	}

	aep := toArtistCreatedEvent(genreIDs, artistEntity.ID.Hex(), artistEntity.Name)

	err = retry.Do(
		func() error {
			return s.jsc.Publish(createCtx, events.SUBJECT_ENTITY_CREATED, aep)
		},
		retry.Attempts(3),
		retry.Delay(time.Second),
		retry.DelayType(retry.BackOffDelay),
		retry.Context(createCtx),
	)
	if err != nil {
		logging.Errorf(createCtx, "failed to publish entity created event: %v", err)
	}

	return nil
}

// FindArtistByID retrieves a single artist by its ID.
func (s *ArtistService) FindArtistByID(ctx context.Context, idStr string) (*entities.Artist, error) {
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
func (s *ArtistService) UpdateArtist(ctx context.Context, idStr string, dto dtos.UpdateArtistDto) (*entities.Artist, error) {
	updateCtx, updateSpan := s.tr.Start(ctx, "artist.update_artist")
	defer updateSpan.End()

	_, parseSpan := s.tr.Start(updateCtx, "artist.update_artist.parse_id")
	defer parseSpan.End()
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		parseSpan.RecordError(err)
		return nil, ErrObjectIdCastFailed
	}
	_, buildSpan := s.tr.Start(ctx, "artist.update.build_update_doc")
	defer buildSpan.End()

	update := make(map[string]any)

	if dto.Name != nil {
		update["name"] = *dto.Name
	}
	if dto.GenreIds != nil {
		embeddedGenres := make([]entities.Genre, 0)
		for _, genreIdStr := range *dto.GenreIds {
			genre, err := s.genreService.FindGenreByID(ctx, genreIdStr)
			if err != nil {
				buildSpan.RecordError(err)
				switch {
				case errors.Is(err, ErrObjectIdCastFailed):
					return nil, ErrObjectIdCastFailed
				case errors.Is(err, ErrGenreNotFound):
					return nil, ErrGenreNotFound
				default:
					return nil, err
				}
			}

			embeddedGenres = append(embeddedGenres, entities.Genre{
				ID:   genre.ID,
				Name: genre.Name,
			})
		}
		update["genres"] = embeddedGenres
	}
	if dto.Description != nil {
		update["description"] = *dto.Description
	}

	if len(update) == 0 {
		err := errors.New("no fields to update")
		buildSpan.RecordError(err)
		return nil, err
	}
	repoCtx, repoSpan := s.tr.Start(ctx, "artist.update.repository_update")
	defer repoSpan.End()

	updatedArtist, err := s.r.UpdateByID(repoCtx, id, update)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			repoSpan.RecordError(err)
			return nil, ErrArtistNotFound
		}
		repoSpan.RecordError(err)
		return nil, err
	}

	eventCtx, eventSpan := s.tr.Start(ctx, "artist.update.update_event")
	defer eventSpan.End()

	aep := toArtistUpdatedEvent(updatedArtist.ID.Hex(), updatedArtist.Name)

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

// ArtistsQuery represents the query parameters for filtering and paginating artist results.
type ArtistsQuery struct {
	Page  int
	Size  int
	Name  string
	Genre string
}

// GetArtists retrieves a paginated list of artists with optional filtering by name or genre.
func (s *ArtistService) GetArtists(ctx context.Context, q ArtistsQuery) (*dtos.ArtistListResponseDto, error) {
	ctx, span := s.tr.Start(ctx, "artists.get_all")
	defer span.End()

	filter := bson.M{}
	if q.Name != "" {
		filter["name"] = bson.M{
			"$regex":   q.Name,
			"$options": "i",
		}
	}

	if q.Genre != "" {
		filter["genres.name"] = bson.M{
			"$regex":   q.Genre,
			"$options": "i",
		}
	}

	p := pagination.NewPagination(q.Page, q.Size)
	items, total, err := s.r.FindAll(ctx, filter, p.Skip(), p.Limit())
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return &dtos.ArtistListResponseDto{
		Items: items,
		Page:  p.Page,
		Size:  p.Size,
		Total: total,
	}, nil
}

// Exists checks if an artist with the given ID exists.
func (s *ArtistService) Exists(ctx context.Context, artistIDstr string) (bool, error) {
	ctx, span := s.tr.Start(ctx, "artists.exists")
	defer span.End()

	artistID, err := primitive.ObjectIDFromHex(artistIDstr)
	if err != nil {
		span.RecordError(err)
		return false, err
	}

	existsCtx, existsSpan := s.tr.Start(ctx, "artists.exists.existence_check")
	defer existsSpan.End()

	exists, err := s.r.Exists(existsCtx, artistID)
	if err != nil {
		existsSpan.RecordError(err)
		return false, err
	}

	return exists, nil
}

func toArtistCreatedEvent(genreIDs []string, artistID string, artistName string) *events.EntityCreatedEventPayload {
	return &events.EntityCreatedEventPayload{
		TargetIDs:  genreIDs,
		EntityID:   artistID,
		EntityName: artistName,
		CreatedAt:  time.Now(),
		EntityType: events.ArtistType,
		EventID:    primitive.NewObjectID().Hex(),
	}
}

func toArtistUpdatedEvent(artistID string, artistName string) *events.EntityUpdatedEventPayload {
	return &events.EntityUpdatedEventPayload{
		EntityID:   artistID,
		EntityName: artistName,
	}
}
