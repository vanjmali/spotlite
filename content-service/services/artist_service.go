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

type ArtistService struct {
	r  *repositories.ArtistRepository
	tr trace.Tracer
}

// NewArtistService builds a ArtistService with repository.
func NewArtistService(r repositories.ArtistRepository) *ArtistService {
	tr := otel.Tracer("artist-service/artist-service")
	s := ArtistService{r: &r, tr: tr}

	return &s
}

func (s *ArtistService) Create(ctx context.Context, reqDto *dtos.ArtistDto) error {
	ctx, span := s.tr.Start(ctx, "artist.create")
	defer span.End()

	// Converts ArtistDto to Artist entity.
	// No uniqueness check for artist name is done here.
	createCtx, createSpan := s.tr.Start(ctx, "artist.create.create_artist")
	artistEntity, err := mappers.ToArtistEntity(reqDto)
	if err != nil {
		createSpan.RecordError(err)
		createSpan.End()
		log.Printf("Error converting to artist entity: %v", err)
		return err
	}

	// Insert the artist in database.
	err = s.r.Create(createCtx, *artistEntity)
	if err != nil {
		createSpan.RecordError(err)
		createSpan.End()
		log.Printf("Error creating artist in database: %v", err)
		return err
	}

	createSpan.End()

	return nil
}
