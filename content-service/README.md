# Content Service

A microservice for managing music content including albums, artists, and songs.

## Tech Stack

- Go 1.25.5
- MongoDB
- OpenTelemetry for distributed tracing

## Development

Additional services are available for local development:

- `localhost:3000/dev/content-service` - serves [mongo-express](https://github.com/mongo-express/mongo-express)
- `localhost:3102` - direct connection to MongoDB
- `localhost:9870` - HDFS NameNode web UI

### E2E Test (Audio Flow)

An end-to-end test is available for the content audio pipeline:

- `content-service/e2e/audio_flow_e2e_test.go`

Run with:

```bash
cd content-service
go test -tags e2e ./e2e -run TestAudioFlowE2E -v
```

Optional base URL override (default is `http://localhost:3000/api/content`):

```bash
E2E_BASE_URL=http://localhost:3000/api/content go test -tags e2e ./e2e -run TestAudioFlowE2E -v
```

> [!NOTE]
> `content-service` derives `length_seconds` with `ffprobe`, so runtime image/environment must have `ffprobe` available.

## Structure

The purpose of this service is to handle the management and retrieval of music content entities:

- **Albums** - Collections of songs with metadata including release date and genres
- **Artists** - Individual or group performers with associated genres and descriptions
- **Songs** - Individual tracks with duration, genre, and artist associations

### The idea behind the data model and database choice

We chose MongoDB for its document-oriented nature, which is ideal for music content management:

- **Embedded relationships**: Albums and songs embed artist and song references for efficient querying without multiple joins
- **Flexible schema**: Music metadata can vary across entities (different genres, formats)
- **Query performance**: Secondary indexes on title, genre, and artist fields enable fast filtering and pagination
- **Scalability**: Document-oriented approach scales well as the music library grows

### Entity structures

#### Album Entity

```json
{
  "_id": "ObjectID",
  "title": "string",
  "release_date": "timestamp",
  "genres": ["string"],
  "songs": [
    {
      "_id": "ObjectID",
      "title": "string",
      "genre": "string",
      "length_seconds": "int",
      "artists": [...]
    }
  ],
  "artists": [
    {
      "_id": "ObjectID",
      "name": "string",
      "genres": ["string"],
      "description": "string"
    }
  ]
}
```

#### Artist Entity

```json
{
  "_id": "ObjectID",
  "name": "string (unique)",
  "genres": ["string"],
  "description": "string"
}
```

#### Song Entity

```json
{
  "_id": "ObjectID",
  "title": "string",
  "genre": "string",
  "length_seconds": "int",
  "artists": [
    {
      "_id": "ObjectID",
      "name": "string",
      "genres": ["string"],
      "description": "string"
    }
  ]
}
```

## API Endpoints

### Albums

- `POST /albums` - Create a new album
- `GET /albums/:id` - Get album by ID
- `GET /albums` - List albums with pagination and filtering
- `PATCH /albums/:id` - Update album (partial)
- `DELETE /albums/:id` - Delete album
- `POST /albums/:id/songs` - Add songs to album
- `GET /albums/:id/songs` - List songs in album
- `DELETE /albums/:id/songs/:songId` - Remove a song from album

### Artists

- `POST /artists` - Create a new artist
- `GET /artists/:id` - Get artist by ID
- `GET /artists` - List artists with pagination and filtering
- `PATCH /artists/:id` - Update artist (partial)
- `DELETE /artists/:id` - Delete artist

### Songs

- `POST /songs` - Create a new song
- `GET /songs/:id` - Get song by ID
- `GET /songs` - List songs with pagination and filtering
- `PATCH /songs/:id` - Update song (partial)
- `DELETE /songs/:id` - Delete song
