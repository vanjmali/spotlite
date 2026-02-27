package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vanjmali/spotlite/analytics-service/entities"
	"github.com/vanjmali/spotlite/analytics-service/repositories"
	"github.com/vanjmali/spotlite/common-lib/subscription"
)

var errFakeRepoSentinel = errors.New("fake repo sentinel")

// Fake repositories for testing
type fakeEventStoreRepo struct {
	storeEventFn     func(ctx context.Context, event *entities.Event) error
	storeEventCalled bool
	lastEvent        *entities.Event
}

func (f *fakeEventStoreRepo) StoreEvent(ctx context.Context, event *entities.Event) error {
	f.storeEventCalled = true
	f.lastEvent = event
	if f.storeEventFn != nil {
		return f.storeEventFn(ctx, event)
	}
	return nil
}

type fakeUserAnalyticsRepo struct {
	upsertFn       func(ctx context.Context, analytics *entities.UserAnalyticsReadModel) error
	getFn          func(ctx context.Context, userID string) (*entities.UserAnalyticsReadModel, error)
	upsertCalled   bool
	getCalled      bool
	lastAnalytics  *entities.UserAnalyticsReadModel
	analyticsStore map[string]*entities.UserAnalyticsReadModel
}

func (f *fakeUserAnalyticsRepo) UpsertUserAnalytics(ctx context.Context, analytics *entities.UserAnalyticsReadModel) error {
	f.upsertCalled = true
	f.lastAnalytics = analytics
	if f.analyticsStore != nil {
		f.analyticsStore[analytics.UserID] = analytics
	}
	if f.upsertFn != nil {
		return f.upsertFn(ctx, analytics)
	}
	return nil
}

func (f *fakeUserAnalyticsRepo) GetUserAnalytics(ctx context.Context, userID string) (*entities.UserAnalyticsReadModel, error) {
	f.getCalled = true
	if f.getFn != nil {
		return f.getFn(ctx, userID)
	}
	if f.analyticsStore != nil {
		if analytics, ok := f.analyticsStore[userID]; ok {
			return analytics, nil
		}
	}
	return nil, repositories.ErrReadModelNotFound
}

type fakeUserActivityHistoryRepo struct {
	upsertFn     func(ctx context.Context, history *entities.UserActivityHistory) error
	getFn        func(ctx context.Context, userID string) (*entities.UserActivityHistory, error)
	upsertCalled bool
	getCalled    bool
	lastHistory  *entities.UserActivityHistory
	historyStore map[string]*entities.UserActivityHistory
}

func (f *fakeUserActivityHistoryRepo) UpsertUserActivityHistory(ctx context.Context, history *entities.UserActivityHistory) error {
	f.upsertCalled = true
	f.lastHistory = history
	if f.historyStore != nil {
		f.historyStore[history.UserID] = history
	}
	if f.upsertFn != nil {
		return f.upsertFn(ctx, history)
	}
	return nil
}

func (f *fakeUserActivityHistoryRepo) GetActivityHistory(ctx context.Context, userID string) (*entities.UserActivityHistory, error) {
	f.getCalled = true
	if f.getFn != nil {
		return f.getFn(ctx, userID)
	}
	if f.historyStore != nil {
		if history, ok := f.historyStore[userID]; ok {
			return history, nil
		}
	}
	return nil, repositories.ErrReadModelNotFound
}

// Test StoreEvent
func TestStoreEventSuccess(t *testing.T) {
	eventRepo := &fakeEventStoreRepo{}
	analyticsRepo := &fakeUserAnalyticsRepo{}
	historyRepo := &fakeUserActivityHistoryRepo{}
	svc := NewAnalyticsService(eventRepo, analyticsRepo, historyRepo)

	event := &entities.Event{
		UserID:    "user123",
		EventType: entities.EventTypeSongPlayed,
		Data: map[string]interface{}{
			"songID": "song456",
		},
		Timestamp: time.Now(),
	}

	err := svc.StoreEvent(context.Background(), event)

	require.NoError(t, err)
	require.True(t, eventRepo.storeEventCalled)
	require.Equal(t, event, eventRepo.lastEvent)
}

