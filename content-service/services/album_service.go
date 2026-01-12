package services

import (
	"context"
	"log"

	"github.com/vanjmali/spotlite/content/dtos"
	"github.com/vanjmali/spotlite/content/mappers"
	"github.com/vanjmali/spotlite/content/repositories"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type AlbumService struct {
	r  *repositories.AlbumRepository
	tr trace.Tracer
}

func NewAlbumService(r repositories.AlbumRepository) *AlbumService {
	tr := otel.Tracer("album-service/album-service")
	s := AlbumService{r: &r, tr: tr}

	return &s
}

func (s *AlbumService) Create(ctx context.Context, albumDto *dtos.CreateAlbumDto) error {
	ctx, span := s.tr.Start(ctx, "album.create")
	defer span.End()

	createCtx, createSpan := s.tr.Start(ctx, "album.create.create_album")
	albumEntity, err := mappers.ToAlbumEntity(albumDto)
	if err != nil {
		createSpan.RecordError(err)
		createSpan.End()
		log.Printf("Error converting to album entity: %v", err)
		return err
	}

	err = s.r.Create(createCtx, *albumEntity)
	if err != nil {
		createSpan.RecordError(err)
		createSpan.End()
		log.Printf("Error creating album in database: %v", err)
		return err
	}

	createSpan.End()
	return nil
}
