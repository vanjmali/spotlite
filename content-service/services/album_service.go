package services

import (
	"context"
	"log"

	"github.com/vanjmali/spotlite/content/dtos"
	"github.com/vanjmali/spotlite/content/entities"
	"github.com/vanjmali/spotlite/content/mappers"
	"github.com/vanjmali/spotlite/content/repositories"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type AlbumService struct {
	albumRepo     *repositories.AlbumRepository
	artistService *ArtistService
	songService   *SongService
	tr            trace.Tracer
}

func NewAlbumService(r repositories.AlbumRepository, artistService ArtistService, songService SongService) *AlbumService {
	tr := otel.Tracer("album-service/album-service")
	s := AlbumService{albumRepo: &r, artistService: &artistService, songService: &songService, tr: tr}

	return &s
}

func (s *AlbumService) Create(ctx context.Context, albumDto *dtos.CreateAlbumDto) error {
	ctx, span := s.tr.Start(ctx, "album.create")
	defer span.End()

	resolveArtistCtx, resolveArtistSpan := s.tr.Start(ctx, "album.create.resolve_artist")

	embeddedArtist := make([]entities.Artist, 0)

	for _, artistsIdStr := range albumDto.ArtistIds {
		artist, err := s.artistService.FindArtistByID(resolveArtistCtx, artistsIdStr)
		if err != nil {
			resolveArtistSpan.RecordError(err)
			resolveArtistSpan.End()
			return ErrArtistNotFound
		}

		embeddedArtist = append(embeddedArtist, entities.Artist{
			ID:          artist.ID,
			Name:        artist.Name,
			Genres:      artist.Genres,
			Description: artist.Description,
		})
	}

	resolveArtistSpan.End()

	resolveSongCtx, resolveSongSpan := s.tr.Start(ctx, "album.create.resolve_song")

	embeddedSong := make([]entities.Song, 0)

	for _, songsIdStr := range albumDto.SongIds {
		song, err := s.songService.FindSongById(resolveSongCtx, songsIdStr)
		if err != nil {
			resolveSongSpan.RecordError(err)
			resolveSongSpan.End()
			return ErrSongNotFound
		}

		embeddedSong = append(embeddedSong, entities.Song{
			ID:            song.ID,
			Title:         song.Title,
			Genre:         song.Genre,
			LengthSeconds: song.LengthSeconds,
			Artists:       song.Artists,
		})
	}

	resolveSongSpan.End()

	createCtx, createSpan := s.tr.Start(ctx, "album.create.create_album")
	albumEntity, err := mappers.ToAlbumEntity(albumDto, embeddedArtist, embeddedSong)
	if err != nil {
		createSpan.RecordError(err)
		createSpan.End()
		log.Printf("Error converting to album entity: %v", err)
		return err
	}

	err = s.albumRepo.Create(createCtx, *albumEntity)
	if err != nil {
		createSpan.RecordError(err)
		createSpan.End()
		log.Printf("Error creating album in database: %v", err)
		return err
	}

	createSpan.End()
	return nil
}
