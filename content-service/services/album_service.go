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

var ErrAlbumNotFound = errors.New("album not found")

type AlbumService struct {
	albumRepo      *repositories.AlbumRepository
	artistService  *ArtistService
	songRepository *repositories.SongRepository
	genreService   *GenreService
	jsc            *events.JetStreamClient
	tr             trace.Tracer
}

// NewAlbumService creates and returns a new AlbumService with the provided repository and dependent services.
func NewAlbumService(
	r repositories.AlbumRepository,
	artistService ArtistService,
	songRepository repositories.SongRepository,
	genreService GenreService,
	jsc events.JetStreamClient,
) *AlbumService {
	tr := otel.Tracer("content-service/album-service")
	s := AlbumService{albumRepo: &r, artistService: &artistService, songRepository: &songRepository, genreService: &genreService, jsc: &jsc, tr: tr}

	return &s
}

// Create creates a new album with the provided data, resolving associated artists and songs.
func (s *AlbumService) Create(ctx context.Context, albumDto *dtos.CreateAlbumDto) error {
	createCtx, createSpan := s.tr.Start(ctx, "album.create")
	defer createSpan.End()

	resolveArtistCtx, resolveArtistSpan := s.tr.Start(createCtx, "album.create.resolve_artist")
	defer resolveArtistSpan.End()

	embeddedArtist := make([]entities.Artist, 0)
	artistIDs := []string{}

	for _, artistsIdStr := range albumDto.ArtistIds {
		artist, err := s.artistService.FindArtistByID(resolveArtistCtx, artistsIdStr)
		if err != nil {
			resolveArtistSpan.RecordError(err)

			switch {
			case errors.Is(err, ErrObjectIdCastFailed):
				return ErrObjectIdCastFailed
			case errors.Is(err, ErrArtistNotFound):
				return ErrArtistNotFound
			default:
				return err
			}
		}

		embeddedArtist = append(embeddedArtist, entities.Artist{
			ID:          artist.ID,
			Name:        artist.Name,
			Genres:      artist.Genres,
			Description: artist.Description,
		})

		artistIDs = append(artistIDs, artist.ID.Hex())
	}

	resolveGenreCtx, resolveGenreSpan := s.tr.Start(createCtx, "album.create.resolve_genre")
	defer resolveGenreSpan.End()

	embeddedGenre := make([]entities.Genre, 0)

	for _, genreIdStr := range albumDto.GenreIds {
		genre, err := s.genreService.FindGenreByID(resolveGenreCtx, genreIdStr)
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
	}

	createAlCtx, createAlSpan := s.tr.Start(createCtx, "album.create.create_album")
	defer createAlSpan.End()

	albumEntity, err := mappers.ToAlbumEntity(albumDto, embeddedArtist, embeddedGenre)
	if err != nil {
		createAlSpan.RecordError(err)
		logging.Errorf(ctx, "error converting to album entity: %v", err)
		return err
	}

	err = s.albumRepo.Create(createAlCtx, *albumEntity)
	if err != nil {
		createAlSpan.RecordError(err)
		logging.Errorf(ctx, "error creating album in database: %v", err)
		return err
	}

	timeoutCtx, cancel := context.WithTimeout(createCtx, 5*time.Second)
	defer cancel()

	eventCtx, eventSpan := s.tr.Start(timeoutCtx, "album.create.event")
	defer eventSpan.End()

	aep := toAlbumCreatedEvent(artistIDs, albumEntity.ID.Hex(), albumEntity.Title)

	err = retry.Do(
		func() error {
			return s.jsc.Publish(eventCtx, events.SUBJECT_ENTITY_CREATED, aep)
		},
		retry.Attempts(3),
		retry.Delay(time.Second),
		retry.DelayType(retry.BackOffDelay),
		retry.Context(eventCtx),
	)
	if err != nil {
		logging.Errorf(eventCtx, "failed to publish entity created event: %v", err)
	}

	return nil
}

// FindAlbumByID retrieves a single album by its ID.
func (s *AlbumService) FindAlbumByID(ctx context.Context, idStr string) (*entities.Album, error) {
	ctx, span := s.tr.Start(ctx, "album.find_by_id")
	defer span.End()

	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		span.RecordError(err)
		return nil, ErrObjectIdCastFailed
	}

	album, err := s.albumRepo.FindByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		return nil, ErrAlbumNotFound
	}

	return album, nil
}

