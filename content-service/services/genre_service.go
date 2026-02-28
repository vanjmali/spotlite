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
	r          *repositories.GenreRepository
	artistRepo *repositories.ArtistRepository
	songRepo   *repositories.SongRepository
	albumRepo  *repositories.AlbumRepository
	jsc        *events.JetStreamClient
	tr         trace.Tracer
}

// NewArtistService builds a ArtistService with repository.
func NewGenreService(
	r repositories.GenreRepository,
	artistRepo repositories.ArtistRepository,
	songRepo repositories.SongRepository,
	albumRepo repositories.AlbumRepository,
	jsc events.JetStreamClient,
) *GenreService {
	tr := otel.Tracer("content-service/genre-service")
	s := GenreService{
		r:          &r,
		artistRepo: &artistRepo,
		songRepo:   &songRepo,
		albumRepo:  &albumRepo,
		jsc:        &jsc,
		tr:         tr,
	}

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

	// fetches current state of the genre for potential roll back
	getCtx, getSpan := s.tr.Start(updateCtx, "genre.update.get_current_state")
	defer getSpan.End()

	currentGenre, err := s.r.FindByID(getCtx, id)
	if err != nil {
		switch {
		case errors.Is(err, mongo.ErrNoDocuments):
			return nil, ErrGenreNotFound
		default:
			return nil, err
		}
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

	eventCtx, eventSpan := s.tr.Start(updateCtx, "genre.update.update_event")
	defer eventSpan.End()

	// prepare payload
	aep := toGenreUpdatedEvent(genre.ID.Hex(), genre.Name)

	// attempts broadcasting event
	err = retry.Do(
		func() error {
			return s.jsc.Publish(eventCtx, events.SUBJECT_ENTITY_UPDATED, aep)
		},
		retry.Attempts(3),
		retry.Delay(time.Second),
		retry.DelayType(retry.BackOffDelay),
		retry.Context(eventCtx),
	)
	// if event couldn't be published rollback to previous genre state
	if err != nil {
		logging.Errorf(eventCtx, "failed to publish entity updated event: %v", err)
		eventSpan.RecordError(err)

		var errs []error

		errs = append(errs, err)

		rbCtx, rbSpan := s.tr.Start(updateCtx, "genre.update.rollback")
		defer rbSpan.End()

		// set param to previous genre state name
		rbUpdate := make(map[string]any)
		if dto.Name != nil {
			rbUpdate["name"] = currentGenre.Name
		}

		_, err := s.r.UpdateByID(rbCtx, id, rbUpdate)
		if err != nil {
			rbSpan.RecordError(err)
			errs = append(errs, err)
		}

		return nil, errors.Join(errs...)
	}

	syncCtx, syncSpan := s.tr.Start(updateCtx, "genre.update.sync_embeds")
	if err := s.syncEmbeddedReferences(syncCtx, genre); err != nil {
		syncSpan.RecordError(err)
		syncSpan.End()
		return nil, err
	}
	syncSpan.End()

	return genre, nil
}

func (s *GenreService) syncEmbeddedReferences(ctx context.Context, genre *entities.Genre) error {
	// Sync embedded genre snapshot in artists.
	artists, _, err := s.artistRepo.FindAll(ctx, bson.M{"genres._id": genre.ID}, 0, 0)
	if err != nil {
		return err
	}

	for _, artist := range artists {
		changed := false
		for i := range artist.Genres {
			if artist.Genres[i].ID != genre.ID {
				continue
			}
			if artist.Genres[i].Name != genre.Name {
				artist.Genres[i].Name = genre.Name
				changed = true
			}
		}
		if !changed {
			continue
		}
		if _, err := s.artistRepo.UpdateByID(ctx, artist.ID, map[string]any{"genres": artist.Genres}); err != nil {
			return err
		}
	}

	// Sync embedded genre snapshot in songs, including song.artists[].genres.
	songFilter := bson.M{
		"$or": []bson.M{
			{"genres._id": genre.ID},
			{"artists.genres._id": genre.ID},
		},
	}
	songs, _, err := s.songRepo.FindAll(ctx, songFilter, 0, 0)
	if err != nil {
		return err
	}

	for _, song := range songs {
		genresChanged := false
		for i := range song.Genres {
			if song.Genres[i].ID != genre.ID {
				continue
			}
			if song.Genres[i].Name != genre.Name {
				song.Genres[i].Name = genre.Name
				genresChanged = true
			}
		}

		artistsChanged := false
		for i := range song.Artists {
			for j := range song.Artists[i].Genres {
				if song.Artists[i].Genres[j].ID != genre.ID {
					continue
				}
				if song.Artists[i].Genres[j].Name != genre.Name {
					song.Artists[i].Genres[j].Name = genre.Name
					artistsChanged = true
				}
			}
		}

		if !genresChanged && !artistsChanged {
			continue
		}

		update := map[string]any{}
		if genresChanged {
			update["genres"] = song.Genres
		}
		if artistsChanged {
			update["artists"] = song.Artists
		}

		if _, err := s.songRepo.UpdateByID(ctx, song.ID, update); err != nil {
			return err
		}
	}

	// Sync embedded genre snapshot in albums, including nested artists/songs.
	albumFilter := bson.M{
		"$or": []bson.M{
			{"genres._id": genre.ID},
			{"artists.genres._id": genre.ID},
			{"songs.genres._id": genre.ID},
			{"songs.artists.genres._id": genre.ID},
		},
	}
	albums, _, err := s.albumRepo.FindAll(ctx, albumFilter, 0, 0)
	if err != nil {
		return err
	}

	for _, album := range albums {
		genresChanged := false
		for i := range album.Genres {
			if album.Genres[i].ID != genre.ID {
				continue
			}
			if album.Genres[i].Name != genre.Name {
				album.Genres[i].Name = genre.Name
				genresChanged = true
			}
		}

		artistsChanged := false
		for i := range album.Artists {
			for j := range album.Artists[i].Genres {
				if album.Artists[i].Genres[j].ID != genre.ID {
					continue
				}
				if album.Artists[i].Genres[j].Name != genre.Name {
					album.Artists[i].Genres[j].Name = genre.Name
					artistsChanged = true
				}
			}
		}

		songsChanged := false
		for i := range album.Songs {
			for j := range album.Songs[i].Genres {
				if album.Songs[i].Genres[j].ID != genre.ID {
					continue
				}
				if album.Songs[i].Genres[j].Name != genre.Name {
					album.Songs[i].Genres[j].Name = genre.Name
					songsChanged = true
				}
			}
			for j := range album.Songs[i].Artists {
				for k := range album.Songs[i].Artists[j].Genres {
					if album.Songs[i].Artists[j].Genres[k].ID != genre.ID {
						continue
					}
					if album.Songs[i].Artists[j].Genres[k].Name != genre.Name {
						album.Songs[i].Artists[j].Genres[k].Name = genre.Name
						songsChanged = true
					}
				}
			}
		}

		if !genresChanged && !artistsChanged && !songsChanged {
			continue
		}

		update := map[string]any{}
		if genresChanged {
			update["genres"] = album.Genres
		}
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
