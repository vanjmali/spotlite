package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/vanjmali/spotlite/analytics-service/entities"
	"github.com/vanjmali/spotlite/analytics-service/repositories"
	"github.com/vanjmali/spotlite/common-lib/subscription"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var (
	ErrAnalyticsNotFound       = errors.New("user analytics not found")
	ErrActivityHistoryNotFound = errors.New("user activity history not found")
	ErrInvalidEvent            = errors.New("invalid event data")
)

// EventStoreRepository defines the interface for event store persistence operations
type EventStoreRepository interface {
	StoreEvent(ctx context.Context, event *entities.Event) error
}

// UserAnalyticsRepository defines the interface for user analytics read model operations
type UserAnalyticsRepository interface {
	UpsertUserAnalytics(ctx context.Context, analytics *entities.UserAnalyticsReadModel) error
	GetUserAnalytics(ctx context.Context, userID string) (*entities.UserAnalyticsReadModel, error)
}

// UserActivityHistoryRepository defines the interface for user activity history read model operations
type UserActivityHistoryRepository interface {
	UpsertUserActivityHistory(ctx context.Context, history *entities.UserActivityHistory) error
	AddActivityToHistory(ctx context.Context, userID string, activity entities.ActivitySummary) error
	GetActivityHistory(ctx context.Context, userID string) (*entities.UserActivityHistory, error)
}

// AnalyticsService provides business logic for analytics and event sourcing operations.
// It coordinates between the event store (write model) and read models (query models)
// following the CQRS pattern.
type AnalyticsService struct {
	eventStoreRepo EventStoreRepository
	analyticsRepo  UserAnalyticsRepository
	historyRepo    UserActivityHistoryRepository
	tracer         trace.Tracer
}

// NewAnalyticsService creates a new AnalyticsService with the provided repositories
func NewAnalyticsService(
	eventStoreRepo EventStoreRepository,
	analyticsRepo UserAnalyticsRepository,
	historyRepo UserActivityHistoryRepository,
) *AnalyticsService {
	return &AnalyticsService{
		eventStoreRepo: eventStoreRepo,
		analyticsRepo:  analyticsRepo,
		historyRepo:    historyRepo,
		tracer:         otel.Tracer("analytics-service/analytics-service"),
	}
}

// StoreEvent persists an event to the event store (write model).
// This is the entry point for all user activity events that need to be tracked.
func (s *AnalyticsService) StoreEvent(ctx context.Context, event *entities.Event) error {
	ctx, span := s.tracer.Start(ctx, "AnalyticsService.StoreEvent")
	defer span.End()

	if event == nil {
		return ErrInvalidEvent
	}

	if err := s.eventStoreRepo.StoreEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to store event: %w", err)
	}

	return nil
}

// GetUserAnalytics retrieves the analytics aggregate for a specific user.
// Returns the denormalized analytics data including play counts, ratings, and subscriptions.
func (s *AnalyticsService) GetUserAnalytics(ctx context.Context, userID string) (*entities.UserAnalyticsReadModel, error) {
	ctx, span := s.tracer.Start(ctx, "AnalyticsService.GetUserAnalytics")
	defer span.End()

	analytics, err := s.analyticsRepo.GetUserAnalytics(ctx, userID)
	if err != nil {
		if errors.Is(err, repositories.ErrReadModelNotFound) {
			return nil, ErrAnalyticsNotFound
		}
		return nil, fmt.Errorf("failed to get user analytics: %w", err)
	}

	return analytics, nil
}

// GetUserActivityHistory retrieves the activity timeline for a specific user.
// Returns the chronological list of user activities (most recent first).
func (s *AnalyticsService) GetUserActivityHistory(ctx context.Context, userID string) (*entities.UserActivityHistory, error) {
	ctx, span := s.tracer.Start(ctx, "AnalyticsService.GetUserActivityHistory")
	defer span.End()

	history, err := s.historyRepo.GetActivityHistory(ctx, userID)
	if err != nil {
		if errors.Is(err, repositories.ErrReadModelNotFound) {
			return nil, ErrActivityHistoryNotFound
		}
		return nil, fmt.Errorf("failed to get user activity history: %w", err)
	}

	return history, nil
}

// GetOrCreateUserAnalytics retrieves existing analytics or creates a new empty aggregate.
// Used by event projections to ensure analytics record exists before updating.
func (s *AnalyticsService) GetOrCreateUserAnalytics(ctx context.Context, userID string) (*entities.UserAnalyticsReadModel, error) {
	ctx, span := s.tracer.Start(ctx, "AnalyticsService.GetOrCreateUserAnalytics")
	defer span.End()

	analytics, err := s.analyticsRepo.GetUserAnalytics(ctx, userID)
	if err != nil {
		if errors.Is(err, repositories.ErrReadModelNotFound) {
			// Create new analytics record
			return entities.NewUserAnalyticsReadModel(userID), nil
		}
		return nil, fmt.Errorf("failed to get user analytics: %w", err)
	}

	return analytics, nil
}