func TestStoreEventRepoError(t *testing.T) {
	eventRepo := &fakeEventStoreRepo{
		storeEventFn: func(ctx context.Context, event *entities.Event) error {
			return errFakeRepoSentinel
		},
	}
	analyticsRepo := &fakeUserAnalyticsRepo{}
	historyRepo := &fakeUserActivityHistoryRepo{}
	svc := NewAnalyticsService(eventRepo, analyticsRepo, historyRepo)

	event := &entities.Event{
		UserID:    "user123",
		EventType: entities.EventTypeSongPlayed,
		Data:      map[string]interface{}{},
		Timestamp: time.Now(),
	}

	err := svc.StoreEvent(context.Background(), event)

	require.ErrorIs(t, err, errFakeRepoSentinel)
}

// Test GetUserAnalytics
func TestGetUserAnalyticsSuccess(t *testing.T) {
	expectedAnalytics := &entities.UserAnalyticsReadModel{
		UserID:           "user123",
		TotalSongsPlayed: 42,
		AverageRating:    4.5,
	}

	eventRepo := &fakeEventStoreRepo{}
	analyticsRepo := &fakeUserAnalyticsRepo{
		getFn: func(ctx context.Context, userID string) (*entities.UserAnalyticsReadModel, error) {
			return expectedAnalytics, nil
		},
	}
	historyRepo := &fakeUserActivityHistoryRepo{}
	svc := NewAnalyticsService(eventRepo, analyticsRepo, historyRepo)

	analytics, err := svc.GetUserAnalytics(context.Background(), "user123")

	require.NoError(t, err)
	require.Equal(t, expectedAnalytics, analytics)
	require.True(t, analyticsRepo.getCalled)
}

func TestGetUserAnalyticsNotFound(t *testing.T) {
	eventRepo := &fakeEventStoreRepo{}
	analyticsRepo := &fakeUserAnalyticsRepo{
		getFn: func(ctx context.Context, userID string) (*entities.UserAnalyticsReadModel, error) {
			return nil, repositories.ErrReadModelNotFound
		},
	}
	historyRepo := &fakeUserActivityHistoryRepo{}
	svc := NewAnalyticsService(eventRepo, analyticsRepo, historyRepo)

	analytics, err := svc.GetUserAnalytics(context.Background(), "nonexistent")

	require.ErrorIs(t, err, ErrAnalyticsNotFound)
	require.Nil(t, analytics)
}

// Test GetUserActivityHistory
func TestGetUserActivityHistorySuccess(t *testing.T) {
	expectedHistory := &entities.UserActivityHistory{
		UserID: "user123",
		Activities: []entities.ActivitySummary{
			{
				ActivityType: entities.EventTypeSongPlayed,
				Timestamp:    time.Now(),
			},
		},
	}

	eventRepo := &fakeEventStoreRepo{}
	analyticsRepo := &fakeUserAnalyticsRepo{}
	historyRepo := &fakeUserActivityHistoryRepo{
		getFn: func(ctx context.Context, userID string) (*entities.UserActivityHistory, error) {
			return expectedHistory, nil
		},
	}
	svc := NewAnalyticsService(eventRepo, analyticsRepo, historyRepo)

	history, err := svc.GetUserActivityHistory(context.Background(), "user123")

	require.NoError(t, err)
	require.Equal(t, expectedHistory, history)
	require.True(t, historyRepo.getCalled)
}

func TestGetUserActivityHistoryNotFound(t *testing.T) {
	eventRepo := &fakeEventStoreRepo{}
	analyticsRepo := &fakeUserAnalyticsRepo{}
	historyRepo := &fakeUserActivityHistoryRepo{
		getFn: func(ctx context.Context, userID string) (*entities.UserActivityHistory, error) {
			return nil, repositories.ErrReadModelNotFound
		},
	}
	svc := NewAnalyticsService(eventRepo, analyticsRepo, historyRepo)

	history, err := svc.GetUserActivityHistory(context.Background(), "nonexistent")

	require.ErrorIs(t, err, ErrActivityHistoryNotFound)
	require.Nil(t, history)
}

// Test ProjectSongPlayedEvent
func TestProjectSongPlayedEventCreatesNewAnalytics(t *testing.T) {
	eventRepo := &fakeEventStoreRepo{}
	analyticsRepo := &fakeUserAnalyticsRepo{
		analyticsStore: make(map[string]*entities.UserAnalyticsReadModel),
	}
	historyRepo := &fakeUserActivityHistoryRepo{
		historyStore: make(map[string]*entities.UserActivityHistory),
	}
	svc := NewAnalyticsService(eventRepo, analyticsRepo, historyRepo)

	err := svc.ProjectSongPlayedEvent(context.Background(), "user123", "genre1", "artist1", time.Now())

	require.NoError(t, err)
	require.True(t, analyticsRepo.upsertCalled)
	require.NotNil(t, analyticsRepo.lastAnalytics)
	require.Equal(t, "user123", analyticsRepo.lastAnalytics.UserID)
	require.Equal(t, 1, analyticsRepo.lastAnalytics.TotalSongsPlayed)
	require.Equal(t, 1, analyticsRepo.lastAnalytics.SongsByGenre["genre1"])
	require.Len(t, analyticsRepo.lastAnalytics.TopArtists, 1)
	require.Equal(t, "artist1", analyticsRepo.lastAnalytics.TopArtists[0].ArtistID)
}

