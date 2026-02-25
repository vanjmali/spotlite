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

// HandleRatingCreated processes rating created events and creates rating relationships in the graph.
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

// HandleRatingUpdated processes rating updated events.
func (c *RecommendationConsumer) HandleRatingUpdated(ctx context.Context, msg jetstream.Msg) error {
	var p events.RatingEventPayload
	if err := json.Unmarshal(msg.Data(), &p); err != nil {
		logging.Errorf(ctx, "critical: failed to unmarshal RatingEventPayload: %v", err)
		return nil
	}

	// Delete old rating and create new one (UPSERT pattern)
	if err := c.relationRepo.DeleteRating(ctx, p.UserID, p.SongID); err != nil {
		logging.Errorf(ctx, "failed to delete old rating: %v", err)
		return err
	}

	rating := entities.Rating{
		UserID: p.UserID,
		SongID: p.SongID,
		Value:  p.Rating,
	}
	if err := c.relationRepo.CreateRating(ctx, rating); err != nil {
		logging.Errorf(ctx, "failed to create updated rating relationship: %v", err)
		return err
	}

	logging.Infof(ctx, "successfully processed rating update event: user=%s, song=%s, rating=%d", p.UserID, p.SongID, p.Rating)
	return nil
}

// HandleRatingDeleted processes rating deleted events.
func (c *RecommendationConsumer) HandleRatingDeleted(ctx context.Context, msg jetstream.Msg) error {
	var p events.RatingEventPayload
	if err := json.Unmarshal(msg.Data(), &p); err != nil {
		logging.Errorf(ctx, "critical: failed to unmarshal RatingEventPayload: %v", err)
		return nil
	}

	if err := c.relationRepo.DeleteRating(ctx, p.UserID, p.SongID); err != nil {
		logging.Errorf(ctx, "failed to delete rating: %v", err)
		return err
	}

	logging.Infof(ctx, "successfully processed rating delete event: user=%s, song=%s", p.UserID, p.SongID)
	return nil
}

// HandleListenCreated processes listen created events and creates listen relationships in the graph.
func (c *RecommendationConsumer) HandleListenCreated(ctx context.Context, msg jetstream.Msg) error {
	var p events.ListenEventPayload
	if err := json.Unmarshal(msg.Data(), &p); err != nil {
		logging.Errorf(ctx, "critical: failed to unmarshal ListenEventPayload: %v", err)
		return nil
	}

	// Ensure user and song nodes exist
	if err := c.userRepo.Create(ctx, entities.UserNode{UserID: p.UserID}); err != nil {
		logging.Errorf(ctx, "failed to create user node for listen: %v", err)
		return err
	}

	if err := c.songRepo.Create(ctx, entities.SongNode{SongID: p.SongID}); err != nil {
		logging.Errorf(ctx, "failed to create song node for listen: %v", err)
		return err
	}

	// Create listen relationship
	listened := entities.Listened{
		UserID: p.UserID,
		SongID: p.SongID,
	}
	if err := c.relationRepo.CreateListened(ctx, listened); err != nil {
		logging.Errorf(ctx, "failed to create listen relationship: %v", err)
		return err
	}

	logging.Infof(ctx, "successfully processed listen event: user=%s, song=%s", p.UserID, p.SongID)
	return nil
}

// HandleSubscriptionCreated processes subscription created events.
func (c *RecommendationConsumer) HandleSubscriptionCreated(ctx context.Context, msg jetstream.Msg) error {
	var p events.SubscriptionEventPayload
	if err := json.Unmarshal(msg.Data(), &p); err != nil {
		logging.Errorf(ctx, "critical: failed to unmarshal SubscriptionEventPayload: %v", err)
		return nil
	}

	// Ensure user node exists
	if err := c.userRepo.Create(ctx, entities.UserNode{UserID: p.UserID}); err != nil {
		logging.Errorf(ctx, "failed to create user node for subscription: %v", err)
		return err
	}

	switch p.EntityType {
	case events.SubscriptionEntityArtist:
		// Ensure artist node exists
		if err := c.artistRepo.Create(ctx, entities.ArtistNode{ArtistID: p.EntityID}); err != nil {
			logging.Errorf(ctx, "failed to create artist node: %v", err)
			return err
		}

		// Create artist subscription
		sub := entities.ArtistSubscription{
			UserID:   p.UserID,
			ArtistID: p.EntityID,
		}
		if err := c.relationRepo.CreateArtistSubscription(ctx, sub); err != nil {
			logging.Errorf(ctx, "failed to create artist subscription: %v", err)
			return err
		}

	case events.SubscriptionEntityGenre:
		// Ensure genre node exists
		if err := c.genreRepo.Create(ctx, entities.GenreNode{GenreID: p.EntityID}); err != nil {
			logging.Errorf(ctx, "failed to create genre node: %v", err)
			return err
		}

		// Create genre subscription
		sub := entities.GenreSubscription{
			UserID:  p.UserID,
			GenreID: p.EntityID,
		}
		if err := c.relationRepo.CreateGenreSubscription(ctx, sub); err != nil {
			logging.Errorf(ctx, "failed to create genre subscription: %v", err)
			return err
		}

	default:
		logging.Warnf(ctx, "unknown entity type in subscription event: %s", p.EntityType)
		return nil
	}

	logging.Infof(ctx, "successfully processed subscription created event: user=%s, entity=%s, type=%s", p.UserID, p.EntityID, p.EntityType)
	return nil
}

// HandleSubscriptionDeleted processes subscription deleted events.
func (c *RecommendationConsumer) HandleSubscriptionDeleted(ctx context.Context, msg jetstream.Msg) error {
	var p events.SubscriptionEventPayload
	if err := json.Unmarshal(msg.Data(), &p); err != nil {
		logging.Errorf(ctx, "critical: failed to unmarshal SubscriptionEventPayload: %v", err)
		return nil
	}

	switch p.EntityType {
	case events.SubscriptionEntityArtist:
		if err := c.relationRepo.DeleteArtistSubscription(ctx, p.UserID, p.EntityID); err != nil {
			logging.Errorf(ctx, "failed to delete artist subscription: %v", err)
			return err
		}

	case events.SubscriptionEntityGenre:
		if err := c.relationRepo.DeleteGenreSubscription(ctx, p.UserID, p.EntityID); err != nil {
			logging.Errorf(ctx, "failed to delete genre subscription: %v", err)
			return err
		}

	default:
		logging.Warnf(ctx, "unknown entity type in subscription delete event: %s", p.EntityType)
		return nil
	}

	logging.Infof(ctx, "successfully processed subscription deleted event: user=%s, entity=%s, type=%s", p.UserID, p.EntityID, p.EntityType)
	return nil
}
