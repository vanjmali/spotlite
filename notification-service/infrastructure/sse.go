package infrastructure

import "log"

type Broker struct {
	// Broadcast is a channel storing notifications that are queued to be sent out
	Broadcast chan *Notification

	// NewClients is storing recently opened client connections
	NewClients chan *ClientConnection

	// ClosingClients is storing connections that were recently closed
	ClosingClients chan *ClientConnection

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

		// unbuffered channels, will make components that try to communicate with
		// them wait for the current connection to be processed
		NewClients:     make(chan *ClientConnection),
		ClosingClients: make(chan *ClientConnection),

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
func (b *Broker) Listen() {
	for {
		select {
		// handling new client connections,
		case s := <-b.NewClients:
			if _, exists := b.clients[s.UserID]; !exists {
				b.clients[s.UserID] = make(map[*ClientConnection]struct{})
			}

			b.clients[s.UserID][s] = struct{}{}

		// handling client closign connections,
		case s := <-b.ClosingClients:
			if userConns, exists := b.clients[s.UserID]; exists {
				delete(userConns, s)

				if len(userConns) == 0 {
					delete(b.clients, s.UserID)
				}
			}

		// handling new client connections,
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
				log.Print("[DEBUG]: The notification can't be sent because the user doesn't have an active connection!")
			}
		}
	}
}
