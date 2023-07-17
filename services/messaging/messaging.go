package messaging

import (
	"context"

	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/streamdal/natty"
	"github.com/streamdal/snitch-protos/build/go/protos"
)

const (
	StreamName    = "snitch"
	StreamSubject = "events"
)

type IMsg interface {
	RunConsumer() error
	BroadcastRegistration(ctx context.Context, req *protos.RegisterRequest) error
	BroadcastCommand(ctx context.Context, cmd *protos.CommandResponse) error
	BroadcastDeregistration(req *protos.RegisterRequest) error
}

type Msg struct {
	NATS natty.INatty
	ctx  context.Context
	log  *logrus.Entry
}

func New(ctx context.Context, natsBackend natty.INatty) (*Msg, error) {
	if err := prepareNATS(ctx, natsBackend); err != nil {
		return nil, errors.Wrap(err, "error creating consumer")
	}

	return &Msg{
		NATS: natsBackend,
		ctx:  ctx,
		log:  logrus.WithField("pkg", "messaging"),
	}, nil
}

// TODO: Figure out node name -- needed for consumer bits
func prepareNATS(ctx context.Context, natsBackend natty.INatty) error {
	// Won't error if stream already exists
	if err := natsBackend.CreateStream(ctx, StreamName, []string{StreamSubject}); err != nil {
		return errors.Wrap(err, "error creating stream")
	}

	// Won't error if consumer already exists
	if err := natsBackend.CreateConsumer(ctx, StreamName, "foo", StreamSubject); err != nil {
		return errors.Wrap(err, "error creating consumer")
	}

	return nil
}

// RunConsumer is used for consuming message from the snitch NATS stream and
// executing a message handler.
func (m *Msg) RunConsumer() error {
	for {
		err := m.NATS.Consume(m.ctx, &natty.ConsumerConfig{
			Subject:      StreamSubject,
			StreamName:   StreamName,
			ConsumerName: "foo", // TODO: generate a unique consumer name
		}, m.handler)
		if err != nil {
			if err == context.Canceled {
				m.log.Debug("context cancellation detected")
				break
			}

			m.log.WithError(err).Error("error consuming messages")
		}
	}

	m.log.Debug("msg consumer exiting")

	return nil
}

// TODO: Implement
func (m *Msg) BroadcastRegistration(ctx context.Context, req *protos.RegisterRequest) error {
	m.log.Debugf("broadcasting registration: %v", req)
	return nil
}

// TODO: Implement
func (m *Msg) BroadcastCommand(ctx context.Context, cmd *protos.CommandResponse) error {
	m.log.Debugf("broadcasting command: %v", cmd)
	return nil
}

// TODO: Implement
func (m *Msg) BroadcastDeregistration(req *protos.RegisterRequest) error {
	m.log.Debugf("broadcasting deregistration: %v", req)
	return nil
}

// handler is the handler that is used for deciding what to do with a message
// consumed from the snitch NATS stream. In other words, when something is
// broadcast across the snitch stream - this handler will be executed;
// *nats.Msg.Data will contain the protos.ResponseCommand payload.
// TODO: Implement
func (m *Msg) handler(ctx context.Context, msg *nats.Msg) error {
	m.log.WithField("msg", string(msg.Data)).Info("Received message")
	return nil
}
