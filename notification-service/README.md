# Notification service

A microservice which main goal is to persist notifications and notify users about new releases which feature their favorite artist or genre.

## Tech stack

- Go 1.25.5
- Cassandra 5.0.6

## Structure

- The purpose of this service is to:
  - Persist and send notifications which tell the user when:
    - An artist they are subscribed to releases a new album,
    - An artist they are subscribed to releases a new song,
    - A new artist belonging to a genre they are subscribed to is added to the system,

### The idea behind the data model and database choice

- After evaluating the access patterns of the notification service, we concluded why a
  wide column database is the most suitable choice. Notifications are always fetched per user,
  and each user views their notifications in a personal inbox ordered by time.

- By using `user_id` as the partition key, we ensure that all notifications for a given user
  reside in the same partition (on a small, known set of nodes), allowing the database to
  efficiently locate the relevant data without performing full-cluster scans.

- The `created_at` timestamp is used as the clustering column, which keeps notifications sorted
  chronologically within each partition. This enables efficient range queries and fast retrieval
  of the most recent notifications, making it well suited for displaying the user's inbox.

### Notification entity structure

// TODO

```

```
