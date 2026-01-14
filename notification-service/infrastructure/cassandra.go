package infrastructure

import (
	"fmt"
	"log"
	"time"

	"github.com/gocql/gocql"
)

// initializes a session type which is used as an API to query the database
func Initialize(host string, keyspace string) (*gocql.Session, error) {
	cluster := gocql.NewCluster(host)

	// defining the default keyspace (db) so we avoid keyspace.table for every query
	cluster.Keyspace = keyspace

	// tells cassandra how many nodes must acknowledge a read or write for it to be considered successful,
	// quorum means that the majority has to approve for an operation to be commited (e.g. 2 out of 3 nodes)
	cluster.Consistency = gocql.Quorum

	// defines the binary protocol used to talk to the server
	cluster.ProtoVersion = 4

	// open 5 TCP connections per host (default is 2) to increase throughput
	cluster.NumConns = 5

	// max wait time for a query to be executed
	cluster.Timeout = 5 * time.Second

	// max wait time for a service to establish a connection with a node
	cluster.ConnectTimeout = 5 * time.Second

	// configuring resilience, cassandra will retry a query 3 times before returning an error
	cluster.RetryPolicy = &gocql.SimpleRetryPolicy{NumRetries: 3}

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}

	return session, nil
}

func CreateKeyspace(host string, keyspace string) error {
	cluster := gocql.NewCluster(host)
	cluster.Keyspace = "system"
	cluster.Consistency = gocql.Quorum
	cluster.ProtoVersion = 4
	cluster.Timeout = 5 * time.Second
	cluster.ConnectTimeout = 5 * time.Second

	session, err := cluster.CreateSession()
	if err != nil {
		return err
	}
	defer session.Close()

	query := fmt.Sprintf(`
		CREATE KEYSPACE IF NOT EXISTS "%s" 
		WITH replication = {
			'class': 'SimpleStrategy', 
			'replication_factor': 1
		};`, keyspace)

	log.Printf("Bootstrapping: Creating keyspace '%s' if not exists...", keyspace)
	return session.Query(query).Exec()
}
