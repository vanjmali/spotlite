package services

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/vanjmali/spotlite/content/dtos"
	"github.com/vanjmali/spotlite/content/entities"
)

type fakeGenreService struct {
	res *dtos.GenreListResponseDto
	err error
}

func (f *fakeGenreService) GetGenres(ctx context.Context, q GenresQuery) (*dtos.GenreListResponseDto, error) {
	return f.res, f.err
}

type fakeAlbumService struct {
	res *dtos.AlbumListResponseDto
	err error
}

func (f *fakeAlbumService) GetAlbums(ctx context.Context, q AlbumsQuery) (*dtos.AlbumListResponseDto, error) {
	return f.res, f.err
}

type fakeSongService struct {
	res *dtos.SongListResponseDto
	err error
}

func (f *fakeSongService) GetSongs(ctx context.Context, q SongsQuery) (*dtos.SongListResponseDto, error) {
	return f.res, f.err
}

type fakeArtistService struct {
	res *dtos.ArtistListResponseDto
	err error
}

func (f *fakeArtistService) GetArtists(ctx context.Context, q ArtistsQuery) (*dtos.ArtistListResponseDto, error) {
	return f.res, f.err
}

func TestGlobalSearchAllSuccess(t *testing.T) {
	genreItems := []entities.Genre{{Name: "Rock"}}
	albumItems := []entities.Album{{Title: "Hits"}}
	songItems := []entities.Song{{Title: "Song A"}}
	artistItems := []entities.Artist{{Name: "Artist A"}}

	svc := NewGlobalSearchService(
		&fakeGenreService{res: &dtos.GenreListResponseDto{Items: genreItems}},
		&fakeSongService{res: &dtos.SongListResponseDto{Items: songItems}},
		&fakeAlbumService{res: &dtos.AlbumListResponseDto{Items: albumItems}},
		&fakeArtistService{res: &dtos.ArtistListResponseDto{Items: artistItems}},
	)

	res, err := svc.GetGlobalSearch(context.Background(), "rock")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !reflect.DeepEqual(res.Genres, genreItems) {
		t.Fatalf("expected genres %v, got %v", genreItems, res.Genres)
	}
	if !reflect.DeepEqual(res.Albums, albumItems) {
		t.Fatalf("expected albums %v, got %v", albumItems, res.Albums)
	}
	if !reflect.DeepEqual(res.Songs, songItems) {
		t.Fatalf("expected songs %v, got %v", songItems, res.Songs)
	}
	if !reflect.DeepEqual(res.Artists, artistItems) {
		t.Fatalf("expected artists %v, got %v", artistItems, res.Artists)
	}
}

func TestGlobalSearchPartialFailure(t *testing.T) {
	albumItems := []entities.Album{{Title: "Hits"}}
	songItems := []entities.Song{{Title: "Song A"}}
	artistItems := []entities.Artist{{Name: "Artist A"}}

	svc := NewGlobalSearchService(
		&fakeGenreService{err: errors.New("genre down")},
		&fakeSongService{res: &dtos.SongListResponseDto{Items: songItems}},
		&fakeAlbumService{res: &dtos.AlbumListResponseDto{Items: albumItems}},
		&fakeArtistService{res: &dtos.ArtistListResponseDto{Items: artistItems}},
	)

	res, err := svc.GetGlobalSearch(context.Background(), "rock")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(res.Genres) != 0 {
		t.Fatalf("expected empty genres on error, got %v", res.Genres)
	}
	if !reflect.DeepEqual(res.Albums, albumItems) {
		t.Fatalf("expected albums %v, got %v", albumItems, res.Albums)
	}
	if !reflect.DeepEqual(res.Songs, songItems) {
		t.Fatalf("expected songs %v, got %v", songItems, res.Songs)
	}
	if !reflect.DeepEqual(res.Artists, artistItems) {
		t.Fatalf("expected artists %v, got %v", artistItems, res.Artists)
	}
}

func TestGlobalSearchAllFail(t *testing.T) {
	svc := NewGlobalSearchService(
		&fakeGenreService{err: errors.New("genre down")},
		&fakeSongService{err: errors.New("song down")},
		&fakeAlbumService{err: errors.New("album down")},
		&fakeArtistService{err: errors.New("artist down")},
	)

	res, err := svc.GetGlobalSearch(context.Background(), "rock")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(res.Genres) != 0 || len(res.Albums) != 0 || len(res.Songs) != 0 || len(res.Artists) != 0 {
		t.Fatalf("expected all empty slices, got %+v", res)
	}
}
