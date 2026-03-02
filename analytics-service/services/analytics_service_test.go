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

func TestStoreEventSuccess(t *testing.T) {
	eventRepo := &fakeEventStoreRepo{}
	analyticsRepo := &fakeUserAnalyticsRepo{}
	svc := NewAnalyticsService(eventRepo, analyticsRepo)

	event := &entities.Event{
		UserID:    "user123",
		EventType: entities.EventTypeSongPlayed,
		Data:      map[string]any{"songID": "song456"},
		Timestamp: time.Now(),
	}

	err := svc.StoreEvent(context.Background(), event)
	require.NoError(t, err)
	require.True(t, eventRepo.storeEventCalled)
	require.Equal(t, event, eventRepo.lastEvent)
}

func TestStoreEventRepoError(t *testing.T) {
	eventRepo := &fakeEventStoreRepo{storeEventFn: func(ctx context.Context, event *entities.Event) error {
		return errFakeRepoSentinel
	}}
	analyticsRepo := &fakeUserAnalyticsRepo{}
	svc := NewAnalyticsService(eventRepo, analyticsRepo)

	err := svc.StoreEvent(context.Background(), &entities.Event{UserID: "user123"})
	require.ErrorIs(t, err, errFakeRepoSentinel)
}

func TestGetUserAnalyticsSuccess(t *testing.T) {
	expected := &entities.UserAnalyticsReadModel{UserID: "user123", TotalSongsPlayed: 42, AverageRating: 4.5}
	eventRepo := &fakeEventStoreRepo{}
	analyticsRepo := &fakeUserAnalyticsRepo{getFn: func(ctx context.Context, userID string) (*entities.UserAnalyticsReadModel, error) {
		return expected, nil
	}}
	svc := NewAnalyticsService(eventRepo, analyticsRepo)

	actual, err := svc.GetUserAnalytics(context.Background(), "user123")
	require.NoError(t, err)
	require.Equal(t, expected, actual)
	require.True(t, analyticsRepo.getCalled)
}

func TestGetUserAnalyticsNotFound(t *testing.T) {
	eventRepo := &fakeEventStoreRepo{}
	analyticsRepo := &fakeUserAnalyticsRepo{getFn: func(ctx context.Context, userID string) (*entities.UserAnalyticsReadModel, error) {
		return nil, repositories.ErrReadModelNotFound
	}}
	svc := NewAnalyticsService(eventRepo, analyticsRepo)

	actual, err := svc.GetUserAnalytics(context.Background(), "missing")
	require.Nil(t, actual)
	require.ErrorIs(t, err, ErrAnalyticsNotFound)
}

func TestProjectSongPlayedEventCreatesNewAnalytics(t *testing.T) {
	eventRepo := &fakeEventStoreRepo{}
	analyticsRepo := &fakeUserAnalyticsRepo{analyticsStore: map[string]*entities.UserAnalyticsReadModel{}}
	svc := NewAnalyticsService(eventRepo, analyticsRepo)

	err := svc.ProjectSongPlayedEvent(context.Background(), "user123", "genre1", "artist1")
	require.NoError(t, err)
	require.True(t, analyticsRepo.upsertCalled)
	require.Equal(t, 1, analyticsRepo.lastAnalytics.TotalSongsPlayed)
	require.Equal(t, 1, analyticsRepo.lastAnalytics.SongsByGenre["genre1"])
	require.Equal(t, "artist1", analyticsRepo.lastAnalytics.TopArtists[0].ArtistID)
}

func TestProjectRatingCreatedEventUpdatesAnalytics(t *testing.T) {
	eventRepo := &fakeEventStoreRepo{}
	analyticsRepo := &fakeUserAnalyticsRepo{analyticsStore: map[string]*entities.UserAnalyticsReadModel{}}
	svc := NewAnalyticsService(eventRepo, analyticsRepo)

	err := svc.ProjectRatingEvent(context.Background(), "user123", entities.EventTypeRatingCreated, 5, 0)
	require.NoError(t, err)
	require.Equal(t, 1, analyticsRepo.lastAnalytics.RatingsCount)
	require.Equal(t, 5, analyticsRepo.lastAnalytics.RatingSum)
	require.InEpsilon(t, 5.0, analyticsRepo.lastAnalytics.AverageRating, 0.0001)
}

