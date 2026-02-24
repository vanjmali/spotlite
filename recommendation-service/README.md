# Recommendation service

A microservice for generating personalized music recommendations based on user interactions.

## Tech stack

- Go 1.25.5
- Neo4j

## Development

Additional services are available for local development:

- `localhost:7474` - Neo4j Browser UI (user: neo4j, pass: password)
- `localhost:7687` - Neo4j Bolt protocol connection

## Structure

- The purpose of this service is to manage the recommendation graph and generate
  personalized music recommendations based on user ratings, listening history,
  subscriptions, and content popularity. The service ingests events from other
  microservices and updates the graph accordingly.

### The idea behind the data model and database choice

- Neo4j is the best choice for this service because recommendations inherently
  involve complex relationships between entities (users, songs, artists, genres).
  Graph databases excel at traversing these relationships efficiently. Neo4j's
  Cypher query language allows us to express sophisticated recommendation algorithms
  naturally. Unlike traditional databases that would require multiple joins,
  a graph database can fetch related data with minimal overhead, resulting in
  faster recommendation generation and better user experience.

### Graph model structure

The recommendation graph consists of the following nodes and relationships:

**Nodes:**

- User (user_id, username)
- Song (song_id, title, duration)
- Artist (artist_id, name)
- Genre (genre_id, name)
- Album (album_id, title)

**Relationships:**

- (User)-[:RATED {rating: int}]->(Song)
- (User)-[:LISTENED]->(Song)
- (User)-[:SUBSCRIBED]->(Artist)
- (User)-[:SUBSCRIBED_GENRE]->(Genre)
- (Song)-[:BELONGS_TO]->(Album)
- (Song)-[:HAS_GENRE]->(Genre)
- (Song)-[:BY]->(Artist)
- (Artist)-[:HAS_GENRE]->(Genre)
