package entities

import (
	"github.com/vanjmali/spotlite/common-lib/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SongRating struct {
	Average float64 `json:"average"`
	Count   int64   `json:"count"`
}

// Song models a song document stored in MongoDB.
type Song struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title         string             `bson:"title" json:"title"`
	Genres        []Genre            `bson:"genres" json:"genres"`
	LengthSeconds int                `bson:"length_seconds" json:"length_seconds"`
	Artists       []Artist           `bson:"artists" json:"artists"`
	Rating        *SongRating        `bson:"-" json:"rating,omitempty"`
	Status        types.EntityStatus `bson:"status" json:"status"`
	AudioPath     string             `bson:"audio_path,omitempty" json:"-"`
	AudioSize     int64              `bson:"audio_size,omitempty" json:"-"`
	AudioMimeType string             `bson:"audio_mime_type,omitempty" json:"-"`
	AudioChecksum string             `bson:"audio_checksum,omitempty" json:"-"`
}
