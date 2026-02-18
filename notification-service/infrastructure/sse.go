package infrastructure

import (
	"context"

	"github.com/vanjmali/spotlite/common-lib/logging"
)

type ClientAction int

const (
	ClientConnect ClientAction = iota
	ClientDisconnect
)

type ClientEvent struct {
	Action ClientAction
	Conn   *ClientConnection
}

type Broker struct {
	// Broadcast is a channel storing notifications that are queued to be sent out
	Broadcast chan *Notification

	ConnectionEvents chan ClientEvent

	// clients represents a map which has a UserID string as its key and a set
	// storing all connections for a particular user, by doing this we make sure
	// users that have multiple tabs opened get the updates on every single one
	// of them
	//
	// It has an empty struct as its value because it uses up 0 bytes of memory
	// and technically allows us to use a map as a set
	clients map[string]map[*ClientConnection]struct{}
}

func NewBroker() *Broker {
	return &Broker{
		// buffered channel, when a component e.g. a service tries to queue
		// a new notification it now has an extra space and it doesn't have
		// to wait for the channel to be empty as it would using an unbuffered
		// channel. It helps us avoid blocking the service handling new requests.
		//
		// The capacity is currently 1 just for testing purposes
		Broadcast: make(chan *Notification, 1),

		// buffered channel which allows us to handle bursts of client connection operations
		// instead of processing one by one (by using unbuffered channels) which
		// would cause pile ups
		ConnectionEvents: make(chan ClientEvent, 100),

		clients: make(map[string]map[*ClientConnection]struct{}),
	}
}

type ClientConnection struct {
	UserID  string
	Channel chan []byte
}

func NewClientConnection(userID string, channel chan []byte) *ClientConnection {
	return &ClientConnection{
		UserID:  userID,
		Channel: channel,
	}
}

type Notification struct {
	TargetUserID string
	Content      []byte
}

func NewNotification(targetUserID string, content []byte) *Notification {
	return &Notification{
		TargetUserID: targetUserID,
		Content:      content,
	}
}

// Constantly listens for new connections, closing connections and for notifications that have to be sent.
func (b *Broker) Listen(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			logging.Infof(ctx, "broker shutting down")
			return
		case event := <-b.ConnectionEvents:
			switch event.Action {
			case ClientConnect:
				if _, exists := b.clients[event.Conn.UserID]; !exists {
					b.clients[event.Conn.UserID] = make(map[*ClientConnection]struct{})
				}
				b.clients[event.Conn.UserID][event.Conn] = struct{}{}

			case ClientDisconnect:
				if userConns, exists := b.clients[event.Conn.UserID]; exists {
					delete(userConns, event.Conn)
					if len(userConns) == 0 {
						delete(b.clients, event.Conn.UserID)
					}
				}
			}

		case notification := <-b.Broadcast:
			if userConns, found := b.clients[notification.TargetUserID]; found {
				for clientConn := range userConns {
					select {
					case clientConn.Channel <- notification.Content:
					default:
						// if he has no available spots in the channel, we drop the notification so we keep
						// the connection alive
					}
				}
			} else {
				logging.Warnf(ctx, "notification cannot be sent because the user doesn't have an active connection")
			}
		}
	}
}
