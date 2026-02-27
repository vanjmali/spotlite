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

// ToUserActivityHistoryResponseDto converts UserActivityHistory entity to DTO
func ToUserActivityHistoryResponseDto(history *entities.UserActivityHistory) dtos.UserActivityHistoryResponseDto {
	if history == nil {
		return dtos.UserActivityHistoryResponseDto{}
	}

	activities := make([]dtos.ActivitySummaryDto, len(history.Activities))
	for i, activity := range history.Activities {
		activities[i] = dtos.ActivitySummaryDto{
			ActivityType: activity.ActivityType,
			Timestamp:    activity.Timestamp,
		}
	}

	return dtos.UserActivityHistoryResponseDto{
		UserID:     history.UserID,
		Activities: activities,
	}
}