// UpdateAlbum updates an existing album with the provided partial data.
func (s *AlbumService) UpdateAlbum(ctx context.Context, idStr string, dto dtos.UpdateAlbumDto) (*entities.Album, error) {
	ctx, span := s.tr.Start(ctx, "album.update_album")
	defer span.End()

	_, parseSpan := s.tr.Start(ctx, "album.update_album.parse_id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		parseSpan.RecordError(err)
		parseSpan.End()
		return nil, ErrObjectIdCastFailed
	}
	parseSpan.End()

	_, buildSpan := s.tr.Start(ctx, "album.update.build_update_doc")
	update := make(map[string]any)

	if dto.Title != nil {
		update["title"] = *dto.Title
	}
	if dto.ReleaseDate != nil {
		update["release_date"] = *dto.ReleaseDate
	}
	if dto.GenreIds != nil {
		embeddedGenre := make([]entities.Genre, 0)
		for _, genreIdStr := range *dto.GenreIds {
			genre, err := s.genreService.FindGenreByID(ctx, genreIdStr)
			if err != nil {
				buildSpan.RecordError(err)
				buildSpan.End()

				switch {
				case errors.Is(err, ErrObjectIdCastFailed):
					return nil, ErrObjectIdCastFailed
				case errors.Is(err, ErrGenreNotFound):
					return nil, ErrGenreNotFound
				default:
					return nil, err
				}
			}

			embeddedGenre = append(embeddedGenre, entities.Genre{
				ID:   genre.ID,
				Name: genre.Name,
			})
		}
		update["genres"] = embeddedGenre
	}
	if dto.ArtistIds != nil {
		embeddedArtist := make([]entities.Artist, 0)
		for _, artistsIdStr := range *dto.ArtistIds {
			artist, err := s.artistService.FindArtistByID(ctx, artistsIdStr)
			if err != nil {
				buildSpan.RecordError(err)
				buildSpan.End()

				switch {
				case errors.Is(err, ErrObjectIdCastFailed):
					return nil, ErrObjectIdCastFailed
				case errors.Is(err, ErrArtistNotFound):
					return nil, ErrArtistNotFound
				default:
					return nil, err
				}
			}

			embeddedArtist = append(embeddedArtist, entities.Artist{
				ID:          artist.ID,
				Name:        artist.Name,
				Genres:      artist.Genres,
				Description: artist.Description,
			})
		}
		update["artists"] = embeddedArtist
	}

	if len(update) == 0 {
		err := errors.New("no fields to update")
		buildSpan.RecordError(err)
		buildSpan.End()
		return nil, err
	}
	buildSpan.End()

	repoCtx, repoSpan := s.tr.Start(ctx, "album.update.repository_update")
	updatedAlbum, err := s.albumRepo.UpdateByID(repoCtx, id, update)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			repoSpan.RecordError(err)
			repoSpan.End()
			return nil, ErrAlbumNotFound
		}
		repoSpan.RecordError(err)
		repoSpan.End()
		return nil, err
	}
	repoSpan.End()

	return updatedAlbum, nil
}

// AddSongsToAlbum appends songs to an album by resolving song IDs.
func (s *AlbumService) AddSongsToAlbum(ctx context.Context, idStr string, dto dtos.AddAlbumSongsDto) (*entities.Album, error) {
	ctx, span := s.tr.Start(ctx, "album.add_songs")
	defer span.End()

	_, parseSpan := s.tr.Start(ctx, "album.add_songs.parse_id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		parseSpan.RecordError(err)
		parseSpan.End()
		return nil, ErrObjectIdCastFailed
	}
	parseSpan.End()

	album, err := s.albumRepo.FindByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		return nil, ErrAlbumNotFound
	}

	existing := make(map[primitive.ObjectID]struct{}, len(album.Songs))
	for _, song := range album.Songs {
		existing[song.ID] = struct{}{}
	}

	embeddedSong := make([]entities.Song, 0, len(dto.Ids))
	for _, songsIdStr := range dto.Ids {
		songId, err := primitive.ObjectIDFromHex(songsIdStr)
		if err != nil {
			span.RecordError(err)
			return nil, ErrObjectIdCastFailed
		}

		song, err := s.songRepository.FindByID(ctx, songId)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				span.RecordError(err)
				return nil, ErrSongNotFound
			}
			span.RecordError(err)
			return nil, err
		}

		if _, ok := existing[song.ID]; ok {
			continue
		}
		embeddedSong = append(embeddedSong, entities.Song{
			ID:            song.ID,
			Title:         song.Title,
			Genres:        song.Genres,
			LengthSeconds: song.LengthSeconds,
			Artists:       song.Artists,
		})
		existing[song.ID] = struct{}{}
	}

	if len(embeddedSong) == 0 {
		return album, nil
	}

	album.Songs = append(album.Songs, embeddedSong...)
	updatedAlbum, err := s.albumRepo.UpdateByID(ctx, id, map[string]any{"songs": album.Songs})
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return updatedAlbum, nil
}

// GetAlbumSongs retrieves songs for an album by ID.
func (s *AlbumService) GetAlbumSongs(ctx context.Context, idStr string) ([]entities.Song, error) {
	ctx, span := s.tr.Start(ctx, "album.get_songs")
	defer span.End()

	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		span.RecordError(err)
		return nil, ErrObjectIdCastFailed
	}

	album, err := s.albumRepo.FindByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		return nil, ErrAlbumNotFound
	}

	return album.Songs, nil
}

