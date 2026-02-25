package services

import (
	"context"
	"errors"
	"reflect"
	"slices"
	"testing"

	"github.com/sony/gobreaker"
	"github.com/stretchr/testify/require"
	"github.com/vanjmali/spotlite/common-lib/events"
	"github.com/vanjmali/spotlite/common-lib/middlewares"
	"github.com/vanjmali/spotlite/common-lib/subscription"
	"github.com/vanjmali/spotlite/subscription-service/dtos"
	"github.com/vanjmali/spotlite/subscription-service/entities"
	"github.com/vanjmali/spotlite/subscription-service/mappers"
	"github.com/vanjmali/spotlite/subscription-service/repositories"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var ErrNotSubscribed = errors.New("subscription not found")

//nolint:unused
type fakeSubscriptionRepo struct {
	createFn func(*entities.Subscription, context.Context) error
	deleteFn func(primitive.ObjectID, primitive.ObjectID, context.Context) (int64, error)

	createCalled    bool
	created         *entities.Subscription
	deleteCalled    bool
	deleteUserID    primitive.ObjectID
	deleteEntity    primitive.ObjectID
	subscriptionIDs []string
	lastID          string
	dataSubCount    map[primitive.ObjectID][]primitive.ObjectID
	batchSize       int
	store           []entities.Subscription
	data            map[primitive.ObjectID][]primitive.ObjectID
}

// FindSubscriptionsByEntityID implements [SubscriptionRepository].
func (f *fakeSubscriptionRepo) FindSubscriptionsByEntityID(
	ctx context.Context,
	targetIDStrs []string,
	batchSize int,
	lastID string,
) ([]*entities.Subscription, string, error) {
	return make([]*entities.Subscription, 0), "", nil
}

func (f *fakeSubscriptionRepo) FindSubscriptionsByUserID(ctx context.Context, filter bson.M, skip int64, limit int64) ([]entities.Subscription, int64, error) {
	targetUserID, ok := filter["subscriber_id"].(primitive.ObjectID)
	if !ok {
		return []entities.Subscription{}, 0, nil
	}

	var filtered []entities.Subscription
	for _, sub := range f.store {
		if sub.SubscriberID == targetUserID {
			filtered = append(filtered, sub)
		}
	}

	totalCount := int64(len(filtered))

	if skip > int64(len(filtered)) {
		return []entities.Subscription{}, totalCount, nil
	}
	filtered = filtered[skip:]

	if limit > 0 && int64(len(filtered)) > limit {
		filtered = filtered[:limit]
	}

	return filtered, totalCount, nil
}

func (f *fakeSubscriptionRepo) Create(s *entities.Subscription, ctx context.Context) error {
	f.createCalled = true
	f.created = s
	if f.createFn != nil {
		return f.createFn(s, ctx)
	}
	return nil
}

func (f *fakeSubscriptionRepo) Delete(
	entityID primitive.ObjectID,
	userID primitive.ObjectID,
	ctx context.Context,
) (int64, error) {
	f.deleteCalled = true
	f.deleteUserID = userID
	f.deleteEntity = entityID
	if f.deleteFn != nil {
		return f.deleteFn(entityID, userID, ctx)
	}
	return 1, nil
}

func (f *fakeSubscriptionRepo) FindEntitySubscriberCount(ctx context.Context, entityID primitive.ObjectID) (int64, error) {
	var count int64 = 0

	for _, subscribedEntities := range f.dataSubCount {
		for _, id := range subscribedEntities {
			if id == entityID {
				count++
				break // Found it for this user, stop checking this user's list (prevent double counting)
			}
		}
	}

	return count, nil
}

type fakeContentGetter struct {
	getEntityFn func(context.Context, string, subscription.SubscriptionType) (string, error)
	lastType    subscription.SubscriptionType
	lastID      string
}

func (f *fakeContentGetter) GetEntity(
	ctx context.Context,
	entityID string,
	subType subscription.SubscriptionType,
) (string, error) {
	f.lastID = entityID
	f.lastType = subType
	if f.getEntityFn != nil {
		return f.getEntityFn(ctx, entityID, subType)
	}
	return "entity-name", nil
}

func contextWithUserID(ctx context.Context, id primitive.ObjectID) context.Context {
	return middlewares.ContextWithUserID(ctx, id.Hex())
}

func (f *fakeSubscriptionRepo) AddSubscription(subID, entID primitive.ObjectID) {
	if f.data == nil {
		f.data = make(map[primitive.ObjectID][]primitive.ObjectID)
	}
	f.data[subID] = append(f.data[subID], entID)
}

func (f *fakeSubscriptionRepo) IsSubscribed(subscriberID, entityID primitive.ObjectID, ctx context.Context) error {
	subscriptions, ok := f.data[subscriberID]
	if !ok {
		return ErrNotSubscribed
	}

	if slices.Contains(subscriptions, entityID) {
		return nil
	}

	return ErrNotSubscribed
}

func TestFakeSubscriptionRepo_FindSubscriptionsByUserID(t *testing.T) {
	// Setup dummy data
	userA := primitive.NewObjectID()
	userB := primitive.NewObjectID()

	item1ID := primitive.NewObjectID()
	item2ID := primitive.NewObjectID()
	item3ID := primitive.NewObjectID()
	item4ID := primitive.NewObjectID()

	// Pre-fill the fake repo
	repo := &fakeSubscriptionRepo{
		store: []entities.Subscription{
			{ID: item1ID, SubscriberID: userA},
			{ID: item2ID, SubscriberID: userA},
			{ID: item3ID, SubscriberID: userA},
			{ID: item4ID, SubscriberID: userB},
		},
	}

	tests := []struct {
		name          string
		filter        bson.M
		skip          int64
		limit         int64
		wantCount     int64 // Total count before pagination
		wantResultIDs []primitive.ObjectID
	}{
		{
			name:          "Find all for User A",
			filter:        bson.M{"subscriber_id": userA},
			skip:          0,
			limit:         10,
			wantCount:     3,
			wantResultIDs: []primitive.ObjectID{item1ID, item2ID, item3ID},
		},
		{
			name:          "Pagination: Skip first 2 for User A",
			filter:        bson.M{"subscriber_id": userA},
			skip:          2,
			limit:         10,
			wantCount:     3,
			wantResultIDs: []primitive.ObjectID{item3ID},
		},
		{
			name:          "Pagination: Limit 1 for User A",
			filter:        bson.M{"subscriber_id": userA},
			skip:          0,
			limit:         1,
			wantCount:     3,
			wantResultIDs: []primitive.ObjectID{item1ID},
		},
		{
			name:          "Find User B (Filtering check)",
			filter:        bson.M{"subscriber_id": userB},
			skip:          0,
			limit:         10,
			wantCount:     1,
			wantResultIDs: []primitive.ObjectID{item4ID},
		},
		{
			name:          "Unknown User",
			filter:        bson.M{"subscriber_id": "ghost"},
			skip:          0,
			limit:         10,
			wantCount:     0,
			wantResultIDs: []primitive.ObjectID{}, // Expect empty slice, not nil
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSubs, gotCount, err := repo.FindSubscriptionsByUserID(context.Background(), tt.filter, tt.skip, tt.limit)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify Total Count matches (ignoring pagination)
			if gotCount != tt.wantCount {
				t.Errorf("expected count %d, got %d", tt.wantCount, gotCount)
			}

			// Map results to IDs for easier comparison
			var gotIDs []primitive.ObjectID
			for _, sub := range gotSubs {
				gotIDs = append(gotIDs, sub.ID)
			}

			// Initialize empty slice if nil to ensure DeepEqual works
			if gotIDs == nil {
				gotIDs = []primitive.ObjectID{}
			}

			if !reflect.DeepEqual(gotIDs, tt.wantResultIDs) {
				t.Errorf("expected IDs %v, got %v", tt.wantResultIDs, gotIDs)
			}
		})
	}
}

