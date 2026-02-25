package consumers

import (
	"context"
	"encoding/json"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/vanjmali/spotlite/common-lib/events"
	"github.com/vanjmali/spotlite/common-lib/logging"
	"github.com/vanjmali/spotlite/recommendation-service/entities"
	"github.com/vanjmali/spotlite/recommendation-service/repositories"
)

type RecommendationConsumer struct {
	userRepo     *repositories.UserNodeRepository
	songRepo     *repositories.SongNodeRepository
	artistRepo   *repositories.ArtistNodeRepository
	genreRepo    *repositories.GenreNodeRepository
	albumRepo    *repositories.AlbumNodeRepository
	relationRepo *repositories.GraphRelationRepository
}

func NewRecommendationConsumer(
	ur *repositories.UserNodeRepository,
	sr *repositories.SongNodeRepository,
	ar *repositories.ArtistNodeRepository,
	gr *repositories.GenreNodeRepository,
	abr *repositories.AlbumNodeRepository,
	rr *repositories.GraphRelationRepository,
) *RecommendationConsumer {
	return &RecommendationConsumer{
		userRepo:     ur,
		songRepo:     sr,
		artistRepo:   ar,
		genreRepo:    gr,
		albumRepo:    abr,
		relationRepo: rr,
	}
}

func (c *RecommendationConsumer) HandleRatingCreated(ctx context.Context, msg jetstream.Msg) error {
	var p events.RatingEventPayload
	if err := json.Unmarshal(msg.Data(), &p); err != nil {
		logging.Errorf(ctx, "critical: failed to unmarshal RatingEventPayload: %v", err)
		return nil
	}

	// Ensure user and song nodes exist
	if err := c.userRepo.Create(ctx, entities.UserNode{UserID: p.UserID}); err != nil {
		logging.Errorf(ctx, "failed to create user node for rating: %v", err)
		return err
	}

	if err := c.songRepo.Create(ctx, entities.SongNode{SongID: p.SongID}); err != nil {
		logging.Errorf(ctx, "failed to create song node for rating: %v", err)
		return err
	}

	// Create rating relationship
	rating := entities.Rating{
		UserID: p.UserID,
		SongID: p.SongID,
		Value:  p.Rating,
	}
	if err := c.relationRepo.CreateRating(ctx, rating); err != nil {
		logging.Errorf(ctx, "failed to create rating relationship: %v", err)
		return err
	}

	logging.Infof(ctx, "successfully processed rating event: user=%s, song=%s, rating=%d", p.UserID, p.SongID, p.Rating)
	return nil
}