// RemoveSongFromAlbum removes a song from an album by ID.
func (s *AlbumService) RemoveSongFromAlbum(ctx context.Context, albumIdStr string, songIdStr string) error {
	ctx, span := s.tr.Start(ctx, "album.remove_song")
	defer span.End()

	albumId, err := primitive.ObjectIDFromHex(albumIdStr)
	if err != nil {
		span.RecordError(err)
		return ErrObjectIdCastFailed
	}

	songId, err := primitive.ObjectIDFromHex(songIdStr)
	if err != nil {
		span.RecordError(err)
		return ErrObjectIdCastFailed
	}

	album, err := s.albumRepo.FindByID(ctx, albumId)
	if err != nil {
		span.RecordError(err)
		return ErrAlbumNotFound
	}

	filtered := make([]entities.Song, 0, len(album.Songs))
	removed := false
	for _, song := range album.Songs {
		if song.ID == songId {
			removed = true
			continue
		}
		filtered = append(filtered, song)
	}

	if !removed {
		return ErrSongNotFound
	}

	_, err = s.albumRepo.UpdateByID(ctx, albumId, map[string]any{"songs": filtered})
	if err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}

// RemoveSongFromAllAlbums removes a song from every album that embeds it.
func (s *AlbumService) RemoveSongFromAllAlbums(ctx context.Context, songIdStr string) error {
	ctx, span := s.tr.Start(ctx, "album.remove_song_all")
	defer span.End()

	songId, err := primitive.ObjectIDFromHex(songIdStr)
	if err != nil {
		span.RecordError(err)
		return ErrObjectIdCastFailed
	}

	filter := bson.M{"songs._id": songId}
	albums, _, err := s.albumRepo.FindAll(ctx, filter, 0, 0)
	if err != nil {
		span.RecordError(err)
		return err
	}

	for _, album := range albums {
		filtered := make([]entities.Song, 0, len(album.Songs))
		for _, song := range album.Songs {
			if song.ID == songId {
				continue
			}
			filtered = append(filtered, song)
		}

		if _, err := s.albumRepo.UpdateByID(ctx, album.ID, map[string]any{"songs": filtered}); err != nil {
			span.RecordError(err)
			return err
		}
	}

	return nil
}

// DeleteAlbum deletes an album by its ID.
func (s *AlbumService) DeleteAlbum(ctx context.Context, idStr string) error {
	ctx, span := s.tr.Start(ctx, "album.delete_album")
	defer span.End()

	_, parseSpan := s.tr.Start(ctx, "album.delete_album.parse_id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		parseSpan.RecordError(err)
		parseSpan.End()
		return ErrObjectIdCastFailed
	}
	parseSpan.End()

	repoCtx, repoSpan := s.tr.Start(ctx, "album.delete.repository_delete")
	res, err := s.albumRepo.DeleteByID(repoCtx, id)
	if err != nil {
		repoSpan.RecordError(err)
		repoSpan.End()
		return err
	}

	if res.DeletedCount == 0 {
		err = ErrAlbumNotFound
		repoSpan.RecordError(err)
		repoSpan.End()
		return err
	}
	repoSpan.End()

	return nil
}

// AlbumsQuery represents the query parameters for filtering and paginating album results.
type AlbumsQuery struct {
	Page     int
	Size     int
	Title    string
	Genres   string
	GenreID  string
	ArtistId string
}

// GetAlbums retrieves a paginated list of albums with optional filtering by title, genre, or artist ID.
func (s *AlbumService) GetAlbums(ctx context.Context, q AlbumsQuery) (*dtos.AlbumListResponseDto, error) {
	ctx, span := s.tr.Start(ctx, "album.get_all")
	defer span.End()

	filter := bson.M{}
	if q.Title != "" {
		filter["title"] = bson.M{
			"$regex":   q.Title,
			"$options": "i",
		}
	}

	if q.Genres != "" {
		filter["genres"] = bson.M{
			"$elemMatch": bson.M{
				"name": bson.M{
					"$regex":   q.Genres,
					"$options": "i",
				},
			},
		}
	}

	if q.GenreID != "" {
		genreId, err := primitive.ObjectIDFromHex(q.GenreID)
		if err != nil {
			return nil, ErrObjectIdCastFailed
		}
		filter["genres._id"] = genreId
	}

	if q.ArtistId != "" {
		artistId, err := primitive.ObjectIDFromHex(q.ArtistId)
		if err != nil {
			return nil, ErrObjectIdCastFailed
		}
		filter["artists._id"] = artistId
	}

	p := pagination.NewPagination(q.Page, q.Size)
	resp, err := commondtos.ListWithPagination(ctx, p, filter, s.albumRepo.FindAll)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return resp, nil
}

func toAlbumCreatedEvent(artistIDs []string, albumID string, albumName string) *events.EntityCreatedEventPayload {
	return &events.EntityCreatedEventPayload{
		TargetIDs:  artistIDs,
		EntityID:   albumID,
		EntityName: albumName,
		CreatedAt:  time.Now(),
		EntityType: events.AlbumType,
		EventID:    primitive.NewObjectID().Hex(),
	}
}
