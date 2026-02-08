package events

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type JetStreamClient struct {
	nc     *nats.Conn
	js     jetstream.JetStream
	tracer trace.Tracer
}

func NewClient(url string) (*JetStreamClient, error) {
	// initialize NATS connection
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}

	// initialize JetStream on top of previously defined NATS connection
	js, err := jetstream.New(nc)
	if err != nil {
		return nil, err
	}

	return &JetStreamClient{
		nc:     nc,
		js:     js,
		tracer: otel.Tracer("nats-jetstream"),
	}, nil
}

// EnsureStream function allows services (both subscribers and publishers) to initialize a stream, it is
// going to be done only once by the service which gets up first and relies on this stream.
func (c *JetStreamClient) EnsureStream(ctx context.Context, streamName string, subjects []string) error {
	// tries fetching the stream
	_, err := c.js.Stream(ctx, streamName)
	// if stream already exists there won't be errors and we want to stop right there
	if err == nil {
		return nil
	}

	// check if the stream didn't exist or it was any other type of error
	if errors.Is(err, jetstream.ErrStreamNotFound) {
		// the stream isn't assigned anywhere and is accessed via the jetstream instance
		_, err = c.js.CreateStream(ctx, jetstream.StreamConfig{
			Name:     streamName,
			Subjects: subjects,
			Storage:  jetstream.FileStorage,
		})
	}

	return err
}

// Publish function is used by.
func (c *JetStreamClient) Publish(ctx context.Context, subject string, payload interface{}) error {
	// start new span of kind SpanKindProducer which is used when messages are sent to a message queue
	ctx, span := c.tracer.Start(ctx, "publish "+subject, trace.WithSpanKind(trace.SpanKindProducer))
	defer span.End()

	// serialize the payload
	data, err := json.Marshal(payload)
	if err != nil {
		span.RecordError(err)
		return err
	}

	// initialize NATS message header
	msg := &nats.Msg{
		Subject: subject,
		Data:    data,
		Header:  nats.Header{},
	}

	// inject trace in NATS message header
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(msg.Header))

	// publish message
	_, err = c.js.PublishMsg(ctx, msg)
	if err != nil {
		// record the publishing error
		span.RecordError(err)
	}
	return err // return the publishing error (if any) or return nil if everything went successfully
}

// SubscriberHandler function signature.
type SubscribeHandler func(ctx context.Context, msg jetstream.Msg) error

func (c *JetStreamClient) StartConsumer(
	ctx context.Context,
	streamName string,
	subject string,
	durableName string,
	handler SubscribeHandler,
) error {
	// if a consumer doesn't exist we initialize a new one, if it does exist (if possible) we update it
	consumer, err := c.js.CreateOrUpdateConsumer(ctx, streamName, jetstream.ConsumerConfig{
		Durable:       durableName,
		FilterSubject: subject,
		AckPolicy:     jetstream.AckExplicitPolicy,
	})
	if err != nil {
		return fmt.Errorf("failed to create consumer: %w", err)
	}

	// initialize a loop which will constantly run and check for a shutdown signal or new messages
	for {
		select {
		case <-ctx.Done():
			log.Printf("Pull consumer for %s stopped.", durableName)
			return ctx.Err()
		default:
			msgs, err := consumer.Fetch(20, jetstream.FetchMaxWait(time.Second*2))

			if err != nil && !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
				log.Printf("Fetch error: %v", err)

				// wait before retrying on critical error
				time.Sleep(time.Second)
				continue
			}

			var wg sync.WaitGroup
			msgCount := 0

			for msg := range msgs.Messages() {
				msgCount += 1
				wg.Add(1)

				go func(m jetstream.Msg) {
					defer wg.Done()
					c.processMessage(m, handler)
				}(msg)

			}

			wg.Wait()

			// if there were no messages processed by the consumer, wait for 100 ms so that the consumer avoids
			// becoming a hot loop and overloads the NATS server
			if msgCount == 0 {
				time.Sleep(time.Millisecond * 100)
			}

		}
	}
}

// processMessage is a private function which is used to process the message and send a signal to the message queue
// based on the operation result (NAK for failure, ACK for success)
func (c *JetStreamClient) processMessage(msg jetstream.Msg, handler SubscribeHandler) {
	parentCtx := otel.GetTextMapPropagator().Extract(context.Background(), propagation.HeaderCarrier(msg.Headers()))
	spanCtx, span := c.tracer.Start(parentCtx, "process "+msg.Subject(), trace.WithSpanKind(trace.SpanKindConsumer))
	defer span.End()

	err := handler(spanCtx, msg)

	if err != nil {
		log.Printf("Error processing message: %v", err)
		_ = msg.Nak() // tell NATS to resend later
		span.RecordError(err)
	} else {
		_ = msg.Ack() // tell NATS we are done
	}
}

func (c *JetStreamClient) Close() {
	c.nc.Close()
}
