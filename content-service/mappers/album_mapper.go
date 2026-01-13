package mappers

import (
	"github.com/vanjmali/spotlite/content/dtos"
	"github.com/vanjmali/spotlite/content/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func ToAlbumEntity(a *dtos.CreateAlbumDto, artists []entities.Artist, songs []entities.Song) (*entities.Album, error) {
	return &entities.Album{
		ID:          primitive.NewObjectID(),
		Name:        a.Name,
		ReleaseDate: a.ReleaseDate,
		Genres:      a.Genres,
		Songs:       songs,
		Artists:     artists,
	}, nil
}
