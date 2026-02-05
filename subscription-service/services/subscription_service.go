package services

import (
	"context"
	"errors"

	"github.com/vanjmali/spotlite/common-lib/middlewares"
	"github.com/vanjmali/spotlite/common-lib/subscription"
	"github.com/vanjmali/spotlite/subscription-service/dtos"
	"github.com/vanjmali/spotlite/subscription-service/entities"
	"github.com/vanjmali/spotlite/subscription-service/mappers"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrEntityNotFound       = errors.New("genre/artist couldn't be found")
	ErrSubscriptionNotFound = errors.New("subscription not found")
	ErrInvalidEntityID      = errors.New("error has ocurred while parsing genre/artist id")
	ErrUpstreamFailure      = errors.New("error has ocurred while fetching artist/genre")
)

type SubscriptionRepository interface {
	Create(s *entities.Subscription, ctx context.Context) error
	Delete(entityID primitive.ObjectID, userID primitive.ObjectID, ctx context.Context) (int64, error)
}

type ContentEntityGetter interface {
	GetEntity(ctx context.Context, entityID string, subType subscription.SubscriptionType) (string, error)
}

type SubscriptionService struct {
	sr  SubscriptionRepository
	gcc ContentEntityGetter
	tr  trace.Tracer
}

func NewSubscriptionService(sr SubscriptionRepository, gcc ContentEntityGetter) *SubscriptionService {
	tr := otel.Tracer("subscription-service/subscription-service")
	s := SubscriptionService{sr: sr, gcc: gcc, tr: tr}

	return &s
}

func (s *SubscriptionService) Subscribe(req *dtos.CreateSubscriptionDto, ctx context.Context) error {
	ctx, span := s.tr.Start(ctx, "subscription.subscribe")
	defer span.End()

	entityExistenceCtx, entityExistenceSpan := s.tr.Start(ctx, "subscription.subscribe.entity_exists")
	defer entityExistenceSpan.End()

	entityName, err := s.gcc.GetEntity(entityExistenceCtx, req.EntityID, req.Type)
	if err != nil {
		entityExistenceSpan.RecordError(err)

		st, ok := status.FromError(err)
		if !ok {
			return err
		}

		// TODO: Handle different types of errors with resiliency mechanisms
		//nolint:exhaustive
		switch st.Code() {
		case codes.NotFound:
			return ErrEntityNotFound
		case codes.InvalidArgument:
			return ErrInvalidEntityID
		default:
			return ErrUpstreamFailure
		}
	}

	userIDstr := middlewares.GetUserIdFromContext(ctx)

	se, err := mappers.ToSubscriptionEntity(req, userIDstr, entityName)
	if err != nil {
		span.RecordError(err)
		return err
	}

	createCtx, createSpan := s.tr.Start(ctx, "subscription.subscribe.save_subscription")
	defer createSpan.End()

	if err := s.sr.Create(se, createCtx); err != nil {
		createSpan.RecordError(err)
		return err
	}

	return nil
}

func (s *SubscriptionService) Unsubscribe(entityId primitive.ObjectID, ctx context.Context) error {
	ctx, span := s.tr.Start(ctx, "subscription.unsubscribe")
	defer span.End()

	userIDstr := middlewares.GetUserIdFromContext(ctx)

	userID, err := primitive.ObjectIDFromHex(userIDstr)
	if err != nil {
		span.RecordError(err)
		return err
	}

	deleteCtx, deleteSpan := s.tr.Start(ctx, "subscription.unsubcribe.delete")
	defer deleteSpan.End()

	ddc, err := s.sr.Delete(entityId, userID, deleteCtx)
	if err != nil {
		deleteSpan.RecordError(err)
		return err
	}

	if ddc != 1 {
		return ErrSubscriptionNotFound
	}

	return nil
}
