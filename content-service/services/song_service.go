package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"time"

	"github.com/avast/retry-go"
	"github.com/vanjmali/spotlite/common-lib/events"
	"github.com/vanjmali/spotlite/common-lib/logging"
	"github.com/vanjmali/spotlite/common-lib/pagination"
	"github.com/vanjmali/spotlite/content/dtos"
	"github.com/vanjmali/spotlite/content/entities"
	"github.com/vanjmali/spotlite/content/mappers"
	"github.com/vanjmali/spotlite/content/repositories"
	"github.com/vanjmali/spotlite/content/storage"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var (
	ErrSongNotFound      = errors.New("song not found")
	ErrAudioUploadFailed = errors.New("audio upload failed")
)

type songRepo interface {
	Create(ctx context.Context, song entities.Song) (primitive.ObjectID, error)
	FindByID(ctx context.Context, id primitive.ObjectID) (*entities.Song, error)
	UpdateByID(ctx context.Context, id primitive.ObjectID, update map[string]any) (*entities.Song, error)
	DeleteByID(ctx context.Context, id primitive.ObjectID) (*mongo.DeleteResult, error)
	FindAll(ctx context.Context, filter bson.M, skip int64, limit int64) ([]entities.Song, int64, error)
	UpdateAudioByID(
		ctx context.Context,
		id primitive.ObjectID,
		audioPath string,
		size int64,
		mime string,
		checksum string,
		lengthSeconds *int,
	) (*entities.Song, error)
}

type artistFinder interface {
	FindArtistByID(ctx context.Context, idStr string) (*entities.Artist, error)
}

type genreFinder interface {
	FindGenreByID(ctx context.Context, idStr string) (*entities.Genre, error)
}

type albumSongManager interface {
	FindAlbumByID(ctx context.Context, idStr string) (*entities.Album, error)
	AddSongsToAlbum(ctx context.Context, idStr string, dto dtos.AddAlbumSongsDto) (*entities.Album, error)
	RemoveSongFromAllAlbums(ctx context.Context, songIdStr string) error
}

type audioStore interface {
	UploadSongAudio(songID string, r io.Reader, ext string) (finalPath string, size int64, err error)
	Open(p string) (io.ReadCloser, error)
	Remove(path string) error
}

type SongService struct {
	songRepo      songRepo
	artistService artistFinder
	genreService  genreFinder
	albumService  albumSongManager
	hdfs          audioStore
	jsc           *events.JetStreamClient
	tr            trace.Tracer
}

// NewSongService creates and returns a new SongService with the provided repository and artist service.
func NewSongService(
	songRepo repositories.SongRepository,
	artistService ArtistService,
	genreService GenreService,
	albumService *AlbumService,
	hdfs *storage.HDFSStorage,
	jsc *events.JetStreamClient,
) *SongService {
	tr := otel.Tracer("content-service/song-service")
	s := SongService{songRepo: &songRepo, artistService: &artistService, genreService: &genreService, albumService: albumService, hdfs: hdfs, jsc: jsc, tr: tr}

	return &s
}

