package consumers

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/require"
	"github.com/vanjmali/spotlite/analytics-service/entities"
	"github.com/vanjmali/spotlite/common-lib/events"
	"github.com/vanjmali/spotlite/common-lib/subscription"
)

var errMetadataUnsupported = errors.New("metadata not supported in fake message")

type fakeJetStreamMsg struct {
	data []byte
}

func (m fakeJetStreamMsg) Metadata() (*jetstream.MsgMetadata, error) {
	return nil, errMetadataUnsupported
}
func (m fakeJetStreamMsg) Data() []byte                     { return m.data }
func (m fakeJetStreamMsg) Headers() nats.Header             { return nil }
func (m fakeJetStreamMsg) Subject() string                  { return "" }
func (m fakeJetStreamMsg) Reply() string                    { return "" }
func (m fakeJetStreamMsg) Ack() error                       { return nil }
func (m fakeJetStreamMsg) DoubleAck(context.Context) error  { return nil }
func (m fakeJetStreamMsg) Nak() error                       { return nil }
func (m fakeJetStreamMsg) NakWithDelay(time.Duration) error { return nil }
func (m fakeJetStreamMsg) InProgress() error                { return nil }
func (m fakeJetStreamMsg) Term() error                      { return nil }
func (m fakeJetStreamMsg) TermWithReason(string) error      { return nil }

type fakeAnalyticsService struct {
	oldRating int
	rating    int
	userID    string
	eventType string
}

func (f *fakeAnalyticsService) StoreEvent(context.Context, *entities.Event) error { return nil }

func (f *fakeAnalyticsService) ProjectSongPlayedEvent(context.Context, string, string, string) error {
	return nil
}

func (f *fakeAnalyticsService) ProjectRatingEvent(_ context.Context, userID string, eventType string, rating int, oldRating int) error {
	f.userID = userID
	f.eventType = eventType
	f.rating = rating
	f.oldRating = oldRating
	return nil
}

func (f *fakeAnalyticsService) ProjectSubscriptionEvent(context.Context, string, string, subscription.SubscriptionType) error {
	return nil
}

func TestHandleRatingUpdatedUsesPayloadOldRating(t *testing.T) {
	payload := events.RatingEventPayload{
		UserID:    "user-1",
		SongID:    "song-1",
		Rating:    5,
		OldRating: 3,
		CreatedAt: time.Now(),
	}
	b, err := json.Marshal(payload)
	require.NoError(t, err)

	service := &fakeAnalyticsService{}
	consumer := NewConsumer(service)

	err = consumer.HandleRatingUpdated(context.Background(), fakeJetStreamMsg{data: b})
	require.NoError(t, err)
	require.Equal(t, "user-1", service.userID)
	require.Equal(t, entities.EventTypeRatingUpdated, service.eventType)
	require.Equal(t, 5, service.rating)
	require.Equal(t, 3, service.oldRating)
}