// GetOrCreateUserActivityHistory retrieves existing history or creates a new empty record.
// Used by event projections to ensure activity history record exists before appending.
func (s *AnalyticsService) GetOrCreateUserActivityHistory(ctx context.Context, userID string) (*entities.UserActivityHistory, error) {
	ctx, span := s.tracer.Start(ctx, "AnalyticsService.GetOrCreateUserActivityHistory")
	defer span.End()

	history, err := s.historyRepo.GetActivityHistory(ctx, userID)
	if err != nil {
		if errors.Is(err, repositories.ErrReadModelNotFound) {
			// Create new activity history record
			return entities.NewUserActivityHistory(userID), nil
		}
		return nil, fmt.Errorf("failed to get user activity history: %w", err)
	}

	return history, nil
}

// UpdateUserAnalytics persists updated analytics aggregate to the read model.
// Used by event projections after applying events to the aggregate.
func (s *AnalyticsService) UpdateUserAnalytics(ctx context.Context, analytics *entities.UserAnalyticsReadModel) error {
	ctx, span := s.tracer.Start(ctx, "AnalyticsService.UpdateUserAnalytics")
	defer span.End()

	if analytics == nil {
		return ErrInvalidEvent
	}

	if err := s.analyticsRepo.UpsertUserAnalytics(ctx, analytics); err != nil {
		return fmt.Errorf("failed to update user analytics: %w", err)
	}

	return nil
}

// UpdateUserActivityHistory persists updated activity history aggregate to the read model.
// Used by event projections after appending new activities.
func (s *AnalyticsService) UpdateUserActivityHistory(ctx context.Context, history *entities.UserActivityHistory) error {
	ctx, span := s.tracer.Start(ctx, "AnalyticsService.UpdateUserActivityHistory")
	defer span.End()

	if history == nil {
		return ErrInvalidEvent
	}

	if err := s.historyRepo.UpsertUserActivityHistory(ctx, history); err != nil {
		return fmt.Errorf("failed to update user activity history: %w", err)
	}

	return nil
}

// ProjectSongPlayedEvent applies a song played event to read models.
// Updates analytics (play counts, genre stats, top artists) and appends activity.
func (s *AnalyticsService) ProjectSongPlayedEvent(
	ctx context.Context,
	userID string,
	genreID string,
	artistID string,
	timestamp time.Time,
) error {
	ctx, span := s.tracer.Start(ctx, "AnalyticsService.ProjectSongPlayedEvent")
	defer span.End()

	// Get or create analytics
	analytics, err := s.GetOrCreateUserAnalytics(ctx, userID)
	if err != nil {
		return err
	}

	// Apply event to analytics
	analytics.AddSongPlayed(genreID, artistID)

	// Persist updated analytics
	if err := s.UpdateUserAnalytics(ctx, analytics); err != nil {
		return err
	}

	// Append activity to history (atomic operation)
	if err := s.historyRepo.AddActivityToHistory(ctx, userID, entities.ActivitySummary{
		ActivityType: entities.EventTypeSongPlayed,
		Timestamp:    timestamp,
	}); err != nil {
		return err
	}

	return nil
}

// ProjectSubscriptionEvent applies a subscription created/deleted event to read models.
// Updates subscription counts and appends activity.
func (s *AnalyticsService) ProjectSubscriptionEvent(
	ctx context.Context,
	userID string,
	eventType string,
	subscriptionType subscription.SubscriptionType,
	timestamp time.Time,
) error {
	ctx, span := s.tracer.Start(ctx, "AnalyticsService.ProjectSubscriptionEvent")
	defer span.End()

	// Get or create analytics
	analytics, err := s.GetOrCreateUserAnalytics(ctx, userID)
	if err != nil {
		return err
	}

	// Apply event to analytics
	if eventType == entities.EventTypeSubscriptionCreated {
		analytics.AddSubscription(subscriptionType)
	} else if eventType == entities.EventTypeSubscriptionDeleted {
		analytics.DeleteSubscription(subscriptionType)
	}

	// Persist updated analytics
	if err := s.UpdateUserAnalytics(ctx, analytics); err != nil {
		return err
	}

	// Append activity to history (atomic operation)
	if err := s.historyRepo.AddActivityToHistory(ctx, userID, entities.ActivitySummary{
		ActivityType: eventType,
		Timestamp:    timestamp,
	}); err != nil {
		return err
	}

	return nil
}

// ProjectRatingEvent applies a rating event to read models.
// Updates analytics (average rating) and appends activity to history.
func (s *AnalyticsService) ProjectRatingEvent(
	ctx context.Context,
	userID string,
	eventType string,
	rating int,
	oldRating int,
	timestamp time.Time,
) error {
	ctx, span := s.tracer.Start(ctx, "AnalyticsService.ProjectRatingEvent")
	defer span.End()

	// Get or create analytics
	analytics, err := s.GetOrCreateUserAnalytics(ctx, userID)
	if err != nil {
		return err
	}

	// Apply event to analytics based on event type
	switch eventType {
	case entities.EventTypeRatingCreated:
		analytics.AddRating(rating)
	case entities.EventTypeRatingUpdated:
		analytics.UpdateRating(oldRating, rating)
	case entities.EventTypeRatingDeleted:
		analytics.DeleteRating(rating)
	}

	// Persist updated analytics
	if err := s.UpdateUserAnalytics(ctx, analytics); err != nil {
		return err
	}

	// Append activity to history (atomic operation)
	if err := s.historyRepo.AddActivityToHistory(ctx, userID, entities.ActivitySummary{
		ActivityType: eventType,
		Timestamp:    timestamp,
	}); err != nil {
		return err
	}

	return nil
}
