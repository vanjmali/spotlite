# Analytics Service

A microservice for aggregating and analyzing user listening patterns, ratings, and subscription data using Event Sourcing and CQRS architecture.

## Tech Stack

- Go 1.25.5
- MongoDB (Event Store & Denormalized Read Models)
- NATS JetStream (Event Pub/Sub via common-lib events)
- Event Sourcing + CQRS Pattern

## Development

Additional services are available for local development:

- `localhost:3000/dev/analytics-service` - serves [mongo-express](https://github.com/mongo-express/mongo-express)
- `localhost:3105` - direct connection to MongoDB

## Structure

The purpose of this service is to:

<ul>
    <li>Track all user listening activities (song plays, ratings, subscriptions)</li>
  <li>Build analytics summaries (top artists, genre preferences)</li>
    <li>Provide user activity timeline and history</li>
    <li>Support query-optimized read models for fast analytics retrieval</li>
    <li>Maintain immutable event log for audit trail and replay capability</li>
</ul>

### Architecture: Event Sourcing + CQRS

This service uses a **dual-model persistence strategy**:

- **Write Model (Event Store)**: Immutable, append-only event log stored in MongoDB `events` collection. Every state change is persisted as an event.
- **Read Model (Denormalized Collections)**: Optimized views for fast queries (`user_analytics`, `user_activity_history` collections)
- **Event Projection**: NATS JetStream consumers subscribe to events and asynchronously update read models
- **Consistency Model**: Eventual consistency - events are immediately committed to the store, read models are updated asynchronously

### Event Types

The Analytics Service consumes events from the following NATS JetStream streams (defined in common-lib/events):

**From LISTENS_STREAM:**

<ul>
    <li><strong>listen.created</strong> - Fired when a user plays a song (from content-service TrackSongPlay method)</li>
</ul>

**From RATINGS_STREAM:**

<ul>
    <li><strong>rating.created</strong> - Fired when a user creates a rating (from rating-service)</li>
    <li><strong>rating.updated</strong> - Fired when a user updates an existing rating (from rating-service)</li>
    <li><strong>rating.deleted</strong> - Fired when a user deletes a rating (from rating-service)</li>
</ul>

**From SUBSCRIPTIONS_STREAM:**

<ul>
    <li><strong>subscription.created</strong> - Fired when a user subscribes to an artist or genre (from subscription-service Subscribe method)</li>
    <li><strong>subscription.deleted</strong> - Fired when a subscription is cancelled (from subscription-service Unsubscribe method)</li>
</ul>

**Note on Song Deletion:** Song deletion is handled via the Saga pattern (requirement 2.13) coordinated through the content service, not tracked as a direct analytics event.

### Subscription Types

The Analytics Service tracks two types of subscriptions via events (using unified `SubscriptionEventPayload`):

- **Artist Subscriptions**: User subscribes to content from a specific artist. `EntityType: "ARTIST"`, `EntityID: artist_id`
- **Genre Subscriptions**: User subscribes to content from a specific genre. `EntityType: "GENRE"`, `EntityID: genre_id`

Both subscription types generate `subscription.created` and `subscription.deleted` events with the entity type and ID included in the event payload.

### The Idea Behind Event Sourcing & CQRS

**Event Sourcing** provides:

- **Immutable audit trail**: Every action is recorded permanently
- **Event replay**: Rebuild analytics at any point in time
- **Fine-grained tracing**: Full context (TraceID, SpanID) for distributed debugging
- **Current state derivation**: Analytics are computed from events, not direct state mutations

**CQRS (Command Query Responsibility Segregation)** provides:

- **Separation of concerns**: Write path (events) vs read path (queries)
- **Performance optimization**: Read models are denormalized for specific queries
- **Scalability**: Event store and read models can scale independently
- **Flexibility**: Different query patterns supported without impacting event store

### Event Projection

The Analytics Service uses NATS JetStream for event projection. Event consumers subscribe to streams and asynchronously update denormalized read models in MongoDB.

#### `events` Collection

Immutable event log - every state change in the analytics domain

```json
{
  "_id": "ObjectID",
  "user_id": "string",
  "event_type": "string (listen_created, rating_created, rating_updated, rating_deleted, subscription_created, subscription_deleted)",
  "data": {},
  "timestamp": "timestamp"
}
```

Event-specific `data` fields (from unified event payloads):

- `listen_created` (from ListenEventPayload): `song_id`, `user_id`, `event_id`, `created_at`
- `rating_created` (from RatingEventPayload): `user_id`, `song_id`, `rating`, `event_id`, `created_at`
- `rating_updated` (from RatingEventPayload): `user_id`, `song_id`, `rating`, `event_id`, `created_at`
- `rating_deleted` (from RatingEventPayload): `user_id`, `song_id`, `rating`, `event_id`, `created_at`
- `subscription_created` (from SubscriptionEventPayload): `user_id`, `entity_id`, `entity_type` (ARTIST|GENRE), `event_id`, `created_at`
- `subscription_deleted` (from SubscriptionEventPayload): `user_id`, `entity_id`, `entity_type` (ARTIST|GENRE), `event_id`, `created_at`

**Indexes:**

<ul>
    <li>default document ID index</li>
    <li>compound index on (user_id, event_type) for filtering events by user and type during read model projection</li>
    <li>timestamp index for time-range queries</li>
</ul>

#### `user_analytics` Collection

Denormalized read model with aggregated user statistics

```json
{
  "_id": "ObjectID",
  "user_id": "string",
  "total_songs_played": "number",
  "average_rating": "number",
  "rating_sum": "number",
  "ratings_count": "number",
  "songs_by_genre": {"genre_id": 12, "genre_id_2": 4, ...},
  "top_artists": [{"artist_id": "string", "play_count": 5}],
  "subscribed_artists_count": "number"
}
```

**Indexes:**

<ul>
  <li>user_id (primary lookup key)</li>
</ul>

#### `user_activity_history` Collection

Individual activity records for user activity timeline

```json
{
  "_id": "ObjectID",
  "user_id": "string",
  "activities": [
    {
      "activity_type": "string (song_played, rating_created, etc)",
      "timestamp": "timestamp"
    }
  ]
}
```

**Indexes:**

<ul>
  <li>compound index on (user_id, activities.timestamp) for activity timeline queries</li>
</ul>

## API Endpoints

All endpoints require authentication via JWT token (Bearer token in Authorization header). The API Gateway routes requests to the analytics-service at `/api/analytics/*`.

### Analytics

#### Get User Analytics

Retrieves aggregated analytics data for a specific user including listening statistics, top artists, genre preferences, and subscription counts.

```
GET /analytics/:userID
```

**Authentication:** Required (JWT Bearer token)

**Path Parameters:**

- `userID` (string, required) - The unique identifier of the user

**Response (200 OK):**

```json
{
  "user_id": "507f1f77bcf86cd799439011",
  "total_songs_played": 42,
  "average_rating": 4.2,
  "songs_by_genre": {
    "rock": 15,
    "jazz": 12,
    "pop": 8,
    "classical": 7
  },
  "top_artists": [
    {
      "artist_id": "507f191e810c19729de860ea",
      "play_count": 18
    },
    {
      "artist_id": "507f191e810c19729de860eb",
      "play_count": 12
    }
  ],
  "subscribed_artists_count": 5
}
```

**Error Responses:**

- `400 Bad Request` - Invalid userID format
- `401 Unauthorized` - Missing or invalid authentication token
- `404 Not Found` - Analytics not found for the specified user
- `500 Internal Server Error` - Server-side error

### Activity History

#### Get User Activity History

Retrieves the chronological activity timeline for a specific user showing recent listening, rating, and subscription events.

```
GET /activity-history/:userID
```

**Authentication:** Required (JWT Bearer token)

**Path Parameters:**

- `userID` (string, required) - The unique identifier of the user

**Response (200 OK):**

```json
{
  "user_id": "507f1f77bcf86cd799439011",
  "activities": [
    {
      "activity_type": "song_played",
      "timestamp": "2026-02-27T14:30:00Z"
    },
    {
      "activity_type": "rating_created",
      "timestamp": "2026-02-27T14:25:00Z"
    },
    {
      "activity_type": "subscription_created",
      "timestamp": "2026-02-27T14:20:00Z"
    }
  ]
}
```

**Activity Types:**

- `listen_created` - User listened to a song
- `rating_created` - User created a new rating
- `rating_updated` - User updated an existing rating
- `rating_deleted` - User deleted a rating
- `subscription_created` - User subscribed to an artist or genre
- `subscription_deleted` - User unsubscribed from an artist or genre

**Error Responses:**

- `400 Bad Request` - Invalid userID format
- `401 Unauthorized` - Missing or invalid authentication token
- `404 Not Found` - Activity history not found for the specified user
- `500 Internal Server Error` - Server-side error

### Notes

- Both endpoints return denormalized read models optimized for fast retrieval
- Analytics data is eventually consistent with the event store
- Activity history is limited to the most recent 1000 activities per user
- All timestamps are in ISO 8601 format (UTC)
- The `songs_by_genre` map returns only genres with at least one play
- The `top_artists` array is sorted by play count (descending) and limited to top 5 artists
