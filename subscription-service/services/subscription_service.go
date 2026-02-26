package services

import (
	"context"
	"errors"
	"time"

	"github.com/avast/retry-go"
	"github.com/sony/gobreaker"
	commondtos "github.com/vanjmali/spotlite/common-lib/dtos"
	"github.com/vanjmali/spotlite/common-lib/events"
	"github.com/vanjmali/spotlite/common-lib/logging"
	"github.com/vanjmali/spotlite/common-lib/middlewares"
	"github.com/vanjmali/spotlite/common-lib/pagination"
	"github.com/vanjmali/spotlite/common-lib/subscription"
	"github.com/vanjmali/spotlite/subscription-service/dtos"
	"github.com/vanjmali/spotlite/subscription-service/entities"
	"github.com/vanjmali/spotlite/subscription-service/mappers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrEntityNotFound       = errors.New("genre/artist couldn't be found")
	ErrSubscriptionNotFound = errors.New("subscription couldn't found")
	ErrInvalidEntityID      = errors.New("error has ocurred while parsing genre/artist id")
	ErrPublish              = errors.New("error has occured while publishing subscriber batch event")
	ErrObjectIdCastFailed   = errors.New("failed to convert hex to objectId")
	ErrUpstreamTimeout      = errors.New("upstream service request timed out")
	ErrUpstreamFailure      = errors.New("upstream service returned an internal error")
	ErrUpstreamUnavailable  = errors.New("upstream service is temporarily unavailable")
	ErrUpstreamThrottled    = errors.New("upstream throttled")
)

const BATCH_SIZE = 500

type SubscriptionRepository interface {
	Create(s *entities.Subscription, ctx context.Context) error
	Delete(entityID primitive.ObjectID, userID primitive.ObjectID, ctx context.Context) (int64, error)
	FindSubscriptionsByEntityID(ctx context.Context, targetIDStrs []string, batchSize int, lastID string) ([]*entities.Subscription, string, error)
	IsSubscribed(subscriberID, entityID primitive.ObjectID, ctx context.Context) error
	FindSubscriptionsByUserID(ctx context.Context, filter bson.M, skip int64, limit int64) ([]entities.Subscription, int64, error)
	FindEntitySubscriberCount(ctx context.Context, entityID primitive.ObjectID) (int64, error)
	UpdateSubscriptionsByEntityID(ctx context.Context, entityID primitive.ObjectID, entityName string) error
}

type ContentEntityGetter interface {
	GetEntity(ctx context.Context, entityID string, subType subscription.SubscriptionType) (string, error)
}

type SubscriptionService struct {
	sr  SubscriptionRepository
	gcc ContentEntityGetter
	jsc events.JetStreamClient
	tr  trace.Tracer
	cb  *gobreaker.CircuitBreaker
}

func NewSubscriptionService(sr SubscriptionRepository, gcc ContentEntityGetter, jsc events.JetStreamClient) *SubscriptionService {
	tr := otel.Tracer("subscription-service/subscription-service")

	settings := gobreaker.Settings{
		Name: "content-service",
		// defines the number of request which will be passed through when the circuit breaker is half open
		// on which we are going to decide will we keep the circuit open or close it
		MaxRequests: 3,

		// defines the time window in which the request states will be saved, when the time is up, all request
		// data is being removed
		Interval: 15 * time.Second,

		// amount of time given to the server to get back up, since the content service dependencies aren't slow
		// to start up like cassandra 10 secs is fair
		Timeout: 10 * time.Second,

		// defines the case in which the circuit will be opened, in this case if more than 10 requests have been
		// executed and more than 30% of them failed, we want to open the circuit
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 10 && failureRatio >= 0.3
			// return counts.TotalFailures >= 1
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			// testing purposes
			// fmt.Print("circuit breaker state changed: ", to.String())
		},
	}
	s := SubscriptionService{sr: sr, gcc: gcc, jsc: jsc, tr: tr, cb: gobreaker.NewCircuitBreaker(settings)}

	return &s
}

func (s *SubscriptionService) IsSubscribed(entityID primitive.ObjectID, ctx context.Context) error {
	existsCtx, existsSpan := s.tr.Start(ctx, "subscription.exists")
	defer existsSpan.End()

	userIDstr := middlewares.GetUserIdFromContext(ctx)

	subscriberID, err := primitive.ObjectIDFromHex(userIDstr)
	if err != nil {
		existsSpan.RecordError(err)
		return err
	}

	err = s.sr.IsSubscribed(subscriberID, entityID, existsCtx)
	if err != nil {
		existsSpan.RecordError(err)
		return err
	}

	return nil
}