func TestFakeSubscriptionRepo_IsSubscribed(t *testing.T) {
	userA := primitive.NewObjectID()
	userB := primitive.NewObjectID()
	entityID1 := primitive.NewObjectID()
	entityID2 := primitive.NewObjectID()

	tests := []struct {
		name         string
		setup        func(*fakeSubscriptionRepo)
		subscriberID primitive.ObjectID
		entityID     primitive.ObjectID
		expectError  bool
	}{
		{
			name: "Success: User is subscribed",
			setup: func(f *fakeSubscriptionRepo) {
				f.AddSubscription(userA, entityID1)
			},
			subscriberID: userA,
			entityID:     entityID1,
			expectError:  false,
		},
		{
			name: "Failure: User exists but subscribed to different entity",
			setup: func(f *fakeSubscriptionRepo) {
				f.AddSubscription(userA, entityID1)
			},
			subscriberID: userA,
			entityID:     entityID2,
			expectError:  true,
		},
		{
			name: "Failure: User has no subscriptions",
			setup: func(f *fakeSubscriptionRepo) {
			},
			subscriberID: userB,
			entityID:     entityID1,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeSubscriptionRepo{}

			if tt.setup != nil {
				tt.setup(repo)
			}

			err := repo.IsSubscribed(tt.subscriberID, tt.entityID, context.Background())

			if tt.expectError {
				if err == nil {
					t.Error("expected an error, but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("expected nil, but got error: %v", err)
				}
			}
		})
	}
}

