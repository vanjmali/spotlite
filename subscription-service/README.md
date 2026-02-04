# Subscription service

A microservice for user subscription management

## Tech stack

- Go 1.25.5
- MongoDB
- gRPC

## Development

Additional services are available for local development:

- `localhost:3000/dev/subscription-service` - serves [mongo-express](https://github.com/mongo-express/mongo-express)

## Structure

The purpose of this service is to manage user genre/artist subscriptions,
with functionalities such as:

<ul>
    <li>Subscribe</li>
    <li>Unsubscribe</li>
    <li>Get user subscription/following list</li>
    <li>Check if user is subscribed to a particular genre/artist</li>
</ul>

### The idea behind the data model and database choice

For this service we had the freedom to choose the DB on our own, so as a team
we decided to use mongoDB since it solves the problem and gives us efficient
queries when combined with indexes.

### Subscription entity structure

```json
{
  "id": "UUID",
  "entity_id": "UUID",
  "subscriber_id": "UUID",
  "subscribed_at": "timestamp",
  "sub_type": "GENRE | ARTIST",
  "entity_name": "string"
}
```

<strong>Indexes:</strong>

<ul>
    <li>default document ID index</li>
    <li>compound, unique index subscriber_id and entity_id which allows fast retrieval which will be used
    to check if the user is subscribed to a genre/ artist and efficient unsubscribing. It is worth mentioning
    that this compound index provides fast subscription retrieval by subscriber_id only but it doesn't do the
    same for entity_id. Read more <a href="https://www.mongodb.com/docs/manual/core/indexes/index-types/index-compound">here</a>.
    </li>
    <li>
    entity_id index which will allow efficient subscription list retrieval by user, which is not handled by
    the compound index but will be needed to fetch subscribers which need to get a notification when new
    content drops
    </li>
    <li>
    subscriber_id, where the data is stored by subscribed_at field so we can fetch the following list sorted
    starting from latest
    </li>

</ul>
