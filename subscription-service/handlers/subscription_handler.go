package handlers

import (
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/vanjmali/spotlite/common-lib/logging"
	"github.com/vanjmali/spotlite/common-lib/requests"
	"github.com/vanjmali/spotlite/common-lib/respond"
	"github.com/vanjmali/spotlite/subscription-service/dtos"
	"github.com/vanjmali/spotlite/subscription-service/mappers"
	"github.com/vanjmali/spotlite/subscription-service/repositories"
	"github.com/vanjmali/spotlite/subscription-service/services"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SubscriptionHandler struct {
	s *services.SubscriptionService
	v *validator.Validate
}

func NewSubscriptionHandler(s services.SubscriptionService, v validator.Validate) *SubscriptionHandler {
	h := SubscriptionHandler{s: &s, v: &v}

	return &h
}

func (h *SubscriptionHandler) HandleIsSubscribed(w http.ResponseWriter, r *http.Request) {
	qEntityIDstr := mux.Vars(r)["entityID"]

	entityID, err := primitive.ObjectIDFromHex(qEntityIDstr)
	if err != nil {
		respond.BadRequest(w)
		return
	}

	err = h.s.IsSubscribed(entityID, r.Context())
	if err != nil {
		switch {
		case errors.Is(err, repositories.ErrSubscriptionNotFound):
			respond.NotFound(w)
			return
		default:
			respond.InternalServerError(w)
			return
		}
	}

	respond.Ok(w, "subscription exists")
}

func (h *SubscriptionHandler) HandleSubscribe(w http.ResponseWriter, r *http.Request) {
	var req dtos.CreateSubscriptionDto

	if ok, err := requests.ReadAndValidateJson(w, h.v, r.Body, &req); !ok {
		if err != nil {
			logging.Warnf(r.Context(), "failed to process subscribe request: %v", err)
			_ = respond.BadRequest(w, respond.ErrorMessage("invalid request body"))
		}
		return
	}

	if err := h.s.Subscribe(&req, r.Context()); err != nil {
		switch {
		case errors.Is(err, mappers.ErrSubscriptionMapping):
			logging.Warnf(r.Context(), "failed to process subscribe request: %v", err)
			_ = respond.BadRequest(w, respond.ErrorMessage(err.Error()))
			return
		case errors.Is(err, services.ErrEntityNotFound):
			logging.Warnf(r.Context(), "failed to process subscribe request: %v", err)
			_ = respond.NotFound(w)
			return
		case errors.Is(err, repositories.ErrSubscriptionAlreadyExists):
			logging.Warnf(r.Context(), "failed to process subscribe request: %v", err)
			_ = respond.Conflict(w, respond.ErrorMessageWithCode("Subscription already exists for this entity.", "subscription_exists"))
			return
		default:
			logging.Errorf(r.Context(), "failed to create subscription: %v", err)
			_ = respond.InternalServerError(w)
			return
		}
	}

	respond.NoContent(w)
}

func (h *SubscriptionHandler) HandleUnsubscribe(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	entityIdStr := vars["entityID"]

	entityId, err := primitive.ObjectIDFromHex(entityIdStr)
	if err != nil {
		logging.Warnf(r.Context(), "failed to process unsubscribe request: %v", err)
		_ = respond.BadRequest(w, respond.ErrorMessage("Invalid entity ID."))
		return
	}

	err = h.s.Unsubscribe(entityId, r.Context())
	if err != nil {
		switch {
		case errors.Is(err, services.ErrSubscriptionNotFound):
			logging.Warnf(r.Context(), "failed to process unsubscribe request: %v", err)
			_ = respond.NotFound(w)
		default:
			logging.Errorf(r.Context(), "failed to process unsubscribe request: %v", err)
			_ = respond.InternalServerError(w)
		}
		return
	}

	respond.NoContent(w)
}
