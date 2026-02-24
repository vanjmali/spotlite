# Rating service

A microservice for managing user ratings for songs in the music collection.

## Tech stack

- Go 1.25.5
- MongoDB
- gRPC

## Development

Additional services are available for local development:

- `localhost:3000/dev/rating-service` - serves [mongo-express](https://github.com/mongo-express/mongo-express)
- `localhost:3104` - direct connection to MongoDB

## Structure

The purpose of this service is to manage user song ratings with functionalities such as:

<ul>
    <li>Create ratings for songs</li>
    <li>Update existing ratings</li>
    <li>Delete ratings</li>
    <li>Retrieve ratings by user</li>
    <li>Retrieve ratings by song</li>
    <li>Calculate average ratings for songs</li>
</ul>

### The idea behind the data model and database choice

We chose MongoDB for its document-oriented nature, which is ideal for managing user ratings:

- **Query flexibility**: Ratings are frequently queried by both user and song, and MongoDB's flexible indexing supports both access patterns efficiently
- **Scalability**: As ratings accumulate, MongoDB's indexing strategy scales well for aggregation queries (like average ratings)
- **Performance**: Compound indexes on user_id and song_id enable fast checks for duplicate ratings and efficient filtering

### Rating entity structure

```json
{
  "_id": "ObjectID",
  "song_id": "ObjectID",
  "user_id": "ObjectID",
  "username": "string",
  "value": "int (1-5)",
  "created_at": "timestamp",
  "is_edited": "boolean"
}
```

<strong>Indexes:</strong>

<ul>
    <li>default document ID index</li>
    <li>compound, unique index on (user_id, song_id) which prevents duplicate ratings from the same user for the same song and enables fast retrieval of user ratings per song</li>
    <li>song_id index which allows efficient aggregation queries for calculating average ratings</li>
    <li>user_id index which supports fast retrieval of all ratings from a specific user</li>
    <li>created_at index which enables chronological sorting of ratings</li>
</ul>