// Create creates a new song with the provided data, resolving associated artists and genre.
func (s *SongService) Create(ctx context.Context, songDto *dtos.SongDto) (*SongPayload, error) {
	createCtx, createSpan := s.tr.Start(ctx, "song.create")
	defer createSpan.End()

	resolveAlbumCtx, resolveAlbumSpan := s.tr.Start(createCtx, "song.create.resolve_album")
	defer resolveAlbumSpan.End()

	_, err := s.albumService.FindAlbumByID(resolveAlbumCtx, songDto.AlbumId)
	if err != nil {
		resolveAlbumSpan.RecordError(err)

		switch {
		case errors.Is(err, ErrObjectIdCastFailed):
			return nil, ErrObjectIdCastFailed
		case errors.Is(err, ErrAlbumNotFound):
			return nil, ErrAlbumNotFound
		default:
			return nil, err
		}
	}

	resolveGenreCtx, resolveGenreSpan := s.tr.Start(createCtx, "song.create.resolve_genre")
	defer resolveGenreSpan.End()

	embeddedGenre := make([]entities.Genre, 0)

	for _, genreIdStr := range songDto.GenreIds {
		genre, err := s.genreService.FindGenreByID(resolveGenreCtx, genreIdStr)
		if err != nil {
			resolveGenreSpan.RecordError(err)

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

	resolveCtx, resolveSpan := s.tr.Start(createCtx, "song.create.resolve_artists")
	defer resolveSpan.End()

	embeddedArtists := make([]entities.Artist, 0)

	for _, artistIdStr := range songDto.ArtistIds {
		artist, err := s.artistService.FindArtistByID(resolveCtx, artistIdStr)
		if err != nil {
			resolveSpan.RecordError(err)
			resolveSpan.End()

			switch {
			case errors.Is(err, ErrObjectIdCastFailed):
				return nil, ErrObjectIdCastFailed
			case errors.Is(err, ErrArtistNotFound):
				return nil, ErrArtistNotFound
			default:
				return nil, err
			}
		}

		embeddedArtists = append(embeddedArtists, entities.Artist{
			ID:          artist.ID,
			Name:        artist.Name,
			Genres:      artist.Genres,
			Description: artist.Description,
		})
	}

	_, mapSpan := s.tr.Start(createCtx, "song.create.map_entity")
	defer mapSpan.End()
	songEntity, err := mappers.ToSongEntity(songDto, embeddedGenre, embeddedArtists)
	if err != nil {
		mapSpan.RecordError(err)
		logging.Errorf(createCtx, "error converting to song entity: %v", err)
		return nil, err
	}

	repoCtx, repoSpan := s.tr.Start(createCtx, "song.create.create_song")
	defer repoSpan.End()

	id, err := s.songRepo.Create(repoCtx, *songEntity)
	if err != nil {
		createSpan.RecordError(err)
		logging.Errorf(repoCtx, "error creating song in database: %v", err)
		return nil, err
	}

	addToAlbumCtx, addToAlbumSpan := s.tr.Start(createCtx, "song.create.add_to_album")
	defer addToAlbumSpan.End()

	_, err = s.albumService.AddSongsToAlbum(addToAlbumCtx, songDto.AlbumId, dtos.AddAlbumSongsDto{Ids: []string{id.Hex()}})
	if err != nil {
		addToAlbumSpan.RecordError(err)
		logging.Errorf(addToAlbumCtx, "error embedding song into album: %v", err)

		rollbackCtx, rollbackSpan := s.tr.Start(createCtx, "song.create.rollback")
		defer rollbackSpan.End()

		if deleteErr := s.DeleteSong(rollbackCtx, id.Hex()); deleteErr != nil {
			rollbackSpan.RecordError(deleteErr)
			logging.Errorf(rollbackCtx, "critical: failed to rollback song creation for song_id=%s: %v", id.Hex(), deleteErr)
			return nil, err
		}
	}

	return &SongPayload{SongID: id.Hex(), Title: songEntity.Title, Duration: songEntity.LengthSeconds, GenreIDs: songDto.GenreIds}, nil
}

// FindSongById retrieves a single song by its ID.
func (s *SongService) FindSongById(ctx context.Context, idStr string) (*entities.Song, error) {
	ctx, span := s.tr.Start(ctx, "song.find_by_id")
	defer span.End()

	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		span.RecordError(err)
		return nil, ErrObjectIdCastFailed
	}

	song, err := s.songRepo.FindByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		return nil, ErrSongNotFound
	}

	return song, nil
}

