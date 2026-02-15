package services

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/vanjmali/spotlite/content/dtos"
	"github.com/vanjmali/spotlite/content/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel"
)

type fakeSongRepo struct {
	findByIDFn      func(context.Context, primitive.ObjectID) (*entities.Song, error)
	updateAudioByID func(context.Context, primitive.ObjectID, string, int64, string, string, *int) (*entities.Song, error)
	deleteByIDFn    func(context.Context, primitive.ObjectID) (*mongo.DeleteResult, error)

	lastUpdateID   primitive.ObjectID
	lastUpdatePath string
	lastUpdateSize int64
	lastUpdateMime string
	lastUpdateHash string
	lastUpdateLen  *int
}

func (f *fakeSongRepo) Create(context.Context, entities.Song) (primitive.ObjectID, error) {
	return primitive.NilObjectID, errors.New("not implemented")
}

func (f *fakeSongRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*entities.Song, error) {
	if f.findByIDFn != nil {
		return f.findByIDFn(ctx, id)
	}
	return nil, mongo.ErrNoDocuments
}

func (f *fakeSongRepo) UpdateByID(context.Context, primitive.ObjectID, map[string]any) (*entities.Song, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeSongRepo) DeleteByID(ctx context.Context, id primitive.ObjectID) (*mongo.DeleteResult, error) {
	if f.deleteByIDFn != nil {
		return f.deleteByIDFn(ctx, id)
	}
	return &mongo.DeleteResult{DeletedCount: 1}, nil
}

func (f *fakeSongRepo) FindAll(context.Context, bson.M, int64, int64) ([]entities.Song, int64, error) {
	return nil, 0, errors.New("not implemented")
}

func (f *fakeSongRepo) UpdateAudioByID(
	ctx context.Context,
	id primitive.ObjectID,
	audioPath string,
	size int64,
	mime string,
	checksum string,
	lengthSeconds *int,
) (*entities.Song, error) {
	f.lastUpdateID = id
	f.lastUpdatePath = audioPath
	f.lastUpdateSize = size
	f.lastUpdateMime = mime
	f.lastUpdateHash = checksum
	f.lastUpdateLen = lengthSeconds
	if f.updateAudioByID != nil {
		return f.updateAudioByID(ctx, id, audioPath, size, mime, checksum, lengthSeconds)
	}
	song := &entities.Song{ID: id, AudioPath: audioPath, AudioSize: size, AudioMimeType: mime, AudioChecksum: checksum}
	if lengthSeconds != nil {
		song.LengthSeconds = *lengthSeconds
	}
	return song, nil
}

type fakeAlbumManager struct {
	removeFromAllFn func(context.Context, string) error
	removedSongID   string
}

func (f *fakeAlbumManager) FindAlbumByID(context.Context, string) (*entities.Album, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeAlbumManager) AddSongsToAlbum(context.Context, string, dtos.AddAlbumSongsDto) (*entities.Album, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeAlbumManager) RemoveSongFromAllAlbums(ctx context.Context, songID string) error {
	f.removedSongID = songID
	if f.removeFromAllFn != nil {
		return f.removeFromAllFn(ctx, songID)
	}
	return nil
}

type fakeAudioStore struct {
	uploadFn    func(string, io.Reader, string) (string, int64, error)
	removeFn    func(string) error
	removedPath string
}

func (f *fakeAudioStore) UploadSongAudio(songID string, r io.Reader, ext string) (string, int64, error) {
	if f.uploadFn != nil {
		return f.uploadFn(songID, r, ext)
	}
	n, _ := io.Copy(io.Discard, r)
	return "/audio/" + songID + ext, n, nil
}

func (f *fakeAudioStore) Open(string) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("")), nil
}

func (f *fakeAudioStore) Remove(path string) error {
	f.removedPath = path
	if f.removeFn != nil {
		return f.removeFn(path)
	}
	return nil
}

func newTestSongService(repo songRepo, album albumSongManager, hdfs audioStore) *SongService {
	return &SongService{
		songRepo:     repo,
		albumService: album,
		hdfs:         hdfs,
		tr:           otel.Tracer("test"),
	}
}

func TestSongServiceUploadAudio_InvalidID(t *testing.T) {
	svc := newTestSongService(&fakeSongRepo{}, &fakeAlbumManager{}, &fakeAudioStore{})

	_, err := svc.UploadAudio(context.Background(), "bad-id", strings.NewReader("x"), ".mp3", "audio/mpeg", nil)
	if !errors.Is(err, ErrObjectIdCastFailed) {
		t.Fatalf("expected ErrObjectIdCastFailed, got %v", err)
	}
}

