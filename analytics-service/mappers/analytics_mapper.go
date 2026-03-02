package mappers

import (
	"sort"

	"github.com/vanjmali/spotlite/analytics-service/dtos"
	"github.com/vanjmali/spotlite/analytics-service/entities"
)

// ToUserAnalyticsResponseDto converts UserAnalyticsReadModel entity to DTO
func ToUserAnalyticsResponseDto(analytics *entities.UserAnalyticsReadModel) dtos.UserAnalyticsResponseDto {
	if analytics == nil {
		return dtos.UserAnalyticsResponseDto{}
	}

	topArtists := make([]dtos.TopArtistDto, 0, len(analytics.TopArtists))
	for artistID, playCount := range analytics.TopArtists {
		topArtists = append(topArtists, dtos.TopArtistDto{
			ArtistID:  artistID,
			PlayCount: playCount,
		})
	}

	sort.Slice(topArtists, func(i, j int) bool {
		if topArtists[i].PlayCount == topArtists[j].PlayCount {
			return topArtists[i].ArtistID < topArtists[j].ArtistID
		}
		return topArtists[i].PlayCount > topArtists[j].PlayCount
	})

	if len(topArtists) > 5 {
		topArtists = topArtists[:5]
	}

	return dtos.UserAnalyticsResponseDto{
		TotalSongsPlayed:       analytics.TotalSongsPlayed,
		AverageRating:          analytics.AverageRating,
		SongsByGenre:           analytics.SongsByGenre,
		TopArtists:             topArtists,
		SubscribedArtistsCount: analytics.SubscribedArtistsCount,
	}
}
