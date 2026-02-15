package entities

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Song models a song document stored in MongoDB.
type Song struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title         string             `bson:"title" json:"title"`
	Genres        []Genre            `bson:"genres" json:"genres"`
	LengthSeconds int                `bson:"length_seconds" json:"lengthSeconds"`
	Artists       []Artist           `bson:"artists" json:"artists"`
	AudioPath     string             `bson:"audio_path,omitempty" json:"audioPath,omitempty"`
	AudioSize     int64              `bson:"audio_size,omitempty" json:"audioSize,omitempty"`
	AudioMimeType string             `bson:"audio_mime_type,omitempty" json:"audioMimeType,omitempty"`
	AudioChecksum string             `bson:"audio_checksum,omitempty" json:"audioChecksum,omitempty"`
}