// UpdateSong updates an existing song with the provided partial data.
func (s *SongService) UpdateSong(ctx context.Context, idStr string, dto dtos.UpdateSongDto) (*entities.Song, error) {
	ctx, span := s.tr.Start(ctx, "song.update_song")
	defer span.End()

	_, parseSpan := s.tr.Start(ctx, "song.update_song.parse_id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		parseSpan.RecordError(err)
		parseSpan.End()
		return nil, ErrObjectIdCastFailed
	}
	parseSpan.End()

	_, buildSpan := s.tr.Start(ctx, "song.update.build_update_doc")
	update := make(map[string]any)

	if dto.Title != nil {
		update["title"] = *dto.Title
	}
	if dto.GenreIds != nil {
		embeddedGenres := make([]entities.Genre, 0)
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

			embeddedGenres = append(embeddedGenres, entities.Genre{
				ID:   genre.ID,
				Name: genre.Name,
			})
		}
		update["genres"] = embeddedGenres
	}
	if dto.ArtistIds != nil {
		embeddedArtists := make([]entities.Artist, 0)
		for _, artistIdStr := range *dto.ArtistIds {
			artist, err := s.artistService.FindArtistByID(ctx, artistIdStr)
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

			embeddedArtists = append(embeddedArtists, entities.Artist{
				ID:          artist.ID,
				Name:        artist.Name,
				Genres:      artist.Genres,
				Description: artist.Description,
			})
		}
		update["artists"] = embeddedArtists
	}

	if len(update) == 0 {
		err := errors.New("no fields to update")
		buildSpan.RecordError(err)
		buildSpan.End()
		return nil, err
	}
	buildSpan.End()

	repoCtx, repoSpan := s.tr.Start(ctx, "song.update.repository_update")
	updatedSong, err := s.songRepo.UpdateByID(repoCtx, id, update)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			repoSpan.RecordError(err)
			repoSpan.End()
			return nil, ErrSongNotFound
		}
		repoSpan.RecordError(err)
		repoSpan.End()
		return nil, err
	}
	repoSpan.End()

	return updatedSong, nil
}

// DeleteSong deletes a song by its ID.
func (s *SongService) DeleteSong(ctx context.Context, idStr string) error {
	ctx, span := s.tr.Start(ctx, "song.delete_song")
	defer span.End()

	_, parseSpan := s.tr.Start(ctx, "song.delete_song.parse_id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		parseSpan.RecordError(err)
		parseSpan.End()
		return ErrObjectIdCastFailed
	}
	parseSpan.End()

	findSongCtx, findSongSpan := s.tr.Start(ctx, "song.delete_song.find_song")
	song, err := s.songRepo.FindByID(findSongCtx, id)
	if err != nil {
		findSongSpan.RecordError(err)
		findSongSpan.End()
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ErrSongNotFound
		}
		return err
	}
	findSongSpan.End()

	removeFromAlbumsCtx, removeFromAlbumsSpan := s.tr.Start(ctx, "song.delete_song.remove_from_albums")
	if err := s.albumService.RemoveSongFromAllAlbums(removeFromAlbumsCtx, id.Hex()); err != nil {
		removeFromAlbumsSpan.RecordError(err)
		removeFromAlbumsSpan.End()
		return err
	}
	removeFromAlbumsSpan.End()

	repoCtx, repoSpan := s.tr.Start(ctx, "song.delete.repository_delete")
	res, err := s.songRepo.DeleteByID(repoCtx, id)
	if err != nil {
		repoSpan.RecordError(err)
		repoSpan.End()
		return err
	}

	if res.DeletedCount == 0 {
		err = ErrSongNotFound
		repoSpan.RecordError(err)
		repoSpan.End()
		return err
	}
	repoSpan.End()

	if song.AudioPath != "" {
		cleanupCtx, cleanupSpan := s.tr.Start(ctx, "song.delete_song.remove_audio")
		if err := s.hdfs.Remove(song.AudioPath); err != nil {
			cleanupSpan.RecordError(err)
			logging.Errorf(cleanupCtx, "failed to delete audio file at path %s: %v", song.AudioPath, err)
		}
		cleanupSpan.End()
	}

	return nil
}

// SongsQuery represents the query parameters for filtering and paginating song results.
type SongsQuery struct {
	Page     int
	Size     int
	Title    string
	Genre    string
	GenreID  string
	ArtistId string
}

