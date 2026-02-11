package mappers

import (
	"github.com/vanjmali/spotlite/content/dtos"
	"github.com/vanjmali/spotlite/content/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ToAlbumEntity converts a CreateAlbumDto to an Album entity with associated artists and genres.
func ToAlbumEntity(a *dtos.CreateAlbumDto, artists []entities.Artist, genres []entities.Genre) (*entities.Album, error) {
	return &entities.Album{
		ID:          primitive.NewObjectID(),
		Title:       a.Title,
		ReleaseDate: a.ReleaseDate,
		Genres:      genres,
		Songs:       []entities.Song{},
		Artists:     artists,
	}, nil
}