func (s *SubscriptionService) Subscribe(req *dtos.CreateSubscriptionDto, ctx context.Context) error {
	ctx, span := s.tr.Start(ctx, "subscription.subscribe")
	defer span.End()

	entityName, err := s.cb.Execute(func() (any, error) {
		entityCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()

		entityExistenceCtx, entityExistenceSpan := s.tr.Start(entityCtx, "subscription.subscribe.entity_exists")
		defer entityExistenceSpan.End()

		name, err := s.gcc.GetEntity(entityExistenceCtx, req.EntityID, req.Type)
		if err != nil {
			return nil, err
		}
		return name, nil
	})
	if err != nil {
		if errors.Is(err, gobreaker.ErrOpenState) {
			return ErrUpstreamUnavailable
		}

		if errors.Is(err, gobreaker.ErrTooManyRequests) {
			return ErrUpstreamThrottled
		}

		st, ok := status.FromError(err)
		if !ok {
			return err
		}

		//nolint:exhaustive
		switch st.Code() {
		case codes.NotFound:
			return ErrEntityNotFound
		case codes.InvalidArgument:
			return ErrInvalidEntityID
		case codes.DeadlineExceeded:
			return ErrUpstreamTimeout
		default:
			return ErrUpstreamFailure
		}
	}

	userIDstr := middlewares.GetUserIdFromContext(ctx)

	se, err := mappers.ToSubscriptionEntity(req, userIDstr, entityName.(string))
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

func (s *SubscriptionService) NotifySubscribers(ctx context.Context, p events.EntityCreatedEventPayload) error {
	notCtx, notSpan := s.tr.Start(ctx, "subscription.notify")
	defer notSpan.End()

	loopCtx, loopSpan := s.tr.Start(notCtx, "subscription.notify.loop")
	defer loopSpan.End()

	var lastID string
	for {
		subscriptions, nextID, err := s.sr.FindSubscriptionsByEntityID(loopCtx, p.TargetIDs, BATCH_SIZE, lastID)
		if err != nil {
			return err
		}

		if len(subscriptions) == 0 {
			logging.Infof(loopCtx, "no subscriptions were found for specified target IDs")
			break
		}

		subscriberIDs := make([]string, 0, len(subscriptions))
		for _, s := range subscriptions {
			subscriberIDs = append(subscriberIDs, s.SubscriberID.Hex())
		}

		sep := events.SubscribersBatchEventPayload{
			EntityID:      p.EntityID,
			EntityName:    p.EntityName,
			EntityType:    p.EntityType,
			CreatedAt:     p.CreatedAt,
			SubscriberIDs: subscriberIDs,
			EventID:       p.EventID,
		}

		err = retry.Do(
			func() error {
				return s.jsc.Publish(loopCtx, events.SUBJECT_SUBSCRIBER_BATCH, sep)
			},
			retry.Attempts(3),
			retry.Delay(time.Second),
			retry.DelayType(retry.BackOffDelay),
			retry.Context(loopCtx),
		)
		if err != nil {
			loopSpan.RecordError(err)
			logging.Errorf(loopCtx, "failed to publish batch: %v", err)
			return err
		}

		lastID = nextID
	}

	return nil
}

type SubsQuery struct {
	Page int
	Size int
}

func (s *SubscriptionService) ListUserSubscriptions(ctx context.Context, q SubsQuery) (*dtos.SubsListResponseDto, error) {
	listCtx, listSpan := s.tr.Start(ctx, "subscription.list")
	defer listSpan.End()

	userIDstr := middlewares.GetUserIdFromContext(ctx)

	subscriberID, err := primitive.ObjectIDFromHex(userIDstr)
	if err != nil {
		listSpan.RecordError(err)
		return nil, err
	}

	findCtx, findSpan := s.tr.Start(listCtx, "subscription.list.find")
	defer findSpan.End()

	filter := bson.M{"subscriber_id": subscriberID}
	p := pagination.NewPagination(q.Page, q.Size)

	return commondtos.ListWithPagination(findCtx, p, filter, s.sr.FindSubscriptionsByUserID)
}

func (s *SubscriptionService) FindEntitySubscriberCount(ctx context.Context, entityID primitive.ObjectID) (*dtos.EntitySubscriberCountDTO, error) {
	countCtx, countSpan := s.tr.Start(ctx, "subscription.count")
	defer countSpan.End()

	c, err := s.sr.FindEntitySubscriberCount(countCtx, entityID)
	if err != nil {
		countSpan.RecordError(err)
		return nil, err
	}

	return &dtos.EntitySubscriberCountDTO{SubscriberCount: c}, nil
}

func (s *SubscriptionService) UpdateSubscriptions(ctx context.Context, p events.EntityUpdatedEventPayload) error {
	updateCtx, updateSpan := s.tr.Start(ctx, "subscriptions.update")
	defer updateSpan.End()

	entityID, err := primitive.ObjectIDFromHex(p.EntityID)
	if err != nil {
		updateSpan.RecordError(err)
		return ErrInvalidEntityID
	}

	repoCtx, repoSpan := s.tr.Start(updateCtx, "subscriptions.update.repo_update")
	defer repoSpan.End()

	err = s.sr.UpdateSubscriptionsByEntityID(repoCtx, entityID, p.EntityName)
	if err != nil {
		repoSpan.RecordError(err)
		return err
	}

	return nil
}
