package mappers

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vanjmali/spotlite/analytics-service/entities"
)

func TestToUserAnalyticsResponseDtoSortsAndLimitsTopArtists(t *testing.T) {
	analytics := &entities.UserAnalyticsReadModel{
		TopArtists: map[string]int{
			"artist-a": 2,
			"artist-b": 10,
			"artist-c": 5,
			"artist-d": 7,
			"artist-e": 3,
			"artist-f": 1,
		},
	}

	dto := ToUserAnalyticsResponseDto(analytics)

	require.Len(t, dto.TopArtists, 5)
	require.Equal(t, "artist-b", dto.TopArtists[0].ArtistID)
	require.Equal(t, 10, dto.TopArtists[0].PlayCount)
	require.Equal(t, "artist-d", dto.TopArtists[1].ArtistID)
	require.Equal(t, 7, dto.TopArtists[1].PlayCount)
	require.Equal(t, "artist-c", dto.TopArtists[2].ArtistID)
	require.Equal(t, 5, dto.TopArtists[2].PlayCount)
	require.Equal(t, "artist-e", dto.TopArtists[3].ArtistID)
	require.Equal(t, 3, dto.TopArtists[3].PlayCount)
	require.Equal(t, "artist-a", dto.TopArtists[4].ArtistID)
	require.Equal(t, 2, dto.TopArtists[4].PlayCount)
}

func TestToUserAnalyticsResponseDtoSortsTiesByArtistID(t *testing.T) {
	analytics := &entities.UserAnalyticsReadModel{
		TopArtists: map[string]int{
			"artist-z": 4,
			"artist-a": 4,
		},
	}

	dto := ToUserAnalyticsResponseDto(analytics)

	require.Len(t, dto.TopArtists, 2)
	require.Equal(t, "artist-a", dto.TopArtists[0].ArtistID)
	require.Equal(t, "artist-z", dto.TopArtists[1].ArtistID)
}