// GetSongs retrieves a paginated list of songs with optional filtering by title, genre, or artist ID.
func (s *SongService) GetSongs(ctx context.Context, q SongsQuery) (*dtos.SongListResponseDto, error) {
	ctx, span := s.tr.Start(ctx, "song.get_all")
	defer span.End()

	filter := bson.M{}
	if q.Title != "" {
		filter["title"] = bson.M{
			"$regex":   q.Title,
			"$options": "i",
		}
	}

	if q.Genre != "" {
		filter["genres"] = bson.M{
			"$elemMatch": bson.M{
				"name": bson.M{
					"$regex":   q.Genre,
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
	items, total, err := s.songRepo.FindAll(ctx, filter, p.Skip(), p.Limit())
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return &dtos.SongListResponseDto{
		Items: items,
		Page:  p.Page,
		Size:  p.Size,
		Total: total,
	}, nil
}

func (s *SongService) UploadAudio(ctx context.Context, p SongPayload, r io.Reader, ext string, mime string, lengthSeconds *int) (*entities.Song, error) {
	ctx, span := s.tr.Start(ctx, "song.upload_audio")
	defer span.End()

	_, parseSpan := s.tr.Start(ctx, "song.upload_audio.parse_id")
	defer parseSpan.End()

	id, err := primitive.ObjectIDFromHex(p.SongID)
	if err != nil {
		parseSpan.RecordError(err)
		return nil, ErrObjectIdCastFailed
	}

	checkExistsCtx, checkExistsSpan := s.tr.Start(ctx, "song.upload_audio.check_exists")
	defer checkExistsSpan.End()

	song, err := s.songRepo.FindByID(checkExistsCtx, id)
	if err != nil {
		checkExistsSpan.RecordError(err)
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrSongNotFound
		}
		return nil, err
	}

	_, uploadSpan := s.tr.Start(ctx, "song.upload_audio.hdfs_upload")
	defer uploadSpan.End()

	audioPath, size, checksum, err := s.uploadAudioWithChecksum(id.Hex(), r, ext)
	if err != nil {
		uploadSpan.RecordError(err)
		return nil, ErrAudioUploadFailed
	}

	updated, err := s.songRepo.UpdateAudioByID(ctx, id, audioPath, size, mime, checksum, lengthSeconds)
	if err != nil {
		span.RecordError(err)
		_ = s.hdfs.Remove(audioPath)
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrSongNotFound
		}
		return nil, err
	}

	if song.AudioPath != "" && song.AudioPath != audioPath {
		if err := s.hdfs.Remove(song.AudioPath); err != nil {
			logging.Errorf(ctx, "failed to delete old audio at path %s: %v", song.AudioPath, err)
		}
	}

	// don't send an event, we are uploading audio files which aren't needed in the recommendation service
	if p.Duration == 0 && p.Title == "" {
		return updated, nil
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	eventCtx, eventSpan := s.tr.Start(timeoutCtx, "song.upload_audio.event")
	defer eventSpan.End()

	scp := toSongCreatedEvent(id.Hex(), p.Title, p.Duration, p.GenreIDs)

	err = retry.Do(
		func() error {
			return s.jsc.Publish(eventCtx, events.SUBJECT_SONG_CREATED, scp)
		},
		retry.Attempts(3),
		retry.Delay(time.Second*1),
		retry.DelayType(retry.BackOffDelay),
		retry.Context(eventCtx),
	)
	if err != nil {
		logging.Errorf(eventCtx, "failed to publish song created event: %v", err)
		eventSpan.RecordError(err)

		var errs []error

		errs = append(errs, err)

		rbCtx, rbSpan := s.tr.Start(ctx, "song.upload_audio.rollback")
		defer rbSpan.End()

		if deleteErr := s.DeleteSong(rbCtx, id.Hex()); deleteErr != nil {
			rbSpan.RecordError(deleteErr)
			logging.Errorf(rbCtx, "critical: failed to rollback song creation for song_id=%s: %v", id.Hex(), deleteErr)
		}
		return nil, errors.Join(errs...)
	}
	return updated, nil
}

func (s *SongService) OpenAudio(ctx context.Context, audioPath string) (io.ReadCloser, error) {
	return s.hdfs.Open(audioPath)
}

func (s *SongService) uploadAudioWithChecksum(songID string, r io.Reader, ext string) (string, int64, string, error) {
	hasher := sha256.New()
	audioPath, size, err := s.hdfs.UploadSongAudio(songID, io.TeeReader(r, hasher), ext)
	if err != nil {
		return "", 0, "", err
	}

	checksum := hex.EncodeToString(hasher.Sum(nil))
	return audioPath, size, checksum, nil
}

func toSongCreatedEvent(songID string, songTitle string, duration int, genreIDs []string) *events.SongCreationPayload {
	return &events.SongCreationPayload{
		SongID:    songID,
		SongTitle: songTitle,
		Duration:  duration,
		GenreIDs:  genreIDs,
	}
}

type SongPayload struct {
	SongID   string
	Title    string
	Duration int
	GenreIDs []string
}
