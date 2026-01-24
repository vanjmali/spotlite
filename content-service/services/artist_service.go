package services

import (
	"context"
	"errors"
	"log"

	"github.com/vanjmali/spotlite/common-lib/pagination"
	"github.com/vanjmali/spotlite/common-lib/telemetry"
	"github.com/vanjmali/spotlite/content/dtos"
	"github.com/vanjmali/spotlite/content/entities"
	"github.com/vanjmali/spotlite/content/mappers"
	"github.com/vanjmali/spotlite/content/repositories"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var (
	ErrObjectIdCastFailed = errors.New("failed to convert hex to objectId")
	ErrArtistNotFound     = errors.New("artist not found")
)

type ArtistService struct {
	r            *repositories.ArtistRepository
	genreService *GenreService
	tr           trace.Tracer
}

// NewArtistService builds a ArtistService with repository.
func NewArtistService(r repositories.ArtistRepository, genreService GenreService) *ArtistService {
	tr := otel.Tracer("content-service/artist-service")
	s := ArtistService{r: &r, genreService: &genreService, tr: tr}
	return &s
}

// Create creates a new artist with the provided data.
func (s *ArtistService) Create(ctx context.Context, reqDto *dtos.ArtistDto) error {
	ctx, span := s.tr.Start(ctx, "artist.create")
	defer span.End()

	resolveGenreCtx, resolveGenreSpan := s.tr.Start(ctx, "artist.create.resolve_genres")

	embeddedGenre := make([]entities.Genre, 0)

	for _, genresIdStr := range reqDto.GenreIds {
		genre, err := s.genreService.FindGenreByID(resolveGenreCtx, genresIdStr)
		if err != nil {
			resolveGenreSpan.RecordError(err)
			resolveGenreSpan.End()

			switch {
			case errors.Is(err, ErrObjectIdCastFailed):
				return ErrObjectIdCastFailed
			case errors.Is(err, ErrGenreNotFound):
				return ErrGenreNotFound
			default:
				return err
			}
		}

		embeddedGenre = append(embeddedGenre, entities.Genre{
			ID:   genre.ID,
			Name: genre.Name,
		})
	}

	resolveGenreSpan.End()

	// Converts ArtistDto to Artist entity.
	// No uniqueness check for artist name is done here.
	createCtx, createSpan := s.tr.Start(ctx, "artist.create.create_artist")
	artistEntity, err := mappers.ToArtistEntity(reqDto, embeddedGenre)
	if err != nil {
		createSpan.RecordError(err)
		createSpan.End()
		log.Printf("trace_id=%s error converting to artist entity: %v", telemetry.TraceID(ctx), err)
		return err
	}

	// Insert the artist in database.
	err = s.r.Create(createCtx, *artistEntity)
	if err != nil {
		createSpan.RecordError(err)
		createSpan.End()
		log.Printf("trace_id=%s error creating artist in database: %v", telemetry.TraceID(ctx), err)
		return err
	}

	createSpan.End()

	return nil
}

// FindArtistByID retrieves a single artist by its ID.
func (s *ArtistService) FindArtistByID(ctx context.Context, idStr string) (*entities.Artist, error) {
	ctx, span := s.tr.Start(ctx, "artist.find_by_id")
	defer span.End()

	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		span.RecordError(err)
		return nil, ErrObjectIdCastFailed
	}

	artist, err := s.r.FindByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		return nil, ErrArtistNotFound
	}
	return artist, nil
}

// UpdateArtist updates an existing artist with the provided partial data.
func (s *ArtistService) UpdateArtist(ctx context.Context, idStr string, dto dtos.UpdateArtistDto) (*entities.Artist, error) {
	ctx, span := s.tr.Start(ctx, "artist.update_artist")
	defer span.End()

	_, parseSpan := s.tr.Start(ctx, "artist.update_artist.parse_id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		parseSpan.RecordError(err)
		parseSpan.End()
		return nil, ErrObjectIdCastFailed
	}
	parseSpan.End()

	_, buildSpan := s.tr.Start(ctx, "artist.update.build_update_doc")

	update := make(map[string]any)

	if dto.Name != nil {
		update["name"] = *dto.Name
	}
	if dto.GenreIds != nil {
		embeddedGenres := make([]entities.Genre, 0)
		for _, genreIdStr := range *dto.GenreIds {
			genre, err := s.genreService.FindGenreByID(ctx, genreIdStr)
			if err != nil {
				buildSpan.RecordError(err)
				buildSpan.End()

				switch {
				case errors.Is(err, ErrObjectIdCastFailed):
					return nil, ErrObjectIdCastFailed
				case errors.Is(err, ErrGenreNotFound):
					return nil, ErrGenreNotFound
				default:
					return nil, err
				}
			}

			embeddedGenres = append(embeddedGenres, entities.Genre{
				ID:   genre.ID,
				Name: genre.Name,
			})
		}
		update["genres"] = embeddedGenres
	}
	if dto.Description != nil {
		update["description"] = *dto.Description
	}

	if len(update) == 0 {
		err := errors.New("no fields to update")
		buildSpan.RecordError(err)
		buildSpan.End()
		return nil, err
	}
	buildSpan.End()

	repoCtx, repoSpan := s.tr.Start(ctx, "artist.update.repository_update")

	updatedArtist, err := s.r.UpdateByID(repoCtx, id, update)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			repoSpan.RecordError(err)
			repoSpan.End()
			return nil, ErrArtistNotFound
		}
		repoSpan.RecordError(err)
		repoSpan.End()
		return nil, err
	}
	repoSpan.End()

	return updatedArtist, nil
}

// DeleteArtist deletes an artist by its ID.
func (s *ArtistService) DeleteArtist(ctx context.Context, idStr string) error {
	ctx, span := s.tr.Start(ctx, "artist.delete_artist")
	defer span.End()

	_, parseSpan := s.tr.Start(ctx, "artist.update_artist.parse_id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		parseSpan.RecordError(err)
		parseSpan.End()
		return ErrObjectIdCastFailed
	}
	parseSpan.End()

	repoCtx, repoSpan := s.tr.Start(ctx, "artist.delete.repository_delete")
	res, err := s.r.DeleteByID(repoCtx, id)
	if err != nil {
		repoSpan.RecordError(err)
		repoSpan.End()
		return err
	}

	if res.DeletedCount == 0 {
		err = ErrArtistNotFound
		repoSpan.RecordError(err)
		repoSpan.End()
		return err
	}
	repoSpan.End()
	return nil
}

// ArtistsQuery represents the query parameters for filtering and paginating artist results.
type ArtistsQuery struct {
	Page  int
	Size  int
	Name  string
	Genre string
}

// GetArtists retrieves a paginated list of artists with optional filtering by name or genre.
func (s *ArtistService) GetArtists(ctx context.Context, q ArtistsQuery) (*dtos.ArtistListResponseDto, error) {
	ctx, span := s.tr.Start(ctx, "artists.get_all")
	defer span.End()

	filter := bson.M{}
	if q.Name != "" {
		filter["name"] = bson.M{
			"$regex":   q.Name,
			"$options": "i",
		}
	}

	if q.Genre != "" {
		filter["genres.name"] = bson.M{
			"$regex":   q.Genre,
			"$options": "i",
		}
	}

	p := pagination.NewPagination(q.Page, q.Size)
	items, total, err := s.r.FindAll(ctx, filter, p.Skip(), p.Limit())
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return &dtos.ArtistListResponseDto{
		Items: items,
		Page:  p.Page,
		Size:  p.Size,
		Total: total,
	}, nil
}
