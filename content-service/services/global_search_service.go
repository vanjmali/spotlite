package services

import (
	"context"
	"sync"
	"time"

	"github.com/vanjmali/spotlite/common-lib/logging"
	"github.com/vanjmali/spotlite/content/dtos"
	"github.com/vanjmali/spotlite/content/entities"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"
)

type GlobalSearchService struct {
	genreService  GenreSearcher
	songService   SongSearcher
	albumService  AlbumSearcher
	artistService ArtistSearcher
	tr            trace.Tracer
}

type GenreSearcher interface {
	GetGenres(ctx context.Context, q GenresQuery) (*dtos.GenreListResponseDto, error)
}

type AlbumSearcher interface {
	GetAlbums(ctx context.Context, q AlbumsQuery) (*dtos.AlbumListResponseDto, error)
}

type SongSearcher interface {
	GetSongs(ctx context.Context, q SongsQuery) (*dtos.SongListResponseDto, error)
}

type ArtistSearcher interface {
	GetArtists(ctx context.Context, q ArtistsQuery) (*dtos.ArtistListResponseDto, error)
}

func NewGlobalSearchService(gs GenreSearcher, ss SongSearcher, as AlbumSearcher, ars ArtistSearcher) *GlobalSearchService {
	tr := otel.Tracer("content-service/global-search-service")
	s := GlobalSearchService{
		genreService:  gs,
		songService:   ss,
		albumService:  as,
		artistService: ars,
		tr:            tr,
	}
	return &s
}

func (s *GlobalSearchService) GetGlobalSearch(ctx context.Context, searchTerm string) (*dtos.GlobalSearchResponseDto, error) {
	ctx, span := s.tr.Start(ctx, "global_search.search")
	defer span.End()

	searchCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	g, searchCtx := errgroup.WithContext(searchCtx)

	var (
		mu      sync.Mutex
		genres  []entities.Genre
		albums  []entities.Album
		songs   []entities.Song
		artists []entities.Artist
	)

	g.Go(func() error {
		_, genreSpan := s.tr.Start(searchCtx, "global_search.search_genres")
		defer genreSpan.End()

		res, err := s.genreService.GetGenres(searchCtx, GenresQuery{
			Page: 1,
			Size: 5,
			Name: searchTerm,
		})
		if err != nil {
			logging.Errorf(searchCtx, "genre search failed: %v", err)
			genreSpan.RecordError(err)
			return nil
		}
		if res != nil {
			mu.Lock()
			genres = res.Items
			mu.Unlock()
		}
		return nil
	})

	g.Go(func() error {
		_, albumSpan := s.tr.Start(searchCtx, "global_search.search_albums")
		defer albumSpan.End()

		res, err := s.albumService.GetAlbums(searchCtx, AlbumsQuery{
			Page:  1,
			Size:  5,
			Title: searchTerm,
		})
		if err != nil {
			logging.Errorf(searchCtx, "album search failed: %v", err)
			albumSpan.RecordError(err)
			return nil
		}
		if res != nil {
			mu.Lock()
			albums = res.Items
			mu.Unlock()
		}
		return nil
	})

	g.Go(func() error {
		_, songSpan := s.tr.Start(searchCtx, "global_search.search_songs")
		defer songSpan.End()

		res, err := s.songService.GetSongs(searchCtx, SongsQuery{
			Page:  1,
			Size:  5,
			Title: searchTerm,
		})
		if err != nil {
			logging.Errorf(searchCtx, "song search failed: %v", err)
			songSpan.RecordError(err)
			return nil
		}
		if res != nil {
			mu.Lock()
			songs = res.Items
			mu.Unlock()
		}
		return nil
	})

	g.Go(func() error {
		_, artistSpan := s.tr.Start(searchCtx, "global_search.search_artists")
		defer artistSpan.End()

		res, err := s.artistService.GetArtists(searchCtx, ArtistsQuery{
			Page: 1,
			Size: 5,
			Name: searchTerm,
		})
		if err != nil {
			logging.Errorf(searchCtx, "artist search failed: %v", err)
			artistSpan.RecordError(err)
			return nil
		}
		if res != nil {
			mu.Lock()
			artists = res.Items
			mu.Unlock()
		}
		return nil
	})

	_ = g.Wait()

	return &dtos.GlobalSearchResponseDto{
		Genres:  genres,
		Albums:  albums,
		Songs:   songs,
		Artists: artists,
	}, nil
}