func TestSongServiceUploadAudio_SongNotFound(t *testing.T) {
	repo := &fakeSongRepo{
		findByIDFn: func(context.Context, primitive.ObjectID) (*entities.Song, error) {
			return nil, mongo.ErrNoDocuments
		},
	}
	svc := newTestSongService(repo, &fakeAlbumManager{}, &fakeAudioStore{})

	_, err := svc.UploadAudio(context.Background(), primitive.NewObjectID().Hex(), strings.NewReader("x"), ".mp3", "audio/mpeg", nil)
	if !errors.Is(err, ErrSongNotFound) {
		t.Fatalf("expected ErrSongNotFound, got %v", err)
	}
}

func TestSongServiceUploadAudio_SuccessUpdatesMetadataAndRemovesOldFile(t *testing.T) {
	id := primitive.NewObjectID()
	oldPath := "/spotlite/audio/old.mp3"
	newPath := "/spotlite/audio/new.mp3"

	repo := &fakeSongRepo{
		findByIDFn: func(context.Context, primitive.ObjectID) (*entities.Song, error) {
			return &entities.Song{ID: id, AudioPath: oldPath}, nil
		},
		updateAudioByID: func(_ context.Context, gotID primitive.ObjectID, gotPath string, gotSize int64, gotMime string, gotHash string, gotLen *int) (*entities.Song, error) {
			song := &entities.Song{
				ID:            gotID,
				AudioPath:     gotPath,
				AudioSize:     gotSize,
				AudioMimeType: gotMime,
				AudioChecksum: gotHash,
			}
			if gotLen != nil {
				song.LengthSeconds = *gotLen
			}
			return song, nil
		},
	}
	store := &fakeAudioStore{
		uploadFn: func(gotSongID string, r io.Reader, gotExt string) (string, int64, error) {
			if gotSongID != id.Hex() {
				t.Fatalf("unexpected song id: %s", gotSongID)
			}
			if gotExt != ".mp3" {
				t.Fatalf("unexpected ext: %s", gotExt)
			}
			n, err := io.Copy(io.Discard, r)
			if err != nil {
				t.Fatalf("failed reading upload reader: %v", err)
			}
			return newPath, n, nil
		},
	}
	svc := newTestSongService(repo, &fakeAlbumManager{}, store)

	lengthSeconds := 42
	got, err := svc.UploadAudio(context.Background(), id.Hex(), strings.NewReader("song-bytes"), ".mp3", "audio/mpeg", &lengthSeconds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.AudioPath != newPath {
		t.Fatalf("expected new audio path, got %s", got.AudioPath)
	}
	if repo.lastUpdatePath != newPath || repo.lastUpdateMime != "audio/mpeg" {
		t.Fatalf("unexpected update payload: path=%s mime=%s", repo.lastUpdatePath, repo.lastUpdateMime)
	}
	if repo.lastUpdateHash == "" {
		t.Fatal("expected checksum to be persisted")
	}
	if repo.lastUpdateLen == nil || *repo.lastUpdateLen != lengthSeconds {
		t.Fatalf("expected duration %d to be persisted, got %+v", lengthSeconds, repo.lastUpdateLen)
	}
	if store.removedPath != oldPath {
		t.Fatalf("expected old path to be removed, got %s", store.removedPath)
	}
}

func TestSongServiceDeleteSong_RemovesFromAlbumsAndAudio(t *testing.T) {
	id := primitive.NewObjectID()
	audioPath := "/spotlite/audio/song.mp3"

	repo := &fakeSongRepo{
		findByIDFn: func(context.Context, primitive.ObjectID) (*entities.Song, error) {
			return &entities.Song{ID: id, AudioPath: audioPath}, nil
		},
		deleteByIDFn: func(context.Context, primitive.ObjectID) (*mongo.DeleteResult, error) {
			return &mongo.DeleteResult{DeletedCount: 1}, nil
		},
	}
	album := &fakeAlbumManager{}
	store := &fakeAudioStore{}
	svc := newTestSongService(repo, album, store)

	if err := svc.DeleteSong(context.Background(), id.Hex()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if album.removedSongID != id.Hex() {
		t.Fatalf("expected remove from albums for %s, got %s", id.Hex(), album.removedSongID)
	}
	if store.removedPath != audioPath {
		t.Fatalf("expected audio cleanup path %s, got %s", audioPath, store.removedPath)
	}
}
