package worker

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/hibiken/asynq"
	"github.com/vanjmali/spotlite/user-service/entities"
	"github.com/vanjmali/spotlite/user-service/internal/tasks"
	"github.com/vanjmali/spotlite/user-service/services"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	// the number of days until the password is expired.
	expiryThreshold = 7

	// the number of users there will be fetched at once.
	batchSize = 500

	// the maximum number of users that can be processed at once.
	maxConcurrency = 20
)

type UserWorker struct {
	ac *asynq.Client
	us *services.UserService
	ms *services.MailService
}

func NewUserWorker(client *asynq.Client, us *services.UserService, ms *services.MailService) *UserWorker {
	return &UserWorker{ac: client, us: us, ms: ms}
}

func (w *UserWorker) HandleExpiryCheck1(ctx context.Context, t *asynq.Task) error {
	log.Printf("DEBUG: HandleExpiryCheck worker started")

	var lastID string

	for {
		users, nextID, err := w.us.FindUsersForExpiryNotification(ctx, expiryThreshold, batchSize, lastID)
		if err != nil {
			return err
		}

		if len(users) == 0 {
			break
		}

		userIDs := make([]primitive.ObjectID, 0, len(users))

		for _, u := range users {
			emailTask, _ := tasks.NewSendExpiryEmailTask(u.ID.Hex(), u.Email)

			_, err := w.ac.EnqueueContext(ctx, emailTask)
			if err != nil {
				log.Printf("ERROR: Failed to enqueue user %s: %v", u.ID.Hex(), err)
				continue
			}
			userIDs = append(userIDs, u.ID)
		}

		if len(userIDs) > 0 {
			err = w.us.MarkExpiryNotificationsSentBulk(ctx, userIDs)
			if err != nil {
				log.Printf("ERROR: Failed to bulk update users in DB: %v", err)
				return err
			}
		}

		if nextID == "" {
			break
		}
		lastID = nextID
	}

	return nil
}

func (w *UserWorker) HandleExpiryCheck2(ctx context.Context, t *asynq.Task) error {
	log.Printf("DEBUG: HandleExpiryCheck worker started")
	var lastID string

	// We are defining a semaphore channel which can handle a maximum of 20 goroutines at once,
	// or in this scenario will be able to process 20 users at once so we balance the load on
	// Redis which stores the tasks.
	//
	// We are initializing the channel with an empty struct because no data will be passed through
	// it, and it will be used to track the number of in progress requests,
	sem := make(chan struct{}, maxConcurrency)

	for {
		// fetch a batch of users,
		users, nextID, err := w.us.FindUsersForExpiryNotification(ctx, expiryThreshold, batchSize, lastID)
		if err != nil {
			return err
		}

		// if there are no users that have to be handled stop,
		if len(users) == 0 {
			log.Printf("DEBUG: No users with expiring password have been found!")
			break
		}

		// The wait group keeps track of users that are being processed as a part of the batch,
		// makes sure that new batches don't get fetched and processed while the processing of
		// the current isn't finished,
		var wg sync.WaitGroup

		for _, u := range users {
			// add an instance to the wait group so it knows the user is being processed,
			wg.Add(1)
			// add an empty struct to the semaphore so it knows that one db connection is being
			// used,
			sem <- struct{}{}

			// start the process in a new goroutine, so we can move one to another user,
			go func(user *entities.User) {
				// removes one instance from the wg group when the processing is finished,
				defer wg.Done()

				// opens up a thread in the semaphore,
				defer func() { <-sem }()

				emailTask, _ := tasks.NewSendExpiryEmailTask(user.ID.Hex(), user.Email)

				// enqueues the task,
				_, err := w.ac.EnqueueContext(ctx, emailTask)
				if err != nil {
					log.Printf("Failed to enqueue: %v", err)
					return
				}

				// update the last notification received for the current user
				_ = w.us.MarkExpiryNotificationSent(ctx, user.ID)
			}(u)
		}

		// stops the line of execution until the wg is empty (until the whole batch is processed),
		wg.Wait()

		// stops the loop if there are no users left in the next batch,
		if nextID == "" {
			break
		}

		// defines the id from which we want to continue the batch fetch,
		lastID = nextID
	}

	return nil
}

func (w *UserWorker) HandleSendExpiryEmail(ctx context.Context, t *asynq.Task) error {
	log.Printf("DEBUG: HandleSendExpiryEmail worker started")
	var p tasks.SendEmailPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}

	log.Printf("DEBUG: Sending email to %s for user %s\n", p.Email, p.UserID)

	err := w.ms.SendExpiryMail(p.Email)
	if err != nil {
		return err
	}

	return nil
}
