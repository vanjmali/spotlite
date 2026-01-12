package mappers

import (
	"github.com/vanjmali/spotlite/content/dtos"
	"github.com/vanjmali/spotlite/content/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func ToAlbumEntity(a *dtos.CreateAlbumDto) (*entities.Album, error) {
	songIds := make([]primitive.ObjectID, 0)
	for _, idStr := range a.SongIds {
		id, err := primitive.ObjectIDFromHex(idStr)
		if err != nil {
			return nil, err
		}
		songIds = append(songIds, id)
	}

	artistIds := make([]primitive.ObjectID, 0)
	for _, idStr := range a.ArtistIds {
		id, err := primitive.ObjectIDFromHex(idStr)
		if err != nil {
			return nil, err
		}
		artistIds = append(artistIds, id)
	}

	return &entities.Album{
		ID:          primitive.NewObjectID(),
		Name:        a.Name,
		ReleaseDate: a.ReleaseDate,
		Genres:      a.Genres,
		SongIds:     songIds,
		ArtistIds:   artistIds,
	}, nil
}
