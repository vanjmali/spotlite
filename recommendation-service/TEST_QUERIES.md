# Neo4j Query Testing Guide

Ovaj fajl sadrži test queries za verifikaciju svih repository metoda.

## Setup test data

Prvo pokreni ove queries u Neo4j Browser (http://localhost:7474):

```cypher
// 1. Kreiraj test nodes
CREATE (u1:User {user_id: "user1", username: "Alice"})
CREATE (u2:User {user_id: "user2", username: "Bob"})
CREATE (a1:Artist {artist_id: "artist1", name: "Queen"})
CREATE (g1:Genre {genre_id: "genre1", name: "Rock"})
CREATE (alb1:Album {album_id: "album1", title: "A Night at the Opera"})
CREATE (s1:Song {song_id: "song1", title: "Bohemian Rhapsody", duration: 354})
CREATE (s2:Song {song_id: "song2", title: "Love of My Life", duration: 218})
CREATE (s3:Song {song_id: "song3", title: "We Will Rock You", duration: 122})

// 2. Kreiraj relationships
CREATE (u1)-[:RATED {rating: 5}]->(s1)
CREATE (u1)-[:RATED {rating: 4}]->(s2)
CREATE (u2)-[:RATED {rating: 5}]->(s1)
CREATE (u2)-[:RATED {rating: 4}]->(s3)
CREATE (u1)-[:LISTENED]->(s1)
CREATE (u1)-[:LISTENED]->(s2)
CREATE (u1)-[:SUBSCRIBED]->(a1)
CREATE (u1)-[:SUBSCRIBED_GENRE]->(g1)
CREATE (s1)-[:BELONGS_TO]->(alb1)
CREATE (s2)-[:BELONGS_TO]->(alb1)
CREATE (s1)-[:HAS_GENRE]->(g1)
CREATE (s2)-[:HAS_GENRE]->(g1)
CREATE (s3)-[:HAS_GENRE]->(g1)
CREATE (s1)-[:BY]->(a1)
CREATE (s2)-[:BY]->(a1)
CREATE (a1)-[:HAS_GENRE]->(g1)
```

## Test individual queries

### 1. GetUserRatings

```cypher
MATCH (u:User {user_id: "user1"})-[r:RATED]->(s:Song)
RETURN u.user_id, s.song_id, r.rating
```

**Expected:** 2 rows (song1=5, song2=4)

### 2. GetUserListenHistory

```cypher
MATCH (u:User {user_id: "user1"})-[:LISTENED]->(s:Song)
RETURN s.song_id
LIMIT 10
```

**Expected:** 2 rows (song1, song2)

### 3. GetUserArtistSubscriptions

```cypher
MATCH (u:User {user_id: "user1"})-[:SUBSCRIBED]->(a:Artist)
RETURN a.artist_id
```

**Expected:** 1 row (artist1)

### 4. GetUserGenreSubscriptions

```cypher
MATCH (u:User {user_id: "user1"})-[:SUBSCRIBED_GENRE]->(g:Genre)
RETURN g.genre_id
```

**Expected:** 1 row (genre1)

### 5. GetSongsByGenre

```cypher
MATCH (s:Song)-[:HAS_GENRE]->(g:Genre {genre_id: "genre1"})
RETURN s.song_id
LIMIT 10
```

**Expected:** 2 rows (song1, song2)

### 6. GetSongsByArtist

```cypher
MATCH (s:Song)-[:BY]->(a:Artist {artist_id: "artist1"})
RETURN s.song_id
LIMIT 10
```

**Expected:** 2 rows (song1, song2)

### 7. GetSongsByAlbum

```cypher
MATCH (s:Song)-[:BELONGS_TO]->(a:Album {album_id: "album1"})
RETURN s.song_id
```

**Expected:** 2 rows (song1, song2)

### 8. GetHighlyRatedSongs

```cypher
MATCH (s:Song)<-[r:RATED]-()
WITH s, avg(r.rating) as avg_rating, count(r) as rating_count
WHERE avg_rating >= 4.0 AND rating_count >= 1
RETURN s.song_id
ORDER BY avg_rating DESC, rating_count DESC
LIMIT 10
```

**Expected:** 2 rows (song1 avg=5.0, song2 avg=4.0)

### 9. GetRecommendedSongsForUser

```cypher
MATCH (u:User {user_id: "user1"})
OPTIONAL MATCH (u)-[:SUBSCRIBED]->(a:Artist)<-[:BY]-(s1:Song)
OPTIONAL MATCH (u)-[:SUBSCRIBED_GENRE]->(g:Genre)<-[:HAS_GENRE]-(s2:Song)
OPTIONAL MATCH (u)-[rated:RATED]->(rated_song:Song)
OPTIONAL MATCH (rated_song)-[:HAS_GENRE]->(g2:Genre)<-[:HAS_GENRE]-(s3:Song)
WHERE rated IS NOT NULL AND rated.rating >= 4
WITH u, collect(DISTINCT s1) + collect(DISTINCT s2) + collect(DISTINCT s3) AS candidate_songs
UNWIND candidate_songs AS s
WITH u, s
WHERE s IS NOT NULL AND NOT (u)-[:RATED]->(s) AND NOT (u)-[:LISTENED]->(s)
OPTIONAL MATCH (s)<-[r:RATED]-()
WITH s, avg(r.rating) AS avg_rating, count(r) AS rating_count
RETURN s.song_id, avg_rating, rating_count
ORDER BY avg_rating DESC, rating_count DESC
LIMIT 10
```

**Expected:** Trebalo bi da preporuči song3 (user1 ga nije ocenio/slušao)

### 10. GetSimilarUsers

```cypher
MATCH (u:User {user_id: "user1"})-[r1:RATED]->(s:Song)<-[r2:RATED]-(other:User)
WHERE other.user_id <> "user1" AND abs(r1.rating - r2.rating) <= 1
WITH other, count(s) as common_songs, avg(abs(r1.rating - r2.rating)) as avg_diff
WHERE common_songs >= 1
RETURN other.user_id
ORDER BY common_songs DESC, avg_diff ASC
LIMIT 10
```

**Expected:** 1 row (user2, jer oba ocenjuju song1)

### 11. GetCollaborativeRecommendations

```cypher
MATCH (u:User {user_id: "user1"})-[r1:RATED]->(s1:Song)<-[r2:RATED]-(similar:User)
WHERE similar.user_id <> "user1" AND abs(r1.rating - r2.rating) <= 1
WITH u, similar, count(s1) AS common_songs
WHERE common_songs >= 1
MATCH (similar)-[r3:RATED]->(s2:Song)
WHERE r3.rating >= 4 AND NOT (u)-[:RATED]->(s2) AND NOT (u)-[:LISTENED]->(s2)
RETURN s2.song_id, count(similar) AS similar_user_count, avg(r3.rating) AS avg_rating
ORDER BY similar_user_count DESC, avg_rating DESC
LIMIT 10
```

**Expected:** Može vratiti song3 (user2 ga je ocenio)

## Cleanup test data

Nakon testiranja, obriši test podatke:

```cypher
MATCH (n)
WHERE n.user_id IN ["user1", "user2"]
   OR n.artist_id = "artist1"
   OR n.genre_id = "genre1"
   OR n.album_id = "album1"
   OR n.song_id IN ["song1", "song2", "song3"]
DETACH DELETE n
```
