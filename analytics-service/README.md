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
    <li>Build analytics summaries (listening time, top artists/songs, genre preferences)</li>
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
    <li><strong>SongPlayedEvent</strong> - Fired when a user plays a song (from content-service or user interaction)</li>
    <li><strong>RatingCreatedEvent</strong> - Fired when a user creates a rating (from rating-service)</li>
    <li><strong>RatingUpdatedEvent</strong> - Fired when a user updates a rating (from rating-service)</li>
    <li><strong>RatingDeletedEvent</strong> - Fired when a user deletes a rating (from rating-service)</li>
    <li><strong>SubscriptionCreatedEvent</strong> - Fired when a user subscribes to an artist or genre. Includes subscription_type ("artist" or "genre") and target_id.</li>
    <li><strong>SubscriptionDeletedEvent</strong> - Fired when a subscription is cancelled. Includes subscription_type and target_id to identify which subscription was removed.</li>
    <li><strong>SongDeletedEvent</strong> - Fired when a song is removed from the platform (from content-service)</li>
</ul>

### Subscription Types

The Analytics Service tracks two types of subscriptions via events:

- **Artist Subscriptions**: User subscribes to content from a specific artist. `subscription_type: "artist"`, `target_id: artist_id`
- **Genre Subscriptions**: User subscribes to content from a specific genre. `subscription_type: "genre"`, `target_id: genre_id`

Both subscription types generate `SubscriptionCreatedEvent` and `SubscriptionDeletedEvent` with the type and target information included in the event data.

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
  "event_type": "string (song.played, rating.created, etc)",
  "aggregate_id": "string",
  "aggregate_type": "string",
  "data": {
    "song_id": "string (for song.played events)",
    "duration_seconds": "number (for song.played events)",
    "genre": "string (for song.played events)",
    "rating_value": "number (for rating.* events)",
    "subscription_type": "string 'artist' or 'genre' (for subscription.* events)",
    "target_id": "string (for subscription.* events - artist_id or genre_id)",
    "...": "event-type specific fields"
  },
  "version": "number",
  "timestamp": "timestamp",
  "trace_id": "string",
  "span_id": "string",
  "metadata": {}
}
```

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
  "total_songs_listened": "number",
  "total_play_count": "number",
  "average_play_duration": "number",
  "total_listening_time_seconds": "number",
  "ratings_count": "number",
  "average_rating": "number",
  "top_genres": {"rock": 150, "pop": 98, ...},
  "top_5_artists": [{artist_id, name, play_count, last_played_at}],
  "top_5_songs": [{song_id, title, artist_name, play_count, user_rating}],
  "active_subscription_count": "number",
  "subscription_history": [{subscription_id, subscription_type, target_id, target_name, start_date, end_date, is_active}],
  "first_activity_date": "timestamp",
  "last_activity_date": "timestamp",
  "updated_at": "timestamp",
  "version": "number"
}
```

**Indexes:**

<ul>
    <li>user_id (primary lookup key)</li>
    <li>last_activity_date (for identifying dormant users)</li>
</ul>

#### `user_activity_history` Collection

Individual activity records for user activity timeline

```json
{
  "_id": "ObjectID",
  "user_id": "string",
  "activity_type": "string (song_played, rating_created, etc)",
  "activity_description": "string",
  "related_entity_id": "string",
  "related_entity_type": "string",
  "metadata": {},
  "occurred_at": "timestamp",
  "trace_id": "string",
  "created_at": "timestamp"
}
```

**Indexes:**

<ul>
    <li>compound index on (user_id, occurred_at) for efficient activity timeline queries</li>
    <li>trace_id for distributed tracing</li>
</ul>