func TestFakeSubscriptionRepo_FindEntitySubscriberCount(t *testing.T) {
	// Generate IDs for use in tests
	userA := primitive.NewObjectID()
	userB := primitive.NewObjectID()
	userC := primitive.NewObjectID()

	targetEntity := primitive.NewObjectID()
	otherEntity := primitive.NewObjectID()

	tests := []struct {
		name          string
		setup         func(*fakeSubscriptionRepo)
		queryEntityID primitive.ObjectID
		expectedCount int64
		expectError   bool
	}{
		{
			name: "Zero count: Repo is empty",
			setup: func(f *fakeSubscriptionRepo) {
				// No data
			},
			queryEntityID: targetEntity,
			expectedCount: 0,
			expectError:   false,
		},
		{
			name: "Count 1: One user subscribed to target",
			setup: func(f *fakeSubscriptionRepo) {
				f.AddSubscriptionSC(userA, targetEntity)
			},
			queryEntityID: targetEntity,
			expectedCount: 1,
			expectError:   false,
		},
		{
			name: "Count 2: Two users subscribed to target",
			setup: func(f *fakeSubscriptionRepo) {
				f.AddSubscriptionSC(userA, targetEntity)
				f.AddSubscriptionSC(userB, targetEntity)
			},
			queryEntityID: targetEntity,
			expectedCount: 2,
			expectError:   false,
		},
		{
			name: "Mixed Data: Users subscribed to different entities",
			setup: func(f *fakeSubscriptionRepo) {
				f.AddSubscriptionSC(userA, targetEntity) // Should count
				f.AddSubscriptionSC(userB, otherEntity)  // Should NOT count
				f.AddSubscriptionSC(userC, targetEntity) // Should count
			},
			queryEntityID: targetEntity,
			expectedCount: 2,
			expectError:   false,
		},
		{
			name: "Zero count: Users exist but subscribed to other entities",
			setup: func(f *fakeSubscriptionRepo) {
				f.AddSubscriptionSC(userA, otherEntity)
			},
			queryEntityID: targetEntity,
			expectedCount: 0,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 1. Initialize Repo
			repo := &fakeSubscriptionRepo{}

			// 2. Setup Data
			if tt.setup != nil {
				tt.setup(repo)
			}

			// 3. Execute
			count, err := repo.FindEntitySubscriberCount(context.Background(), tt.queryEntityID)

			// 4. Verify Error
			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// 5. Verify Count
			if count != tt.expectedCount {
				t.Errorf("expected count %d, got %d", tt.expectedCount, count)
			}
		})
	}
}

func TestSubscribeSuccess(t *testing.T) {
	repo := &fakeSubscriptionRepo{}
	getter := &fakeContentGetter{
		getEntityFn: func(context.Context, string, subscription.SubscriptionType) (string, error) {
			return "My Genre", nil
		},
	}
	svc := NewSubscriptionService(repo, getter, events.JetStreamClient{})

	userID := primitive.NewObjectID()
	entityID := primitive.NewObjectID()
	req := &dtos.CreateSubscriptionDto{
		EntityID: entityID.Hex(),
		Type:     subscription.GenreSubscription,
	}

	err := svc.Subscribe(req, contextWithUserID(context.Background(), userID))

	require.NoError(t, err)
	require.True(t, repo.createCalled)
	require.NotNil(t, repo.created)
	require.Equal(t, "My Genre", repo.created.EntityName)
	require.Equal(t, subscription.GenreSubscription, repo.created.Type)
}

