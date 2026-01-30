package handlers

import (
	"errors"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/vanjmali/spotlite/common-lib/requests"
	"github.com/vanjmali/spotlite/common-lib/respond"
	"github.com/vanjmali/spotlite/common-lib/telemetry"
	"github.com/vanjmali/spotlite/subscriptions/dtos"
	"github.com/vanjmali/spotlite/subscriptions/mappers"
	"github.com/vanjmali/spotlite/subscriptions/services"
)

type SubscriptionHandler struct {
	s *services.SubscriptionService
	v *validator.Validate
}

func NewSubscriptionHandler(s services.SubscriptionService, v validator.Validate) *SubscriptionHandler {
	h := SubscriptionHandler{s: &s, v: &v}

	return &h
}

func (h *SubscriptionHandler) HandleSubscribe(w http.ResponseWriter, r *http.Request) {
	var req dtos.CreateSubscriptionDto

	if ok, err := requests.ReadAndValidateJson(w, h.v, r.Body, &req); !ok {
		if err != nil {
			log.Printf("trace_id=%s failed to process subscribe request: %v", telemetry.TraceID(r.Context()), err)
			_ = respond.BadRequest(w, "invalid request body")
		}
		return
	}

	if err := h.s.Subscribe(&req, r.Context()); err != nil {
		switch {
		case errors.Is(err, mappers.ErrSubscriptionMapping):
			log.Printf("trace_id=%s failed to process subscribe request: %v", telemetry.TraceID(r.Context()), err)
			_ = respond.BadRequest(w, err.Error())
		case errors.Is(err, services.ErrEntityNotFound):
			log.Printf("trace_id=%s failed to process subscribe request: %v", telemetry.TraceID(r.Context()), err)
			_ = respond.UnprocessableEntity(w, err.Error())
		default:
			log.Printf("trace_id=%s failed to create subscription: %v", telemetry.TraceID(r.Context()), err)
			_ = respond.InternalServerError(w)
			return
		}
	}

	respond.NoContent(w)
}
