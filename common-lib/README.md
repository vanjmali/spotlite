# Spotlite - common-lib

Shared Go utilities and helpers used across microservices.

## Packages

- `account` - shared account types (roles, statuses).
- `middlewares` - JWT auth + RBAC helpers.
- `requests` - JSON decoding + validation helpers.
- `respond` - consistent JSON success/error responses.
- `telemetry` - OpenTelemetry tracing helpers.
- `utils` - environment helpers and JWT key loading.

## Usage examples

### Requests + validation

```go
val := validator.New()
_ = requests.RegisterValidation(val, requests.CustomValidator{
	Tag:          "strong_password",
	Func:         validation.CheckStrongPassword,
	ErrorMessage: func(_ validator.FieldError) string { return "Password too weak" },
})

var req dtos.UserLoginDto
if ok, err := requests.ReadAndValidateJson(w, val, r.Body, &req); !ok {
	if err != nil {
		log.Printf("failed to process request: %v", err)
	}
	return
}
```

### Responses

```go
_ = respond.OkJson(w, map[string]any{"message": "ok"})
_ = respond.ValidationError(w, map[string]string{"email": "Invalid email"})
```

### Auth + RBAC middleware

```go
r := mux.NewRouter()
r.Use(middlewares.ValidateJWT)
r.Handle("/admin", middlewares.ValidatePermission(account.RoleAdmin)(adminHandler))
```

### Telemetry (tracing)

Init tracing once on startup:

```go
shutdown, err := telemetry.Init(ctx, "user-service")
if err != nil {
	log.Fatalf("failed to init telemetry: %v", err)
}
defer shutdown(ctx)
```

Attach tracing to Gorilla Mux routes:

```go
r := mux.NewRouter()
telemetry.AttachMuxTracing(r, "user-service")
```

Outbound HTTP tracing:

```go
client := telemetry.NewHTTPClient()
req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
resp, err := client.Do(req)
```

Trace ID access (for logs/debug headers):

```go
traceID := telemetry.TraceID(r.Context())
```

### Utils

```go
port := utils.GetEnv("APP_PORT", "3000")
pubKey, _ := utils.GetPublicKey()
```

## Telemetry requirements

Set `OTEL_EXPORTER_OTLP_ENDPOINT` (example: `http://otel-collector:4318`) in the service environment.
If `telemetry.Init` gets an empty service name, it falls back to `OTEL_SERVICE_NAME` or `unknown-service`.