func TestSubscribeEntityNotFound(t *testing.T) {
	repo := &fakeSubscriptionRepo{}
	getter := &fakeContentGetter{
		getEntityFn: func(context.Context, string, subscription.SubscriptionType) (string, error) {
			return "", status.Error(codes.NotFound, "not found")
		},
	}
	svc := NewSubscriptionService(repo, getter, events.JetStreamClient{})

	req := &dtos.CreateSubscriptionDto{
		EntityID: primitive.NewObjectID().Hex(),
		Type:     subscription.ArtistSubscription,
	}

	err := svc.Subscribe(req, contextWithUserID(context.Background(), primitive.NewObjectID()))

	require.ErrorIs(t, err, ErrEntityNotFound)
	require.False(t, repo.createCalled)
}

func TestSubscribeInvalidEntityID(t *testing.T) {
	repo := &fakeSubscriptionRepo{}
	getter := &fakeContentGetter{
		getEntityFn: func(context.Context, string, subscription.SubscriptionType) (string, error) {
			return "", status.Error(codes.InvalidArgument, "bad id")
		},
	}
	svc := NewSubscriptionService(repo, getter, events.JetStreamClient{})

	req := &dtos.CreateSubscriptionDto{
		EntityID: primitive.NewObjectID().Hex(),
		Type:     subscription.ArtistSubscription,
	}

	err := svc.Subscribe(req, contextWithUserID(context.Background(), primitive.NewObjectID()))

	require.ErrorIs(t, err, ErrInvalidEntityID)
	require.False(t, repo.createCalled)
}

func (f *fakeSubscriptionRepo) AddSubscriptionSC(subID, entID primitive.ObjectID) {
	if f.dataSubCount == nil {
		f.dataSubCount = make(map[primitive.ObjectID][]primitive.ObjectID)
	}
	f.dataSubCount[subID] = append(f.dataSubCount[subID], entID)
}

func TestSubscribeUpstreamFailure(t *testing.T) {
	repo := &fakeSubscriptionRepo{}
	getter := &fakeContentGetter{
		getEntityFn: func(context.Context, string, subscription.SubscriptionType) (string, error) {
			return "", status.Error(codes.Internal, "boom")
		},
	}
	svc := NewSubscriptionService(repo, getter, events.JetStreamClient{})

	req := &dtos.CreateSubscriptionDto{
		EntityID: primitive.NewObjectID().Hex(),
		Type:     subscription.ArtistSubscription,
	}

	err := svc.Subscribe(req, contextWithUserID(context.Background(), primitive.NewObjectID()))

	require.ErrorIs(t, err, ErrUpstreamFailure)
	require.False(t, repo.createCalled)
}

func TestSubscribeUpstreamTimeout(t *testing.T) {
	repo := &fakeSubscriptionRepo{}
	getter := &fakeContentGetter{
		getEntityFn: func(context.Context, string, subscription.SubscriptionType) (string, error) {
			return "", status.Error(codes.DeadlineExceeded, "timeout")
		},
	}
	svc := NewSubscriptionService(repo, getter, events.JetStreamClient{})

	req := &dtos.CreateSubscriptionDto{
		EntityID: primitive.NewObjectID().Hex(),
		Type:     subscription.ArtistSubscription,
	}

	err := svc.Subscribe(req, contextWithUserID(context.Background(), primitive.NewObjectID()))

	require.ErrorIs(t, err, ErrUpstreamTimeout)
	require.False(t, repo.createCalled)
}

func TestSubscribeUpstreamUnavailable(t *testing.T) {
	repo := &fakeSubscriptionRepo{}
	getter := &fakeContentGetter{
		getEntityFn: func(context.Context, string, subscription.SubscriptionType) (string, error) {
			return "", gobreaker.ErrOpenState
		},
	}
	svc := NewSubscriptionService(repo, getter, events.JetStreamClient{})

	req := &dtos.CreateSubscriptionDto{
		EntityID: primitive.NewObjectID().Hex(),
		Type:     subscription.ArtistSubscription,
	}

	err := svc.Subscribe(req, contextWithUserID(context.Background(), primitive.NewObjectID()))

	require.ErrorIs(t, err, ErrUpstreamUnavailable)
	require.False(t, repo.createCalled)
}

