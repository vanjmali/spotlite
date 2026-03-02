package mappers

import (
	"github.com/vanjmali/spotlite/analytics-service/dtos"
	"github.com/vanjmali/spotlite/analytics-service/entities"
)

// ToUserAnalyticsResponseDto converts UserAnalyticsReadModel entity to DTO
func ToUserAnalyticsResponseDto(analytics *entities.UserAnalyticsReadModel) dtos.UserAnalyticsResponseDto {
	if analytics == nil {
		return dtos.UserAnalyticsResponseDto{}
	}

	topArtists := make([]dtos.TopArtistDto, len(analytics.TopArtists))
	for i, artist := range analytics.TopArtists {
		topArtists[i] = dtos.TopArtistDto{
			ArtistID:  artist.ArtistID,
			PlayCount: artist.PlayCount,
		}
	}

	return dtos.UserAnalyticsResponseDto{
		UserID:                 analytics.UserID,
		TotalSongsPlayed:       analytics.TotalSongsPlayed,
		AverageRating:          analytics.AverageRating,
		SongsByGenre:           analytics.SongsByGenre,
		TopArtists:             topArtists,
		SubscribedArtistsCount: analytics.SubscribedArtistsCount,
	}
}
