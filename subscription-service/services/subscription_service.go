package services

import (
	"context"
	"errors"

	"github.com/vanjmali/spotlite/common-lib/middlewares"
	"github.com/vanjmali/spotlite/subscriptions/dtos"
	"github.com/vanjmali/spotlite/subscriptions/entities"
	adapters "github.com/vanjmali/spotlite/subscriptions/infrastructure/grpc"
	"github.com/vanjmali/spotlite/subscriptions/mappers"
	"github.com/vanjmali/spotlite/subscriptions/repositories"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var ErrEntityNotFound = errors.New("genre/artist couldn't be found")

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

	entityName, err := s.gcc.GetEntity(entityExistenceCtx, req.EntityID, entities.SubscriptionType(req.Type))

	// TODO handle different error types differently
	if err != nil {
		entityExistenceSpan.RecordError(err)
		return ErrEntityNotFound
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
