package routers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/vanjmali/spotlite/common-lib/middlewares"
	"github.com/vanjmali/spotlite/common-lib/telemetry"
	"github.com/vanjmali/spotlite/subscription-service/handlers"
)

func HandleRequests(sh *handlers.SubscriptionHandler) http.Handler {
	r := mux.NewRouter()
	middlewares.HandleHealthz(r)

	api := r.PathPrefix("/").Subrouter()
	telemetry.AttachMuxTracing(api, "subscription-service")

	api.Handle("/", middlewares.RequireAuthenticated(sh.HandleSubscribe)).Methods("POST")
	api.Handle("/{entityID}", middlewares.RequireAuthenticated(sh.HandleUnsubscribe)).Methods("DELETE")
	api.Handle("/", middlewares.RequireAuthenticated(sh.HandleUserSubscriptionsList)).Methods("GET")

	return r
}