// Test ProjectRatingEvent
func TestProjectRatingCreatedEventUpdatesAnalytics(t *testing.T) {
	eventRepo := &fakeEventStoreRepo{}
	analyticsRepo := &fakeUserAnalyticsRepo{
		analyticsStore: make(map[string]*entities.UserAnalyticsReadModel),
	}
	historyRepo := &fakeUserActivityHistoryRepo{
		historyStore: make(map[string]*entities.UserActivityHistory),
	}
	svc := NewAnalyticsService(eventRepo, analyticsRepo, historyRepo)

	err := svc.ProjectRatingEvent(context.Background(), "user123", entities.EventTypeRatingCreated, 5, 0, time.Now())

	require.NoError(t, err)
	require.True(t, analyticsRepo.upsertCalled)
	require.NotNil(t, analyticsRepo.lastAnalytics)
	require.Equal(t, 1, analyticsRepo.lastAnalytics.RatingsCount)
	require.Equal(t, 5, analyticsRepo.lastAnalytics.RatingSum)
	require.Equal(t, 5.0, analyticsRepo.lastAnalytics.AverageRating)
}

func TestProjectRatingUpdatedEventUpdatesAverageCorrectly(t *testing.T) {
	existing := entities.NewUserAnalyticsReadModel("user123")
	existing.AddRating(4) // Initial rating: sum=4, count=1, avg=4.0

	eventRepo := &fakeEventStoreRepo{}
	analyticsRepo := &fakeUserAnalyticsRepo{
		analyticsStore: map[string]*entities.UserAnalyticsReadModel{
			"user123": existing,
		},
	}
	historyRepo := &fakeUserActivityHistoryRepo{
		historyStore: make(map[string]*entities.UserActivityHistory),
	}
	svc := NewAnalyticsService(eventRepo, analyticsRepo, historyRepo)

	// Update rating from 4 to 5
	err := svc.ProjectRatingEvent(context.Background(), "user123", entities.EventTypeRatingUpdated, 5, 4, time.Now())

	require.NoError(t, err)
	require.Equal(t, 1, analyticsRepo.lastAnalytics.RatingsCount) // Count stays the same
	require.Equal(t, 5, analyticsRepo.lastAnalytics.RatingSum)    // Sum: 4-4+5 = 5
	require.Equal(t, 5.0, analyticsRepo.lastAnalytics.AverageRating)
}

// Test ProjectSubscriptionEvent
func TestProjectSubscriptionCreatedEventIncrementsCount(t *testing.T) {
	eventRepo := &fakeEventStoreRepo{}
	analyticsRepo := &fakeUserAnalyticsRepo{
		analyticsStore: make(map[string]*entities.UserAnalyticsReadModel),
	}
	historyRepo := &fakeUserActivityHistoryRepo{
		historyStore: make(map[string]*entities.UserActivityHistory),
	}
	svc := NewAnalyticsService(eventRepo, analyticsRepo, historyRepo)

	err := svc.ProjectSubscriptionEvent(context.Background(), "user123", entities.EventTypeSubscriptionCreated, subscription.ArtistSubscription, time.Now())

	require.NoError(t, err)
	require.True(t, analyticsRepo.upsertCalled)
	require.NotNil(t, analyticsRepo.lastAnalytics)
	require.Equal(t, 1, analyticsRepo.lastAnalytics.SubscribedArtistsCount)
}

