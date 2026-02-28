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
	songRepo     *repositories.SongRepository
	albumRepo    *repositories.AlbumRepository
	genreService *GenreService
	jsc          *events.JetStreamClient
	tr           trace.Tracer
}

// NewArtistService builds a ArtistService with repository.
func NewArtistService(
	r repositories.ArtistRepository,
	songRepo repositories.SongRepository,
	albumRepo repositories.AlbumRepository,
	genreService GenreService,
	jsc events.JetStreamClient,
) *ArtistService {
	tr := otel.Tracer("content-service/artist-service")
	s := ArtistService{
		r:            &r,
		songRepo:     &songRepo,
		albumRepo:    &albumRepo,
		genreService: &genreService,
		tr:           tr,
		jsc:          &jsc,
	}
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
	updateCtx, updateSpan := s.tr.Start(ctx, "artist.update")
	defer updateSpan.End()

	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		updateSpan.RecordError(err)
		return nil, ErrObjectIdCastFailed
	}

	// fetches current state of the genre for potential roll back
	getCtx, getSpan := s.tr.Start(updateCtx, "artist.update.get_current_state")
	defer getSpan.End()

	currentArtist, err := s.r.FindByID(getCtx, id)
	if err != nil {
		getSpan.RecordError(err)
		switch {
		case errors.Is(err, mongo.ErrNoDocuments):
			updateSpan.RecordError(err)
			return nil, ErrArtistNotFound
		default:
			updateSpan.RecordError(err)
			return nil, err
		}
	}

	update := make(map[string]any)

	if dto.Name != nil {
		update["name"] = *dto.Name
	}
	if dto.GenreIds != nil {
		embeddedGenres := make([]entities.Genre, 0)
		for _, genreIdStr := range *dto.GenreIds {
			genre, err := s.genreService.FindGenreByID(updateCtx, genreIdStr)
			if err != nil {
				updateSpan.RecordError(err)
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
		updateSpan.RecordError(err)
		return nil, err
	}
	repoCtx, repoSpan := s.tr.Start(updateCtx, "artist.update.repo")
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

	eventCtx, eventSpan := s.tr.Start(updateCtx, "artist.update.update_event")
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

		var errs []error

		errs = append(errs, err)

		rbCtx, rbSpan := s.tr.Start(updateCtx, "artist.update.rollback")
		defer rbSpan.End()

		// set param to previous genre state name
		rbUpdate := make(map[string]any)

		if dto.Name != nil {
			rbUpdate["name"] = currentArtist.Name
		}

		if dto.GenreIds != nil {
			rbUpdate["genres"] = currentArtist.Genres
		}

		if dto.Description != nil {
			rbUpdate["description"] = currentArtist.Description
		}

		_, err := s.r.UpdateByID(rbCtx, id, rbUpdate)
		if err != nil {
			rbSpan.RecordError(err)
			errs = append(errs, err)
		}

		return nil, errors.Join(errs...)
	}

	syncCtx, syncSpan := s.tr.Start(updateCtx, "artist.update.sync_embeds")
	if err := s.syncEmbeddedReferences(syncCtx, updatedArtist); err != nil {
		syncSpan.RecordError(err)
		syncSpan.End()
		return nil, err
	}
	syncSpan.End()

	return updatedArtist, nil
}

func (s *ArtistService) syncEmbeddedReferences(ctx context.Context, artist *entities.Artist) error {
	// Sync embedded artist snapshot in songs.
	songs, _, err := s.songRepo.FindAll(ctx, bson.M{"artists._id": artist.ID}, 0, 0)
	if err != nil {
		return err
	}

	for _, song := range songs {
		changed := false
		for i := range song.Artists {
			if song.Artists[i].ID != artist.ID {
				continue
			}
			song.Artists[i] = *artist
			changed = true
		}

		if !changed {
			continue
		}

		if _, err := s.songRepo.UpdateByID(ctx, song.ID, map[string]any{"artists": song.Artists}); err != nil {
			return err
		}
	}

	// Sync embedded artist snapshot in albums (top-level artists and song artists).
	filter := bson.M{
		"$or": []bson.M{
			{"artists._id": artist.ID},
			{"songs.artists._id": artist.ID},
		},
	}

	albums, _, err := s.albumRepo.FindAll(ctx, filter, 0, 0)
	if err != nil {
		return err
	}

	for _, album := range albums {
		artistsChanged := false
		for i := range album.Artists {
			if album.Artists[i].ID != artist.ID {
				continue
			}
			album.Artists[i] = *artist
			artistsChanged = true
		}

		songsChanged := false
		for i := range album.Songs {
			for j := range album.Songs[i].Artists {
				if album.Songs[i].Artists[j].ID != artist.ID {
					continue
				}
				album.Songs[i].Artists[j] = *artist
				songsChanged = true
			}
		}

		if !artistsChanged && !songsChanged {
			continue
		}

		update := map[string]any{}
		if artistsChanged {
			update["artists"] = album.Artists
		}
		if songsChanged {
			update["songs"] = album.Songs
		}

		if _, err := s.albumRepo.UpdateByID(ctx, album.ID, update); err != nil {
			return err
		}
	}

	return nil
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