func TestProjectRatingUpdatedEventUpdatesAverageCorrectly(t *testing.T) {
	existing := entities.NewUserAnalyticsReadModel("user123")
	existing.AddRating(4)

	eventRepo := &fakeEventStoreRepo{}
	analyticsRepo := &fakeUserAnalyticsRepo{analyticsStore: map[string]*entities.UserAnalyticsReadModel{"user123": existing}}
	svc := NewAnalyticsService(eventRepo, analyticsRepo)

	err := svc.ProjectRatingEvent(context.Background(), "user123", entities.EventTypeRatingUpdated, 5, 4)
	require.NoError(t, err)
	require.Equal(t, 1, analyticsRepo.lastAnalytics.RatingsCount)
	require.Equal(t, 5, analyticsRepo.lastAnalytics.RatingSum)
	require.InEpsilon(t, 5.0, analyticsRepo.lastAnalytics.AverageRating, 0.0001)
}

func TestProjectSubscriptionCreatedEventIncrementsCount(t *testing.T) {
	eventRepo := &fakeEventStoreRepo{}
	analyticsRepo := &fakeUserAnalyticsRepo{analyticsStore: map[string]*entities.UserAnalyticsReadModel{}}
	svc := NewAnalyticsService(eventRepo, analyticsRepo)

	err := svc.ProjectSubscriptionEvent(context.Background(), "user123", entities.EventTypeSubscriptionCreated, subscription.ArtistSubscription)
	require.NoError(t, err)
	require.Equal(t, 1, analyticsRepo.lastAnalytics.SubscribedArtistsCount)
}

func TestProjectSubscriptionDeletedEventDecrementsCount(t *testing.T) {
	existing := entities.NewUserAnalyticsReadModel("user123")
	existing.AddSubscription(subscription.ArtistSubscription)

	eventRepo := &fakeEventStoreRepo{}
	analyticsRepo := &fakeUserAnalyticsRepo{analyticsStore: map[string]*entities.UserAnalyticsReadModel{"user123": existing}}
	svc := NewAnalyticsService(eventRepo, analyticsRepo)

	err := svc.ProjectSubscriptionEvent(context.Background(), "user123", entities.EventTypeSubscriptionDeleted, subscription.ArtistSubscription)
	require.NoError(t, err)
	require.Equal(t, 0, analyticsRepo.lastAnalytics.SubscribedArtistsCount)
}

func TestProjectSubscriptionGenreTypeIgnored(t *testing.T) {
	eventRepo := &fakeEventStoreRepo{}
	analyticsRepo := &fakeUserAnalyticsRepo{analyticsStore: map[string]*entities.UserAnalyticsReadModel{}}
	svc := NewAnalyticsService(eventRepo, analyticsRepo)

	err := svc.ProjectSubscriptionEvent(context.Background(), "user123", entities.EventTypeSubscriptionCreated, subscription.GenreSubscription)
	require.NoError(t, err)
	require.Equal(t, 0, analyticsRepo.lastAnalytics.SubscribedArtistsCount)
}

func TestGetOrCreateUserAnalyticsReturnsExisting(t *testing.T) {
	existing := entities.NewUserAnalyticsReadModel("user123")
	existing.TotalSongsPlayed = 10

	eventRepo := &fakeEventStoreRepo{}
	analyticsRepo := &fakeUserAnalyticsRepo{analyticsStore: map[string]*entities.UserAnalyticsReadModel{"user123": existing}}
	svc := NewAnalyticsService(eventRepo, analyticsRepo)

	actual, err := svc.GetOrCreateUserAnalytics(context.Background(), "user123")
	require.NoError(t, err)
	require.Equal(t, existing, actual)
}

func TestGetOrCreateUserAnalyticsCreatesNew(t *testing.T) {
	eventRepo := &fakeEventStoreRepo{}
	analyticsRepo := &fakeUserAnalyticsRepo{analyticsStore: map[string]*entities.UserAnalyticsReadModel{}}
	svc := NewAnalyticsService(eventRepo, analyticsRepo)

	actual, err := svc.GetOrCreateUserAnalytics(context.Background(), "newuser")
	require.NoError(t, err)
	require.NotNil(t, actual)
	require.Equal(t, "newuser", actual.UserID)
	require.Equal(t, 0, actual.TotalSongsPlayed)
}
