package subscription

type SubscriptionType string

const (
	GenreSubscription  SubscriptionType = "GENRE"
	ArtistSubscription SubscriptionType = "ARTIST"
)

func IsValidSubscriptionType(t SubscriptionType) bool {
	switch t {
	case GenreSubscription, ArtistSubscription:
		return true
	default:
		return false
	}
}
