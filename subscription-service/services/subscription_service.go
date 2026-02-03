package services

import (
	"context"
	"errors"

	"github.com/vanjmali/spotlite/common-lib/middlewares"
	"github.com/vanjmali/spotlite/common-lib/subscription"
	"github.com/vanjmali/spotlite/subscription-service/dtos"
	adapters "github.com/vanjmali/spotlite/subscription-service/infrastructure/grpc"
	"github.com/vanjmali/spotlite/subscription-service/mappers"
	"github.com/vanjmali/spotlite/subscription-service/repositories"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrEntityNotFound  = errors.New("genre/artist couldn't be found")
	ErrInvalidEntityID = errors.New("error has ocurred while parsing genre/artist id")
	ErrUpstreamFailure = errors.New("error has ocurred while fetching artist/genre")
)

type SubscriptionService struct {
	sr  *repositories.SubscriptionRepository
	gcc *adapters.GrpcContentEntityGetter
	tr  trace.Tracer
}

func NewSubscriptionService(sr *repositories.SubscriptionRepository, gcc *adapters.GrpcContentEntityGetter) *SubscriptionService {
	tr := otel.Tracer("subscription-service/subscription-service")
	s := SubscriptionService{sr: sr, gcc: gcc, tr: tr}

	return &s
}

func (s *SubscriptionService) Subscribe(req *dtos.CreateSubscriptionDto, ctx context.Context) error {
	ctx, span := s.tr.Start(ctx, "subscription.create")
	defer span.End()

	entityExistenceCtx, entityExistenceSpan := s.tr.Start(ctx, "subscription.create.entity_exists")
	defer entityExistenceSpan.End()

	entityName, err := s.gcc.GetEntity(entityExistenceCtx, req.EntityID, subscription.SubscriptionType(req.Type))

	if err != nil {
		entityExistenceSpan.RecordError(err)

		st, ok := status.FromError(err)
		if !ok {
			return err
		}

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

	createCtx, createSpan := s.tr.Start(ctx, "subscription.create.save_subscription")
	defer createSpan.End()

	if err := s.sr.Create(se, createCtx); err != nil {
		createSpan.RecordError(err)
		return err
	}

	return nil
}
