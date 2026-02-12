package services

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vanjmali/spotlite/common-lib/events"
	"github.com/vanjmali/spotlite/common-lib/middlewares"
	"github.com/vanjmali/spotlite/common-lib/subscription"
	"github.com/vanjmali/spotlite/subscription-service/dtos"
	"github.com/vanjmali/spotlite/subscription-service/entities"
	"github.com/vanjmali/spotlite/subscription-service/mappers"
	"github.com/vanjmali/spotlite/subscription-service/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeSubscriptionRepo struct {
	createFn                      func(*entities.Subscription, context.Context) error
	deleteFn                      func(primitive.ObjectID, primitive.ObjectID, context.Context) (int64, error)
	findSubscriptionsByEntityIDfn func([]string, int, string, context.Context) ([]*entities.Subscription, string, error)

	createCalled    bool
	created         *entities.Subscription
	deleteCalled    bool
	deleteUserID    primitive.ObjectID
	deleteEntity    primitive.ObjectID
	subscriptionIDs []string
	lastID          string
	batchSize       int
}

// FindSubscriptionsByEntityID implements [SubscriptionRepository].
func (f *fakeSubscriptionRepo) FindSubscriptionsByEntityID(ctx context.Context, targetIDStrs []string, batchSize int, lastID string) ([]*entities.Subscription, string, error) {
	return make([]*entities.Subscription, 0), "", nil
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
