package infrastructure

import (
	"fmt"
	"time"

	"github.com/gocql/gocql"
)

func createBaseCluster(host string) *gocql.ClusterConfig {
	cluster := gocql.NewCluster(host)
	// tells cassandra how many nodes must acknowledge a read or write for it to be considered successful,
	// quorum means that the majority has to approve for an operation to be commited (e.g. 2 out of 3 nodes)
	cluster.Consistency = gocql.Quorum

	// defines the binary protocol used to talk to the server
	cluster.ProtoVersion = 4

	// max wait time for a query to be executed
	cluster.Timeout = 10 * time.Second

	// max wait time for a service to establish a connection with a node
	cluster.ConnectTimeout = 10 * time.Second

	return cluster
}

// initializes a session type which is used as an API to query the database
func Initialize(host string, keyspace string) (*gocql.Session, error) {
	cluster := createBaseCluster(host)

	// defining the default keyspace (db) so we avoid keyspace.table for every query
	cluster.Keyspace = keyspace

	// open 5 TCP connections per host (default is 2) to increase throughput
	cluster.NumConns = 5

	// configuring resilience, cassandra will retry a query 3 times before returning an error
	cluster.RetryPolicy = &gocql.SimpleRetryPolicy{NumRetries: 3}

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}

	return session, nil
}

func InitializeSchema(host string, keyspace string) error {
	cluster := createBaseCluster(host)
	cluster.Keyspace = "system"

	session, err := cluster.CreateSession()
	if err != nil {
		return err
	}
	defer session.Close()

	ksQuery := fmt.Sprintf(`
		CREATE KEYSPACE IF NOT EXISTS "%s" 
		WITH replication = {
			'class': 'SimpleStrategy', 
			'replication_factor': 1
		};`, keyspace)

	if err := session.Query(ksQuery).Exec(); err != nil {
		return err
	}

	tableQuery := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS "%s".notifications (
			user_id TEXT,
			created_at TIMESTAMP,
			notification_id UUID,
			type TEXT,
			is_read BOOLEAN,
			read_at TIMESTAMP,
			message TEXT,
			PRIMARY KEY ((user_id), created_at, notification_id)
		) WITH CLUSTERING ORDER BY (created_at DESC, notification_id ASC);`, keyspace)

	return session.Query(tableQuery).Exec()
}
