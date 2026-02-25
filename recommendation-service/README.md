# Recommendation Service

A microservice for generating personalized music recommendations based on user interactions.

## Tech Stack

- Go 1.25.5
- Neo4j

## Development

Additional services are available for local development:

- `localhost:7474` - Neo4j Browser UI (user: `neo4j`, password: `password`)
- `localhost:7687` - Neo4j Bolt protocol connection

## Structure

The purpose of this service is to manage a recommendation graph and generate
personalized music recommendations from:

- user ratings
- listening history
- user subscriptions
- content popularity signals

The service ingests events from other microservices and updates the graph accordingly.

### The Idea Behind the Data Model and Database Choice

Neo4j is used because recommendation logic is relationship-heavy across users,
songs, artists, genres, and albums. A graph model allows efficient traversal
for recommendation generation without complex multi-join query patterns.

### Graph Model Structure

The recommendation graph consists of the following nodes and relationships.

#### Nodes

- `User` (`user_id`, `username`)
- `Song` (`song_id`, `title`, `duration`)
- `Artist` (`artist_id`, `name`)
- `Genre` (`genre_id`, `name`)
- `Album` (`album_id`, `title`)

#### Relationships

- `(User)-[:RATED {rating: int}]->(Song)`
- `(User)-[:LISTENED]->(Song)`
- `(User)-[:SUBSCRIBED]->(Artist)`
- `(User)-[:SUBSCRIBED_GENRE]->(Genre)`
- `(Song)-[:BELONGS_TO]->(Album)`
- `(Song)-[:HAS_GENRE]->(Genre)`
- `(Song)-[:BY]->(Artist)`
- `(Artist)-[:HAS_GENRE]->(Genre)`
