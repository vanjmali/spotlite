package services

import (
	"context"
	"errors"
	"io"
	"log"

	"github.com/vanjmali/spotlite/common-lib/pagination"
	"github.com/vanjmali/spotlite/common-lib/telemetry"
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

type SongService struct {
	songRepo      *repositories.SongRepository
	artistService *ArtistService
	genreService  *GenreService
	albumService  *AlbumService
	hdfs          *storage.HDFSStorage
	tr            trace.Tracer
}

// NewSongService creates and returns a new SongService with the provided repository and artist service.
func NewSongService(songRepo repositories.SongRepository, artistService ArtistService, genreService GenreService, albumService *AlbumService, hdfs *storage.HDFSStorage) *SongService {
	tr := otel.Tracer("content-service/song-service")
	s := SongService{songRepo: &songRepo, artistService: &artistService, genreService: &genreService, albumService: albumService, hdfs: hdfs, tr: tr}

	return &s
}

// Create creates a new song with the provided data, resolving associated artists and genre.
func (s *SongService) Create(ctx context.Context, songDto *dtos.SongDto) (primitive.ObjectID, error) {
	ctx, span := s.tr.Start(ctx, "song.create")
	defer span.End()

	resolveAlbumCtx, resolveAlbumSpan := s.tr.Start(ctx, "song.create.resolve_album")
	_, err := s.albumService.FindAlbumByID(resolveAlbumCtx, songDto.AlbumId)
	if err != nil {
		resolveAlbumSpan.RecordError(err)
		resolveAlbumSpan.End()

		switch {
		case errors.Is(err, ErrObjectIdCastFailed):
			return ErrObjectIdCastFailed
		case errors.Is(err, ErrAlbumNotFound):
			return ErrAlbumNotFound
		default:
			return err
		}
	}

	resolveAlbumSpan.End()

	resolveGenreCtx, resolveGenreSpan := s.tr.Start(ctx, "song.create.resolve_genre")

	embeddedGenre := make([]entities.Genre, 0)

	for _, genreIdStr := range songDto.GenreIds {
		genre, err := s.genreService.FindGenreByID(resolveGenreCtx, genreIdStr)
		if err != nil {
			resolveGenreSpan.RecordError(err)
			resolveGenreSpan.End()

			switch {
			case errors.Is(err, ErrObjectIdCastFailed):
				return primitive.NilObjectID, ErrObjectIdCastFailed
			case errors.Is(err, ErrGenreNotFound):
				return primitive.NilObjectID, ErrGenreNotFound
			default:
				return primitive.NilObjectID, err
			}
		}

		embeddedGenre = append(embeddedGenre, entities.Genre{
			ID:   genre.ID,
			Name: genre.Name,
		})
	}

	resolveGenreSpan.End()

	resolveCtx, resolveSpan := s.tr.Start(ctx, "song.create.resolve_artists")
	embeddedArtists := make([]entities.Artist, 0)

	for _, artistIdStr := range songDto.ArtistIds {
		artist, err := s.artistService.FindArtistByID(resolveCtx, artistIdStr)
		if err != nil {
			resolveSpan.RecordError(err)
			resolveSpan.End()

			switch {
			case errors.Is(err, ErrObjectIdCastFailed):
				return primitive.NilObjectID, ErrObjectIdCastFailed
			case errors.Is(err, ErrArtistNotFound):
				return primitive.NilObjectID, ErrArtistNotFound
			default:
				return primitive.NilObjectID, err
			}
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
	songEntity, err := mappers.ToSongEntity(songDto, embeddedGenre, embeddedArtists)
	if err != nil {
		mapSpan.RecordError(err)
		mapSpan.End()
		log.Printf("trace_id=%s error converting to song entity: %v", telemetry.TraceID(ctx), err)
		return primitive.NilObjectID, err
	}
	mapSpan.End()

	id := songEntity.ID
	createCtx, createSpan := s.tr.Start(ctx, "song.create.create_song")
	_, err = s.songRepo.Create(createCtx, *songEntity)
	if err != nil {
		createSpan.RecordError(err)
		createSpan.End()
		log.Printf("trace_id=%s error creating song in database: %v", telemetry.TraceID(ctx), err)
		return primitive.NilObjectID, err
	}

	_, err = s.albumService.AddSongsToAlbum(ctx, songDto.AlbumId, dtos.AddAlbumSongsDto{Ids: []string{songEntity.ID.Hex()}})
	if err != nil {
		createSpan.RecordError(err)
		createSpan.End()
		log.Printf("trace_id=%s error embedding song into album: %v", telemetry.TraceID(ctx), err)
		return primitive.NilObjectID, err
	}
	createSpan.End()

	return id, nil
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
	if dto.LengthSeconds != nil {
		update["length_seconds"] = *dto.LengthSeconds
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
		if err := s.hdfs.Remove(song.AudioPath); err != nil {
			log.Printf("trace_id=%s failed to delete audio file at path %s: %v", telemetry.TraceID(ctx), song.AudioPath, err)

		}
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

func (s *SongService) UploadAudio(ctx context.Context, idStr string, r io.Reader, ext string, mime string) (*entities.Song, error) {
	ctx, span := s.tr.Start(ctx, "song.upload_audio")
	defer span.End()

	_, parseSpan := s.tr.Start(ctx, "song.upload_audio.parse_id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		parseSpan.RecordError(err)
		parseSpan.End()
		return nil, ErrObjectIdCastFailed
	}
	parseSpan.End()

	checkExistsCtx, checkExistsSpan := s.tr.Start(ctx, "song.upload_audio.check_exists")
	song, err := s.songRepo.FindByID(checkExistsCtx, id)
	if err != nil {
		checkExistsSpan.RecordError(err)
		checkExistsSpan.End()
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrSongNotFound
		}
		return nil, err
	}
	checkExistsSpan.End()

	_, uploadSpan := s.tr.Start(ctx, "song.upload_audio.hdfs_upload")
	audioPath, size, err := s.hdfs.UploadSongAudio(id.Hex(), r, ext)
	if err != nil {
		uploadSpan.RecordError(err)
		uploadSpan.End()
		return nil, ErrAudioUploadFailed
	}
	uploadSpan.End()

	updated, err := s.songRepo.UpdateAudioByID(ctx, id, audioPath, size, mime)
	if err != nil {
		span.RecordError(err)
		_ = s.hdfs.Remove(audioPath)
		return nil, err
	}

	if song.AudioPath != "" && song.AudioPath != audioPath {
		if err := s.hdfs.Remove(song.AudioPath); err != nil {
			log.Printf("trace_id=%s failed to delete old audio at path %s: %v",
				telemetry.TraceID(ctx), song.AudioPath, err)
		}
	}
	return updated, nil
}

func (s *SongService) OpenAudio(ctx context.Context, audioPath string) (io.ReadCloser, error) {
	return s.hdfs.Open(audioPath)
}