func TestSubscribeUpstreamThrottled(t *testing.T) {
	repo := &fakeSubscriptionRepo{}
	getter := &fakeContentGetter{
		getEntityFn: func(context.Context, string, subscription.SubscriptionType) (string, error) {
			return "", gobreaker.ErrTooManyRequests
		},
	}
	svc := NewSubscriptionService(repo, getter, events.JetStreamClient{})

	req := &dtos.CreateSubscriptionDto{
		EntityID: primitive.NewObjectID().Hex(),
		Type:     subscription.ArtistSubscription,
	}

	err := svc.Subscribe(req, contextWithUserID(context.Background(), primitive.NewObjectID()))

	require.ErrorIs(t, err, ErrUpstreamThrottled)
	require.False(t, repo.createCalled)
}

func TestSubscribeRepoDuplicate(t *testing.T) {
	repo := &fakeSubscriptionRepo{
		createFn: func(*entities.Subscription, context.Context) error {
			return repositories.ErrSubscriptionAlreadyExists
		},
	}
	getter := &fakeContentGetter{}
	svc := NewSubscriptionService(repo, getter, events.JetStreamClient{})

	req := &dtos.CreateSubscriptionDto{
		EntityID: primitive.NewObjectID().Hex(),
		Type:     subscription.GenreSubscription,
	}

	err := svc.Subscribe(req, contextWithUserID(context.Background(), primitive.NewObjectID()))

	require.ErrorIs(t, err, repositories.ErrSubscriptionAlreadyExists)
}

func TestSubscribeMappingError(t *testing.T) {
	repo := &fakeSubscriptionRepo{}
	getter := &fakeContentGetter{}
	svc := NewSubscriptionService(repo, getter, events.JetStreamClient{})

	req := &dtos.CreateSubscriptionDto{
		EntityID: "invalid-id",
		Type:     subscription.GenreSubscription,
	}

	err := svc.Subscribe(req, contextWithUserID(context.Background(), primitive.NewObjectID()))

	require.ErrorIs(t, err, mappers.ErrSubscriptionMapping)
	require.False(t, repo.createCalled)
}

func TestUnsubscribeSuccess(t *testing.T) {
	repo := &fakeSubscriptionRepo{
		deleteFn: func(primitive.ObjectID, primitive.ObjectID, context.Context) (int64, error) {
			return 1, nil
		},
	}
	svc := NewSubscriptionService(repo, &fakeContentGetter{}, events.JetStreamClient{})

	err := svc.Unsubscribe(primitive.NewObjectID(), contextWithUserID(context.Background(), primitive.NewObjectID()))

	require.NoError(t, err)
	require.True(t, repo.deleteCalled)
}

func TestUnsubscribeNotFound(t *testing.T) {
	repo := &fakeSubscriptionRepo{
		deleteFn: func(primitive.ObjectID, primitive.ObjectID, context.Context) (int64, error) {
			return 0, nil
		},
	}
	svc := NewSubscriptionService(repo, &fakeContentGetter{}, events.JetStreamClient{})

	err := svc.Unsubscribe(primitive.NewObjectID(), contextWithUserID(context.Background(), primitive.NewObjectID()))

	require.ErrorIs(t, err, ErrSubscriptionNotFound)
}

func TestUnsubscribeRepoError(t *testing.T) {
	repo := &fakeSubscriptionRepo{
		deleteFn: func(primitive.ObjectID, primitive.ObjectID, context.Context) (int64, error) {
			return 0, errors.New("delete failed")
		},
	}
	svc := NewSubscriptionService(repo, &fakeContentGetter{}, events.JetStreamClient{})

	err := svc.Unsubscribe(primitive.NewObjectID(), contextWithUserID(context.Background(), primitive.NewObjectID()))

	require.Error(t, err)
}

func TestUnsubscribeInvalidUserID(t *testing.T) {
	repo := &fakeSubscriptionRepo{}
	svc := NewSubscriptionService(repo, &fakeContentGetter{}, events.JetStreamClient{})

	err := svc.Unsubscribe(primitive.NewObjectID(), context.Background())

	require.Error(t, err)
}
