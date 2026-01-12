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

type SongService struct {
	r  *repositories.SongRepository
	tr trace.Tracer
}

func NewSongService(r repositories.SongRepository) *SongService {
	tr := otel.Tracer("song-service/song-service")
	s := SongService{r: &r, tr: tr}

	return &s
}

func (s *SongService) Create(ctx context.Context, songDto *dtos.CreateSongDto) error {
	ctx, span := s.tr.Start(ctx, "song.create")
	defer span.End()

	createCtx, createSpan := s.tr.Start(ctx, "song.create.create_song")
	songEntity, err := mappers.ToSongEntity(songDto)
	if err != nil {
		createSpan.RecordError(err)
		createSpan.End()
		log.Printf("Error converting to song entity: %v", err)
		return err
	}

	err = s.r.Create(createCtx, *songEntity)
	if err != nil {
		createSpan.RecordError(err)
		createSpan.End()
		log.Printf("Error creating song in database: %v", err)
		return err
	}

	createSpan.End()
	return nil

}
