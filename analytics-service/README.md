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

**From CONTENT_STREAM:**

<ul>
    <li><strong>content.created</strong> - Fired when artists, albums, or songs are created. Analytics extracts song creation events.</li>
    <li><strong>content.updated</strong> - Fired when content is updated. Analytics tracks updates to song metadata (genre, etc).</li>
</ul>

**From SUBSCRIPTIONS_STREAM:**

<ul>
    <li><strong>subscribers.batch.process</strong> - Fired with subscriber batch information when artists/genres are created. Used to track which users subscribed.</li>
</ul>

**Additional Events (to be emitted by other services or defined in analytics-service):**

<ul>
    <li><strong>song_played</strong> - Fired when a user plays a song (from content-service or user interaction)</li>
    <li><strong>rating_created</strong> - Fired when a user creates a rating (from rating-service)</li>
    <li><strong>rating_updated</strong> - Fired when a user updates a rating (from rating-service)</li>
    <li><strong>rating_deleted</strong> - Fired when a user deletes a rating (from rating-service)</li>
    <li><strong>subscription_created</strong> - Fired when a user subscribes to an artist or genre. Includes subscriptionType ("artist" or "genre") and targetID.</li>
    <li><strong>subscription_deleted</strong> - Fired when a subscription is cancelled. Includes subscriptionType and targetID to identify which subscription was removed.</li>
    <li><strong>song_deleted</strong> - Fired when a song is removed from the platform (from content-service)</li>
</ul>

### Subscription Types

The Analytics Service tracks two types of subscriptions via events:

- **Artist Subscriptions**: User subscribes to content from a specific artist. `subscriptionType: "artist"`, `targetID: artist_id`
- **Genre Subscriptions**: User subscribes to content from a specific genre. `subscriptionType: "genre"`, `targetID: genre_id`

Both subscription types generate `subscription_created` and `subscription_deleted` events with the type and target information included in the event data.

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
  "event_type": "string (song_played, rating_created, etc)",
  "aggregate_id": "string",
  "aggregate_type": "string",
  "data": {},
  "version": "number",
  "timestamp": "timestamp",
  "trace_id": "string",
  "span_id": "string",
  "metadata": {}
}
```

Event-specific `data` fields:

- `song_played`: `songID`, `artistID`, `albumID`, `genreID`, `durationMS`, `playedAt`
- `rating_created`: `songID`, `rating`, `createdAt`
- `rating_updated`: `songID`, `oldRating`, `newRating`, `updatedAt`
- `rating_deleted`: `songID`, `deletedRating`, `deletedAt`
- `subscription_created`: `subscriptionType`, `targetID`, `createdAt`
- `subscription_deleted`: `subscriptionType`, `targetID`, `deletedAt`
- `song_deleted`: `songID`, `deletedAt`

**Indexes:**

<ul>
    <li>default document ID index</li>
    <li>compound index on (user_id, event_type) for filtering events by user and type</li>
    <li>compound index on (aggregate_id, aggregate_type) for ordering events per aggregate</li>
    <li>timestamp index for time-range queries</li>
    <li>trace_id index for distributed tracing correlation</li>
</ul>

#### `user_analytics` Collection

Denormalized read model with aggregated user statistics

```json
{
  "_id": "ObjectID",
  "user_id": "string",
  "total_songs_played": "number",
  "average_rating": "number",
  "songs_by_genre": {"genre_id": 12, "genre_id_2": 4, ...},
  "top_artists": [{"artist_id": "string", "artist_name": "string", "play_count": 5}],
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
