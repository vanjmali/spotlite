package repositories

import (
	"context"
	"testing"

	"github.com/vanjmali/spotlite/content/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestSongRepositoryCreate(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("returns inserted object id", func(mt *mtest.T) {
		repo := NewSongRepository(mt.DB.Name(), "songs", mt.Client)
		oid := primitive.NewObjectID()
		mt.AddMockResponses(mtest.CreateSuccessResponse())

		gotID, err := repo.Create(context.Background(), entities.Song{ID: oid, Title: "Song"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotID != oid {
			t.Fatalf("expected inserted id %s, got %s", oid.Hex(), gotID.Hex())
		}
	})
}

func TestSongRepositoryUpdateAudioByID(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("updates audio metadata and decodes updated song", func(mt *mtest.T) {
		repo := NewSongRepository(mt.DB.Name(), "songs", mt.Client)
		oid := primitive.NewObjectID()
		expectedPath := "/spotlite/audio/song.mp3"

		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "value", Value: bson.D{
				{Key: "_id", Value: oid},
				{Key: "title", Value: "Song"},
				{Key: "audio_path", Value: expectedPath},
				{Key: "audio_size", Value: int64(42)},
				{Key: "audio_mime_type", Value: "audio/mpeg"},
			}},
			bson.E{Key: "ok", Value: 1},
		))

		got, err := repo.UpdateAudioByID(context.Background(), oid, expectedPath, 42, "audio/mpeg")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.AudioPath != expectedPath {
			t.Fatalf("expected audio path %q, got %q", expectedPath, got.AudioPath)
		}
		if got.AudioSize != 42 {
			t.Fatalf("expected audio size 42, got %d", got.AudioSize)
		}
		if got.AudioMimeType != "audio/mpeg" {
			t.Fatalf("expected audio mime audio/mpeg, got %q", got.AudioMimeType)
		}
	})
}