func TestProjectSubscriptionDeletedEventDecrementsCount(t *testing.T) {
	existing := entities.NewUserAnalyticsReadModel("user123")
	existing.AddSubscription(subscription.ArtistSubscription) // Start with 1 subscription

	eventRepo := &fakeEventStoreRepo{}
	analyticsRepo := &fakeUserAnalyticsRepo{
		analyticsStore: map[string]*entities.UserAnalyticsReadModel{
			"user123": existing,
		},
	}
	historyRepo := &fakeUserActivityHistoryRepo{
		historyStore: make(map[string]*entities.UserActivityHistory),
	}
	svc := NewAnalyticsService(eventRepo, analyticsRepo, historyRepo)

	err := svc.ProjectSubscriptionEvent(context.Background(), "user123", entities.EventTypeSubscriptionDeleted, subscription.ArtistSubscription, time.Now())

	require.NoError(t, err)
	require.Equal(t, 0, analyticsRepo.lastAnalytics.SubscribedArtistsCount)
}

func TestProjectSubscriptionGenreTypeIgnored(t *testing.T) {
	eventRepo := &fakeEventStoreRepo{}
	analyticsRepo := &fakeUserAnalyticsRepo{
		analyticsStore: make(map[string]*entities.UserAnalyticsReadModel),
	}
	historyRepo := &fakeUserActivityHistoryRepo{
		historyStore: make(map[string]*entities.UserActivityHistory),
	}
	svc := NewAnalyticsService(eventRepo, analyticsRepo, historyRepo)

	// Genre subscriptions should not affect SubscribedArtistsCount
	err := svc.ProjectSubscriptionEvent(context.Background(), "user123", entities.EventTypeSubscriptionCreated, subscription.GenreSubscription, time.Now())

	require.NoError(t, err)
	require.Equal(t, 0, analyticsRepo.lastAnalytics.SubscribedArtistsCount)
}

// Test GetOrCreateUserAnalytics
func TestGetOrCreateUserAnalyticsReturnsExisting(t *testing.T) {
	existing := entities.NewUserAnalyticsReadModel("user123")
	existing.TotalSongsPlayed = 10

	eventRepo := &fakeEventStoreRepo{}
	analyticsRepo := &fakeUserAnalyticsRepo{
		analyticsStore: map[string]*entities.UserAnalyticsReadModel{
			"user123": existing,
		},
	}
	historyRepo := &fakeUserActivityHistoryRepo{}
	svc := NewAnalyticsService(eventRepo, analyticsRepo, historyRepo)

	analytics, err := svc.GetOrCreateUserAnalytics(context.Background(), "user123")

	require.NoError(t, err)
	require.Equal(t, existing, analytics)
	require.Equal(t, 10, analytics.TotalSongsPlayed)
}

func TestGetOrCreateUserAnalyticsCreatesNew(t *testing.T) {
	eventRepo := &fakeEventStoreRepo{}
	analyticsRepo := &fakeUserAnalyticsRepo{
		analyticsStore: make(map[string]*entities.UserAnalyticsReadModel),
	}
	historyRepo := &fakeUserActivityHistoryRepo{}
	svc := NewAnalyticsService(eventRepo, analyticsRepo, historyRepo)

	analytics, err := svc.GetOrCreateUserAnalytics(context.Background(), "newuser")

	require.NoError(t, err)
	require.NotNil(t, analytics)
	require.Equal(t, "newuser", analytics.UserID)
	require.Equal(t, 0, analytics.TotalSongsPlayed)
}

// Test GetOrCreateUserActivityHistory
func TestGetOrCreateUserActivityHistoryReturnsExisting(t *testing.T) {
	existing := entities.NewUserActivityHistory("user123")

	eventRepo := &fakeEventStoreRepo{}
	analyticsRepo := &fakeUserAnalyticsRepo{}
	historyRepo := &fakeUserActivityHistoryRepo{
		historyStore: map[string]*entities.UserActivityHistory{
			"user123": existing,
		},
	}
	svc := NewAnalyticsService(eventRepo, analyticsRepo, historyRepo)

	history, err := svc.GetOrCreateUserActivityHistory(context.Background(), "user123")

	require.NoError(t, err)
	require.Equal(t, existing, history)
}

func TestGetOrCreateUserActivityHistoryCreatesNew(t *testing.T) {
	eventRepo := &fakeEventStoreRepo{}
	analyticsRepo := &fakeUserAnalyticsRepo{}
	historyRepo := &fakeUserActivityHistoryRepo{
		historyStore: make(map[string]*entities.UserActivityHistory),
	}
	svc := NewAnalyticsService(eventRepo, analyticsRepo, historyRepo)

	history, err := svc.GetOrCreateUserActivityHistory(context.Background(), "newuser")

	require.NoError(t, err)
	require.NotNil(t, history)
	require.Equal(t, "newuser", history.UserID)
	require.Empty(t, history.Activities)
}
