package store

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	telTypes "github.com/streamdal/server/types"

	"github.com/cactus/go-statsd-client/v5/statsd"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"

	"github.com/streamdal/protos/build/go/protos"

	"github.com/streamdal/server/services/encryption"
	"github.com/streamdal/server/services/store/types"
	"github.com/streamdal/server/util"
	"github.com/streamdal/server/validate"
)

/*

!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!

Storage strategy is defined here:

https://www.notion.so/streamdal/server-Storage-Spec-417bfa71f04b481082373ad18cbdb0e9

!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!

`store` is a service that handles storage and retrieval of data such as service
registrations and service commands.

`store` is backed by a `natty.INatty` instance, which is a wrapper for RedisBackend.

All reads, writes and deletes are performed via RedisBackend -- server does NOT
store any persistent state in memory!
*/

var (
	ErrPipelineNotFound = errors.New("pipeline not found")
	ErrConfigNotFound   = errors.New("config not found")
)

const (
	// RedisKeyWatchPrefix is the key under which redis publishes key events.
	// The format is  __keyspace@{$database_number}__
	// We're always defaulting to db 0, so we can use this prefix to watch for key changes
	// See https://redis.io/docs/manual/keyspace-notifications/
	RedisKeyWatchPrefix = "__keyspace@0__:"

	// InstallIDKey is a unique ID for this streamdal server cluster
	// Each cluster will get a unique UUID. This is used to track the number of
	// installs for telemetry and is completely random for anonymization purposes.
	InstallIDKey = "install_id"

	RedisCreationDateKey = "streamdal_settings:creation_date"
)

type IStore interface {
	AddRegistration(ctx context.Context, req *protos.RegisterRequest) error
	DeleteRegistration(ctx context.Context, req *protos.DeregisterRequest) error
	RecordRegistration(ctx context.Context, req *protos.RegisterRequest) error
	SeenRegistration(ctx context.Context, req *protos.RegisterRequest) bool
	GetPipelines(ctx context.Context) (map[string]*protos.Pipeline, error)
	GetPipeline(ctx context.Context, pipelineID string) (*protos.Pipeline, error)
	GetConfig(ctx context.Context) (map[*protos.Audience][]string, error) // v: pipeline_id
	GetConfigByAudience(ctx context.Context, audience *protos.Audience) ([]string, error)
	GetLive(ctx context.Context) ([]*types.LiveEntry, error)
	GetPaused(ctx context.Context) (map[string]*types.PausedEntry, error)
	CreatePipeline(ctx context.Context, pipeline *protos.Pipeline) error
	AddAudience(ctx context.Context, req *protos.NewAudienceRequest) error
	DeleteAudience(ctx context.Context, req *protos.DeleteAudienceRequest) error
	DeletePipeline(ctx context.Context, pipelineID string) error
	UpdatePipeline(ctx context.Context, pipeline *protos.Pipeline) error
	AttachPipeline(ctx context.Context, req *protos.AttachPipelineRequest) error
	DetachPipeline(ctx context.Context, req *protos.DetachPipelineRequest) error
	PausePipeline(ctx context.Context, req *protos.PausePipelineRequest) error
	ResumePipeline(ctx context.Context, req *protos.ResumePipelineRequest) error
	IsPaused(ctx context.Context, audience *protos.Audience, pipelineID string) (bool, error)
	GetAudiences(ctx context.Context) ([]*protos.Audience, error)
	IsPipelineAttached(ctx context.Context, audience *protos.Audience, pipelineID string) (bool, error)
	GetNotificationConfig(ctx context.Context, req *protos.GetNotificationRequest) (*protos.NotificationConfig, error)
	GetNotificationConfigs(ctx context.Context) (map[string]*protos.NotificationConfig, error)
	GetNotificationConfigsByPipeline(ctx context.Context, pipelineID string) ([]*protos.NotificationConfig, error)
	CreateNotificationConfig(ctx context.Context, req *protos.CreateNotificationRequest) error
	UpdateNotificationConfig(ctx context.Context, req *protos.UpdateNotificationRequest) error
	DeleteNotificationConfig(ctx context.Context, req *protos.DeleteNotificationRequest) error
	AttachNotificationConfig(ctx context.Context, req *protos.AttachNotificationRequest) error
	DetachNotificationConfig(ctx context.Context, req *protos.DetachNotificationRequest) error
	GetAttachCommandsByService(ctx context.Context, serviceName string) ([]*protos.Command, error)
	GetPipelineUsage(ctx context.Context) ([]*PipelineUsage, error)
	GetActivePipelineUsage(ctx context.Context, pipelineID string) ([]*PipelineUsage, error)
	GetActiveTailCommandsByService(ctx context.Context, serviceName string) ([]*protos.Command, error)
	AddActiveTailRequest(ctx context.Context, req *protos.TailRequest) (string, error) // Returns key that tail req is stored under

	// GetPipelineHistory returns the history of a pipeline as a map. The key is the version, the value is a protos.Pipeline
	GetPipelineHistory(ctx context.Context, pipelineID string) (map[int32]*protos.Pipeline, error)

	// DeletePipelineHistory deletes any history entries for the given pipeline ID
	DeletePipelineHistory(ctx context.Context, pipelineId string) error

	// StorePipelineHistory stores a pipeline history entry in redis
	StorePipelineHistory(ctx context.Context, pipeline *protos.Pipeline) error

	// GetAudiencesBySessionID returns all audiences for a given session id
	// This is needed to inject an inferschema pipeline for each announced audience
	// to the session via a goroutine in internal.Register()
	GetAudiencesBySessionID(ctx context.Context, sessionID string) ([]*protos.Audience, error)

	GetAudiencesByService(ctx context.Context, serviceName string) ([]*protos.Audience, error)

	// WatchKeys will watch for key changes for given key pattern; every time key
	// is updated, it will send a message on the return channel.
	// WatchKeys launches a dedicated goroutine for the watch - it can be stopped
	// via the provided context.
	WatchKeys(quitCtx context.Context, key string) chan struct{}

	AddSchema(ctx context.Context, req *protos.SendSchemaRequest) error

	GetSchema(ctx context.Context, aud *protos.Audience) (*protos.Schema, error)

	// GetInstallID returns the unique ID for this server cluster.
	// If an ID has not been set yet, a new one is generated and returned
	GetInstallID(ctx context.Context) (string, error)

	// GetCreationDate returns the creation date of this server cluster. This is used
	// for sending server_timestamp_created_seconds metric to telemetry
	GetCreationDate(ctx context.Context) (int64, error)

	// SetCreationDate sets the creation date of this server cluster
	SetCreationDate(ctx context.Context, ts int64) error

	// IsPipelineAttachedAny returns if pipeline is attached to any audience. Used for telemetry tags
	IsPipelineAttachedAny(ctx context.Context, pipelineID string) bool

	// PauseTailRequest pauses a tail request by its ID
	PauseTailRequest(ctx context.Context, req *protos.PauseTailRequest) (*protos.TailRequest, error)

	// ResumeTailRequest resumes a tail request by its ID
	ResumeTailRequest(ctx context.Context, req *protos.ResumeTailRequest) (*protos.TailRequest, error)

	// GetTailRequestById returns a TailRequest by its ID
	GetTailRequestById(ctx context.Context, tailID string) (*protos.TailRequest, error)
}

type Options struct {
	Encryption   encryption.IEncryption
	RedisBackend *redis.Client
	ShutdownCtx  context.Context
	NodeName     string
	SessionTTL   time.Duration
	Telemetry    statsd.Statter
	InstallID    string
}

type Store struct {
	options *Options
	log     *logrus.Entry
}

func New(opts *Options) (*Store, error) {
	if err := opts.validate(); err != nil {
		return nil, errors.Wrap(err, "error validating options")
	}

	// https://redis.io/docs/manual/keyspace-notifications/
	// Set redis to publish "set" events, we need this for heartbeat detection
	opts.RedisBackend.ConfigSet(opts.ShutdownCtx, "notify-keyspace-events", "KEAs")

	return &Store{
		options: opts,
		log:     logrus.WithField("pkg", "store"),
	}, nil
}

func (s *Store) AddRegistration(ctx context.Context, req *protos.RegisterRequest) error {
	llog := s.log.WithField("method", "AddRegistration")
	llog.Debug("received request to add registration")

	registrationKey := RedisLiveKey(req.SessionId, s.options.NodeName, "register")

	// Populate these so we can save them in K/V; needed for GetAll()
	req.ClientInfo.XSessionId = &req.SessionId
	req.ClientInfo.XServiceName = &req.ServiceName
	req.ClientInfo.XNodeName = &s.options.NodeName

	clientInfoBytes, err := proto.Marshal(req.ClientInfo)
	if err != nil {
		return errors.Wrap(err, "error marshalling client info")
	}

	s.log.Debugf("attempting to save registration under key '%s'", registrationKey)

	status := s.options.RedisBackend.Set(ctx, registrationKey, clientInfoBytes, s.options.SessionTTL)
	if err := status.Err(); err != nil {
		return errors.Wrap(err, "error adding registration to K/V")
	}

	// Save audience(s) if they are present in Register request
	if req.Audiences != nil && len(req.Audiences) > 0 {
		for _, audience := range req.Audiences {
			llog.Debugf("adding audience '%s' for session '%s'", audience, req.SessionId)

			if err := s.AddAudience(ctx, &protos.NewAudienceRequest{
				SessionId: req.SessionId,
				Audience:  audience,
			}); err != nil {
				return errors.Wrap(err, "error adding audience")
			}
		}
	}

	return nil
}

func (s *Store) SeenRegistration(ctx context.Context, req *protos.RegisterRequest) bool {
	registrationKey := RedisLiveKey(req.SessionId, s.options.NodeName, "register")

	return s.options.RedisBackend.Exists(ctx, registrationKey).Val() == 1
}

func (s *Store) RecordRegistration(ctx context.Context, req *protos.RegisterRequest) error {
	registrationKey := RedisTelemetryRegistrationKey(
		req.ServiceName,
		req.ClientInfo.Os,
		req.ClientInfo.LibraryName,
		req.ClientInfo.Arch,
	)

	req.ClientInfo.XSessionId = &req.SessionId
	req.ClientInfo.XServiceName = &req.ServiceName
	req.ClientInfo.XNodeName = &s.options.NodeName

	clientInfoBytes, err := proto.Marshal(req.ClientInfo)
	if err != nil {
		return errors.Wrap(err, "error marshalling client info")
	}

	status := s.options.RedisBackend.Set(ctx, registrationKey, clientInfoBytes, s.options.SessionTTL)
	if err := status.Err(); err != nil {
		return errors.Wrap(err, "error record permanent registration to K/V")
	}

	return nil
}

func (s *Store) DeleteRegistration(ctx context.Context, req *protos.DeregisterRequest) error {
	llog := s.log.WithField("method", "DeleteRegistration")
	llog.Debug("received request to delete registration")

	// Remove all keys referencing the same session_id
	entries, err := s.GetLive(ctx)
	if err != nil {
		return errors.Wrap(err, "error fetching live entries")
	}

	for _, e := range entries {
		if e.SessionID != req.SessionId {
			continue
		}

		// Same session id - remove the key
		if err := s.options.RedisBackend.Del(ctx, e.Key).Err(); err != nil {
			s.log.Errorf("unable to remove key '%s' from K/V", e.Key)
		}
	}

	return nil
}

func (s *Store) GetPipelines(ctx context.Context) (map[string]*protos.Pipeline, error) {
	llog := s.log.WithField("method", "GetPipelines")
	llog.Debug("received request to get pipelines")

	pipelineIds, err := s.options.RedisBackend.Keys(ctx, RedisPipelinePrefix+":*").Result()
	if err != nil {
		return nil, errors.Wrap(err, "error fetching pipeline keys from store")
	}

	// k == pipelineId
	pipelines := make(map[string]*protos.Pipeline)

	for _, pipelineId := range pipelineIds {
		pipelineData, err := s.options.RedisBackend.Get(ctx, pipelineId).Result()
		if err != nil {
			return nil, errors.Wrapf(err, "error fetching pipeline '%s' from store", pipelineId)
		}

		pipeline := &protos.Pipeline{}

		if err := proto.Unmarshal([]byte(pipelineData), pipeline); err != nil {
			return nil, errors.Wrapf(err, "error unmarshaling pipeline '%s'", pipelineId)
		}

		pipelineId = strings.TrimPrefix(pipelineId, RedisPipelinePrefix+":")
		pipelines[pipelineId] = pipeline
	}

	return pipelines, nil
}

func (s *Store) GetPipeline(ctx context.Context, pipelineId string) (*protos.Pipeline, error) {
	llog := s.log.WithField("method", "GetPipeline")
	llog.Debug("received request to get pipeline")

	pipelineData, err := s.options.RedisBackend.Get(ctx, RedisPipelineKey(pipelineId)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrPipelineNotFound
		}

		return nil, errors.Wrap(err, "error fetching pipeline from store")
	}

	pipeline := &protos.Pipeline{}

	if err := proto.Unmarshal([]byte(pipelineData), pipeline); err != nil {
		return nil, errors.Wrap(err, "error deserializing pipeline")
	}

	return pipeline, nil
}

func (s *Store) CreatePipeline(ctx context.Context, pipeline *protos.Pipeline) error {
	llog := s.log.WithField("method", "CreatePipeline")
	llog.Debug("received request to create pipeline")

	applyPipelineDefaults(pipeline)

	// Save to K/V
	pipelineData, err := proto.Marshal(pipeline)
	if err != nil {
		return errors.Wrap(err, "error serializing pipeline to protobuf")
	}

	if err := s.options.RedisBackend.Set(ctx, RedisPipelineKey(pipeline.Id), pipelineData, 0).Err(); err != nil {
		return errors.Wrap(err, "error saving pipeline to store")
	}

	return nil
}

func (s *Store) DeletePipeline(ctx context.Context, pipelineId string) error {
	llog := s.log.WithField("method", "DeletePipeline")
	llog.Debug("received request to delete pipeline")

	// Does this pipeline exist?
	if _, err := s.GetPipeline(ctx, pipelineId); err != nil {
		return errors.Wrap(err, "error fetching pipeline")
	}

	if err := s.options.RedisBackend.Del(ctx, RedisPipelineKey(pipelineId)).Err(); err != nil {
		return errors.Wrap(err, "error deleting pipeline from store")
	}

	return nil
}

func (s *Store) DeletePipelineHistory(ctx context.Context, pipelineId string) error {
	if err := s.options.RedisBackend.Del(ctx, RedisPipelineHistoryKey(pipelineId)).Err(); err != nil {
		return errors.Wrap(err, "error deleting pipeline history from store")
	}

	return nil
}

func (s *Store) UpdatePipeline(ctx context.Context, pipeline *protos.Pipeline) error {
	llog := s.log.WithField("method", "UpdatePipeline")
	llog.Debug("received request to update pipeline")

	// Save to K/V
	pipelineData, err := proto.Marshal(pipeline)
	if err != nil {
		return errors.Wrap(err, "error serializing pipeline to protobuf")
	}

	if err := s.options.RedisBackend.Set(ctx, RedisPipelineKey(pipeline.Id), pipelineData, 0).Err(); err != nil {
		return errors.Wrap(err, "error saving pipeline to store")
	}

	return nil
}

func (s *Store) AttachPipeline(ctx context.Context, req *protos.AttachPipelineRequest) error {
	llog := s.log.WithField("method", "AttachPipeline")
	llog.Debug("received request to attach pipeline")

	// Does this pipeline exist?
	if _, err := s.GetPipeline(ctx, req.PipelineId); err != nil {
		return errors.Wrap(err, "error fetching pipeline")
	}

	// Store attachment in RedisBackend
	key := RedisConfigKey(req.Audience, req.PipelineId)

	if err := s.options.RedisBackend.Set(ctx, key, []byte(``), 0).Err(); err != nil {
		return errors.Wrap(err, "error saving pipeline attachment to store")
	}

	return nil
}

func (s *Store) DetachPipeline(ctx context.Context, req *protos.DetachPipelineRequest) error {
	llog := s.log.WithField("method", "DetachPipeline")
	llog.Debug("received request to detach pipeline")

	// Does this pipeline exist?
	if _, err := s.GetPipeline(ctx, req.PipelineId); err != nil {
		if err == ErrPipelineNotFound {
			llog.Debugf("pipeline '%s' not found - nothing to do", req.PipelineId)
			return nil
		}

		return errors.Wrap(err, "error fetching pipeline")
	}

	// Delete audience association
	if err := s.options.RedisBackend.Del(ctx, RedisConfigKey(req.Audience, req.PipelineId)).Err(); err != nil {
		return errors.Wrap(err, "error deleting pipeline attachment from store")
	}

	// Delete from paused
	if err := s.options.RedisBackend.Del(ctx, RedisPausedKey(util.AudienceToStr(req.Audience), req.PipelineId)).Err(); err != nil {
		return errors.Wrap(err, "error deleting pipeline pause state")
	}

	return nil
}

func (s *Store) PausePipeline(ctx context.Context, req *protos.PausePipelineRequest) error {
	llog := s.log.WithField("method", "PausePipeline")
	llog.Debug("received request to pause pipeline")

	// Does this pipeline exist?
	if _, err := s.GetPipeline(ctx, req.PipelineId); err != nil {
		return errors.Wrap(err, "error fetching pipeline")
	}

	paused, err := s.IsPaused(ctx, req.Audience, req.PipelineId)
	if err != nil {
		return errors.Wrap(err, "error checking if pipeline is paused")
	}

	if paused {
		llog.Debugf("pipeline '%s' already paused; nothing to do", req.PipelineId)
		return nil
	}

	llog.Debugf("pipeline '%s' not paused; setting pause state now", req.PipelineId)

	if err := s.options.RedisBackend.Set(
		ctx,
		RedisPausedKey(util.AudienceToStr(req.Audience), req.PipelineId),
		[]byte(``),
		0,
	).Err(); err != nil {
		return errors.Wrap(err, "error saving pipeline pause state")
	}

	return nil
}

// IsPaused returns if pipeline is paused and if it exists
func (s *Store) IsPaused(ctx context.Context, audience *protos.Audience, pipelineID string) (bool, error) {
	llog := s.log.WithField("method", "IsPaused")
	llog.Debug("received request to check if pipeline is paused")

	if err := s.options.RedisBackend.Get(ctx, RedisPausedKey(util.AudienceToStr(audience), pipelineID)).Err(); err != nil {
		if err == redis.Nil {
			return false, nil
		}

		return false, errors.Wrap(err, "error fetching pipeline pause state")
	}

	return true, nil
}

func (s *Store) ResumePipeline(ctx context.Context, req *protos.ResumePipelineRequest) error {
	llog := s.log.WithField("method", "ResumePipeline")
	llog.Debug("received request to resume pipeline")

	paused, err := s.IsPaused(ctx, req.Audience, req.PipelineId)
	if err != nil {
		return errors.Wrap(err, "error checking if pipeline is paused")
	}

	if !paused {
		llog.Debugf("pipeline '%s' not paused; nothing to do", req.PipelineId)
		return nil
	}

	llog.Debugf("pipeline '%s' paused; removing pause state now", req.PipelineId)
	if err := s.options.RedisBackend.Del(
		ctx,
		RedisPausedKey(util.AudienceToStr(req.Audience), req.PipelineId),
	).Err(); err != nil {
		return errors.Wrap(err, "error deleting pipeline pause state")
	}

	return nil
}

func (s *Store) sendAudienceTelemetry(ctx context.Context, aud *protos.Audience, val int64) {
	if aud == nil {
		return
	}

	// Unique services
	serviceKeys := s.options.RedisBackend.Keys(ctx, RedisAudiencePrefix+":"+aud.ServiceName+":*").Val()
	if len(serviceKeys) == 0 {
		// Completely new audience, send telemetry
		_ = s.options.Telemetry.GaugeDelta(telTypes.GaugeUsageNumServices, val, 1.0, statsd.Tag{"install_id", s.options.InstallID})
	}

	dataSourceKeys := s.options.RedisBackend.Keys(ctx, RedisAudiencePrefix+"*:*:*:"+aud.ComponentName+":*").Val()
	if len(dataSourceKeys) == 0 {
		// Completely new data source, send telemetry
		_ = s.options.Telemetry.GaugeDelta(telTypes.GaugeUsageNumDataSources, val, 1.0, statsd.Tag{"install_id", s.options.InstallID})
	}

	audExists := s.options.RedisBackend.Exists(ctx, RedisTelemetryAudience(aud)).Val() == 1

	if audExists && val < 0 || !audExists && val > 0 {
		if aud.OperationType == protos.OperationType_OPERATION_TYPE_PRODUCER {
			_ = s.options.Telemetry.GaugeDelta(telTypes.GaugeUsageNumProducers, val, 1.0, statsd.Tag{"install_id", s.options.InstallID})
		} else if aud.OperationType == protos.OperationType_OPERATION_TYPE_CONSUMER {
			_ = s.options.Telemetry.GaugeDelta(telTypes.GaugeUsageNumConsumers, val, 1.0, statsd.Tag{"install_id", s.options.InstallID})
		}
	}
}

func (s *Store) AddAudience(ctx context.Context, req *protos.NewAudienceRequest) error {
	if req == nil {
		return errors.New("request cannot be nil")
	}

	// Add it to the live bucket
	if err := s.options.RedisBackend.Set(
		ctx,
		RedisLiveKey(req.SessionId, s.options.NodeName, util.AudienceToStr(req.Audience)),
		[]byte(``),
		s.options.SessionTTL,
	).Err(); err != nil {
		return errors.Wrap(err, "error saving audience to store")
	}

	// And add it to more permanent storage (that doesn't care about the session id)
	if err := s.options.RedisBackend.Set(
		ctx,
		RedisAudienceKey(util.AudienceToStr(req.Audience)),
		[]byte(``),
		0,
	).Err(); err != nil {
		return errors.Wrap(err, "error saving audience to store")
	}

	s.sendAudienceTelemetry(ctx, req.Audience, 1)

	return nil
}

func (s *Store) DeleteAudience(ctx context.Context, req *protos.DeleteAudienceRequest) error {
	llog := s.log.WithField("method", "DeleteAudience")
	llog.Debug("received request to delete audience")

	// Check if there are any attached audiences
	attached, err := s.GetConfigByAudience(ctx, req.Audience)
	if err != nil {
		return errors.Wrapf(err, "error fetching configs for audience '%s'", util.AudienceToStr(req.Audience))
	}

	if len(attached) > 0 {
		err = fmt.Errorf("audience '%s' has one or more attached pipelines - cannot delete", util.AudienceToStr(req.Audience))

		return err
	}

	// Delete audience from bucket
	audStr := util.AudienceToStr(req.Audience)
	if err := s.options.RedisBackend.Del(ctx, RedisAudienceKey(audStr)).Err(); err != nil {
		return errors.Wrap(err, "error deleting audience from store")
	}

	// Send Analytics
	s.sendAudienceTelemetry(ctx, req.Audience, -1)

	return nil
}

func (s *Store) GetConfig(ctx context.Context) (map[*protos.Audience][]string, error) {
	cfgs := make(map[*protos.Audience][]string)

	audienceKeys, err := s.options.RedisBackend.Keys(ctx, RedisConfigPrefix+":*").Result()
	if err != nil {
		return nil, errors.Wrap(err, "error fetching config keys from store")
	}

	for _, aud := range audienceKeys {
		audience, pipelineID := util.ParseConfigKey(aud)
		if audience == nil {
			return nil, fmt.Errorf("invalid config key '%s'", aud)
		}

		if cfgs[audience] == nil {
			cfgs[audience] = make([]string, 0)
		}

		cfgs[audience] = append(cfgs[audience], pipelineID)
	}

	return cfgs, nil
}

// GetConfigByAudience returns a list of pipeline IDs attached to given audience
func (s *Store) GetConfigByAudience(ctx context.Context, audience *protos.Audience) ([]string, error) {
	cfgs, err := s.GetConfig(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "error fetching config from store")
	}

	pipelineIDs := make([]string, 0)

	for aud, ids := range cfgs {
		if util.AudienceEquals(aud, audience) {
			pipelineIDs = append(pipelineIDs, ids...)
		}
	}

	return pipelineIDs, nil
}

func (s *Store) GetLive(ctx context.Context) ([]*types.LiveEntry, error) {
	live := make([]*types.LiveEntry, 0)

	// Fetch all live keys from store
	keys, err := s.options.RedisBackend.Keys(ctx, RedisLivePrefix+":*").Result()
	if err != nil {
		return nil, errors.Wrap(err, "error fetching live keys from store")
	}

	// key is of the format:
	//
	// <sessionID>:<nodeName>:<<service>:<operation_type>:<operation_name>:<component_name>>
	// OR
	// <sessionID>:<nodeName>:register

	for _, key := range keys {
		parts := strings.SplitN(strings.TrimPrefix(key, RedisLivePrefix+":"), ":", 3)

		if len(parts) != 3 {
			return nil, errors.Errorf("invalid live key '%s', parts: %d", key, len(parts))
		}

		sessionID := parts[0]
		nodeName := parts[1]
		maybeAud := parts[2]

		entry := &types.LiveEntry{
			Key:       key,
			SessionID: sessionID,
			NodeName:  nodeName,
		}

		if maybeAud == "register" {
			entry.Register = true

			registerData, err := s.options.RedisBackend.Get(ctx, key).Result()
			if err != nil {
				return nil, errors.Wrapf(err, "error fetching register data for live key '%s'", key)
			}

			clientInfo := &protos.ClientInfo{}
			if err := proto.Unmarshal([]byte(registerData), clientInfo); err != nil {
				return nil, errors.Wrapf(err, "error unmarshaling register data for live key '%s'", key)
			}

			entry.Value = clientInfo
		} else {
			aud := util.AudienceFromStr(maybeAud)
			if aud == nil {
				return nil, errors.Errorf("invalid live key '%s'", key)
			}

			entry.Audience = aud
		}

		live = append(live, entry)
	}

	return live, nil
}

func (s *Store) GetAttachCommandsByService(ctx context.Context, serviceName string) ([]*protos.Command, error) {
	cmds := make([]*protos.Command, 0)

	search := fmt.Sprintf("%s:%s:*", RedisConfigPrefix, serviceName)

	keys, err := s.options.RedisBackend.Keys(ctx, search).Result()
	if err != nil {
		return nil, errors.Wrap(err, "error fetching config keys from store")
	}

	for _, key := range keys {
		aud, pipelineID := util.ParseConfigKey(key)
		if aud.ServiceName != serviceName {
			continue
		}

		pipeline, err := s.GetPipeline(ctx, pipelineID)
		if err != nil {
			return nil, errors.Wrapf(err, "error fetching pipeline '%s'", pipelineID)
		}

		cmds = append(cmds, &protos.Command{
			Audience: aud,
			Command: &protos.Command_AttachPipeline{
				AttachPipeline: &protos.AttachPipelineCommand{
					Pipeline: pipeline,
				},
			},
		})
	}

	return cmds, nil
}

func (s *Store) GetNotificationConfig(ctx context.Context, req *protos.GetNotificationRequest) (*protos.NotificationConfig, error) {
	data, err := s.options.RedisBackend.Get(ctx, RedisNotificationConfigKey(req.NotificationId)).Result()
	if err != nil {
		return nil, errors.Wrapf(err, "error fetching notification config '%s' from store", req.NotificationId)
	}

	// Decrypt data
	decrypted, err := s.options.Encryption.Decrypt([]byte(data))
	if err != nil {
		return nil, errors.Wrapf(err, "error decrypting notification config '%s'", req.NotificationId)
	}

	cfg := &protos.NotificationConfig{}
	if err := proto.Unmarshal(decrypted, cfg); err != nil {
		return nil, errors.Wrapf(err, "error unmarshaling notification config '%s'", req.NotificationId)
	}

	return cfg, nil
}

func (s *Store) GetNotificationConfigs(ctx context.Context) (map[string]*protos.NotificationConfig, error) {
	notificationConfigs := make(map[string]*protos.NotificationConfig)

	keys, err := s.options.RedisBackend.Keys(ctx, RedisNotificationConfigPrefix+":*").Result()
	if err != nil {
		return nil, errors.Wrap(err, "error fetching notification config keys from store")
	}

	for _, key := range keys {
		data, err := s.options.RedisBackend.Get(ctx, key).Result()
		if err != nil {
			return nil, errors.Wrapf(err, "error fetching notification config '%s' from store", key)
		}

		// Decrypt data
		decrypted, err := s.options.Encryption.Decrypt([]byte(data))
		if err != nil {
			return nil, errors.Wrapf(err, "error decrypting notification config '%s'", key)
		}

		notificationConfig := &protos.NotificationConfig{}
		if err := proto.Unmarshal(decrypted, notificationConfig); err != nil {
			return nil, errors.Wrapf(err, "error unmarshaling notification config '%s'", key)
		}

		notificationConfigs[key] = notificationConfig
	}

	return notificationConfigs, nil
}

func (s *Store) CreateNotificationConfig(ctx context.Context, req *protos.CreateNotificationRequest) error {
	data, err := proto.Marshal(req.Notification)
	if err != nil {
		return errors.Wrap(err, "error marshaling notification config")
	}

	// Encrypt data
	data, err = s.options.Encryption.Encrypt(data)
	if err != nil {
		return errors.Wrap(err, "error encrypting notification config")
	}

	key := RedisNotificationConfigKey(*req.Notification.Id)
	if err := s.options.RedisBackend.Set(ctx, key, data, 0).Err(); err != nil {
		return errors.Wrap(err, "error saving notification config to store")
	}

	return nil
}
func (s *Store) UpdateNotificationConfig(ctx context.Context, req *protos.UpdateNotificationRequest) error {
	if err := s.fillSensitiveFields(ctx, req); err != nil {
		return errors.Wrap(err, "error filling sensitive fields")
	}

	data, err := proto.Marshal(req.Notification)
	if err != nil {
		return errors.Wrap(err, "error marshaling notification config")
	}

	// Encrypt data
	data, err = s.options.Encryption.Encrypt(data)
	if err != nil {
		return errors.Wrap(err, "error encrypting notification config")
	}

	key := RedisNotificationConfigKey(*req.Notification.Id)
	if err := s.options.RedisBackend.Set(ctx, key, data, 0).Err(); err != nil {
		return errors.Wrap(err, "error saving notification config to store")
	}

	return nil
}

// fillSensitiveFields fills in any sensitive fields that were not provided in the request, assuming
// that the notification type has not changed.
func (s *Store) fillSensitiveFields(ctx context.Context, req *protos.UpdateNotificationRequest) error {
	existing, err := s.GetNotificationConfig(ctx, &protos.GetNotificationRequest{NotificationId: *req.Notification.Id})
	if err != nil {
		return errors.Wrap(err, "error fetching existing notification config")
	}

	// If the notification type changed, we don't need to fill in any sensitive fields, the user will have done that
	if req.Notification.Type != existing.Type {
		return nil
	}

	switch req.Notification.Type {
	case protos.NotificationType_NOTIFICATION_TYPE_EMAIL:
		email := req.Notification.GetEmail()
		switch email.Type {
		case protos.NotificationEmail_TYPE_SMTP:
			smtp := email.GetSmtp()
			if smtp.Password == "" {
				smtp.Password = existing.GetEmail().GetSmtp().GetPassword()
			}
		case protos.NotificationEmail_TYPE_SES:
			ses := email.GetSes()
			if ses.SesSecretAccessKey == "" {
				ses.SesSecretAccessKey = existing.GetEmail().GetSes().SesSecretAccessKey
			}
		}
	case protos.NotificationType_NOTIFICATION_TYPE_PAGERDUTY:
		pd := req.Notification.GetPagerduty()
		if pd.Token == "" {
			pd.Token = existing.GetPagerduty().Token
		}
	case protos.NotificationType_NOTIFICATION_TYPE_SLACK:
		slack := req.Notification.GetSlack()
		if slack.BotToken == "" {
			slack.BotToken = existing.GetSlack().BotToken
		}
	}

	return nil
}

func (s *Store) DeleteNotificationConfig(ctx context.Context, req *protos.DeleteNotificationRequest) error {
	configKey := RedisNotificationConfigKey(req.NotificationId)
	if err := s.options.RedisBackend.Del(ctx, configKey).Err(); err != nil {
		return errors.Wrap(err, "error deleting notification config from store")
	}

	// Delete all associations with pipelines
	keys, err := s.options.RedisBackend.Keys(ctx, RedisNotificationAssocPrefix+":*").Result()
	if err != nil {
		return errors.Wrap(err, "error fetching notification assoc keys from store")
	}

	for _, key := range keys {
		if !strings.HasSuffix(key, ":"+req.NotificationId) {
			continue
		}

		if err := s.options.RedisBackend.Del(ctx, key).Err(); err != nil {
			return errors.Wrap(err, "error deleting notification association from store")
		}
	}

	return nil
}

func (s *Store) AttachNotificationConfig(ctx context.Context, req *protos.AttachNotificationRequest) error {
	key := RedisNotificationAssocKey(req.PipelineId, req.NotificationId)
	if err := s.options.RedisBackend.Set(ctx, key, []byte(``), 0).Err(); err != nil {
		return errors.Wrap(err, "error saving notification association to store")
	}

	return nil
}

func (s *Store) DetachNotificationConfig(ctx context.Context, req *protos.DetachNotificationRequest) error {
	key := RedisNotificationAssocKey(req.PipelineId, req.NotificationId)
	if err := s.options.RedisBackend.Del(ctx, key).Err(); err != nil {
		return errors.Wrap(err, "error deleting notification association from store")
	}

	return nil
}

func (s *Store) GetNotificationConfigsByPipeline(ctx context.Context, pipelineID string) ([]*protos.NotificationConfig, error) {
	cfgs := make([]*protos.NotificationConfig, 0)

	// Fetch all notify config keys from store
	keys, err := s.options.RedisBackend.Keys(ctx, RedisNotificationAssocPrefix).Result()
	if err != nil {
		return nil, errors.Wrap(err, "error fetching notify config keys from store")
	}

	for _, key := range keys {
		if !strings.HasPrefix(key, pipelineID+":") {
			continue
		}

		parts := strings.Split(strings.Trim(RedisNotificationAssocPrefix+":", key), ":")
		if len(parts) != 2 {
			return nil, errors.Errorf("invalid notify config key '%s'", key)
		}

		configID := parts[1]

		// Fetch key so we get the notify config
		data, err := s.options.RedisBackend.Get(ctx, RedisNotificationConfigKey(configID)).Result()
		if err != nil {
			return nil, errors.Wrapf(err, "error fetching notify config key '%s' from store", configID)
		}

		decrypted, err := s.options.Encryption.Decrypt([]byte(data))
		if err != nil {
			return nil, errors.Wrapf(err, "error decrypting notification config '%s'", configID)
		}

		notifyConfig := &protos.NotificationConfig{}
		if err := proto.Unmarshal(decrypted, notifyConfig); err != nil {
			return nil, errors.Wrapf(err, "error unmarshalling notify config for key '%s'", configID)
		}

		cfgs = append(cfgs, notifyConfig)
	}

	return cfgs, nil
}

func (o *Options) validate() error {
	if o == nil {
		return errors.New("options cannot be nil")
	}

	if o.NodeName == "" {
		return errors.New("node name cannot be empty")
	}

	if o.RedisBackend == nil {
		return errors.New("RedisBackend backend cannot be nil")
	}

	if o.ShutdownCtx == nil {
		return errors.New("shutdown context cannot be nil")
	}

	if o.SessionTTL < time.Second || o.SessionTTL > time.Minute {
		return errors.New("session TTL must be between 1 second and 1 minute")
	}

	if o.InstallID == "" {
		return errors.New("install ID cannot be empty")
	}

	return nil
}

func (s *Store) GetAudiences(ctx context.Context) ([]*protos.Audience, error) {
	audiences := make([]*protos.Audience, 0)

	keys, err := s.options.RedisBackend.Keys(ctx, RedisAudiencePrefix+":*").Result()
	if err != nil {
		return nil, errors.Wrap(err, "error fetching audience keys from store")
	}

	for _, key := range keys {
		key = strings.TrimPrefix(key, RedisAudiencePrefix+":")
		aud := util.AudienceFromStr(key)
		if aud == nil {
			return nil, errors.Errorf("invalid audience key '%s'", key)
		}

		audiences = append(audiences, aud)
	}

	return audiences, nil
}

func (s *Store) GetAudiencesByService(ctx context.Context, serviceName string) ([]*protos.Audience, error) {
	audiences := make([]*protos.Audience, 0)

	keys, err := s.options.RedisBackend.Keys(ctx, RedisAudiencePrefix+":"+serviceName+":*").Result()
	if err != nil {
		return nil, errors.Wrap(err, "error fetching audience keys from store")
	}

	for _, key := range keys {
		key = strings.TrimPrefix(key, RedisAudiencePrefix+":")
		aud := util.AudienceFromStr(key)
		if aud == nil {
			return nil, errors.Errorf("invalid audience key '%s'", key)
		}

		audiences = append(audiences, aud)
	}

	return audiences, nil
}

func (s *Store) GetPaused(ctx context.Context) (map[string]*types.PausedEntry, error) {
	keys, err := s.options.RedisBackend.Keys(ctx, RedisPausedPrefix+":*").Result()
	if err != nil {
		return nil, errors.Wrap(err, "error fetching paused keys from store")
	}

	paused := make(map[string]*types.PausedEntry)

	for _, key := range keys {
		entry := &types.PausedEntry{
			Key:        key,
			Audience:   nil,
			PipelineID: "",
		}

		parts := strings.SplitN(strings.TrimPrefix(key, RedisPausedPrefix+":"), ":", 2)
		if len(parts) != 2 {
			return nil, errors.Errorf("invalid paused key '%s' (incorrect number of parts '%d')", key, len(parts))
		}

		pipelineID := parts[0]
		audStr := parts[1]

		aud := util.AudienceFromStr(audStr)
		if aud == nil {
			return nil, errors.Errorf("invalid paused key '%s' (unable to convert audience str '%s' to *Audience)", key, audStr)
		}

		entry.Audience = aud
		entry.PipelineID = pipelineID

		paused[pipelineID] = entry
	}

	return paused, nil
}

func (s *Store) IsPipelineAttached(ctx context.Context, audience *protos.Audience, pipelineID string) (bool, error) {
	key := RedisConfigKey(audience, pipelineID)

	if err := s.options.RedisBackend.Get(ctx, key).Err(); err != nil {
		if err == redis.Nil {
			return false, nil
		}

		return false, errors.Wrap(err, "error fetching pipeline attachment from store")
	}

	return true, nil
}

func (s *Store) IsPipelineAttachedAny(ctx context.Context, pipelineID string) bool {
	search := fmt.Sprintf(RedisConfigKeyFormat, "*", pipelineID)
	keys := s.options.RedisBackend.Keys(ctx, search).Val()
	return len(keys) > 0
}

type PipelineUsage struct {
	PipelineId string
	Active     bool
	NodeName   string
	SessionId  string
	Audience   *protos.Audience
}

// GetPipelineUsage gets usage across the entire cluster
func (s *Store) GetPipelineUsage(ctx context.Context) ([]*PipelineUsage, error) {
	pipelines := make([]*PipelineUsage, 0)

	// Get config for all pipelines & audiences
	cfgs, err := s.GetConfig(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "error getting configs")
	}

	// No cfgs == no pipelines
	if len(cfgs) == 0 {
		return pipelines, nil
	}

	// Get live clients
	live, err := s.GetLive(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "error getting live audiences")
	}

	// Build list of all used pipelines
	for aud, pipelineIDs := range cfgs {
		for _, pid := range pipelineIDs {
			pu := &PipelineUsage{
				PipelineId: pid,
				Audience:   aud,
			}

			// Check if this pipeline is "active"
			for _, l := range live {
				if util.AudienceEquals(l.Audience, aud) {
					pu.Active = true
					pu.NodeName = l.NodeName
					pu.SessionId = l.SessionID
				}
			}

			pipelines = append(pipelines, pu)
		}
	}

	return pipelines, nil
}

func (s *Store) GetPipelineHistory(ctx context.Context, pipelineID string) (map[int32]*protos.Pipeline, error) {
	history, err := s.options.RedisBackend.HGetAll(ctx, RedisPipelineHistoryKey(pipelineID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrPipelineNotFound
		}

		return nil, errors.Wrap(err, "error fetching pipeline history from store")
	}

	pipelines := map[int32]*protos.Pipeline{}

	for _, data := range history {
		pipeline := &protos.Pipeline{}

		if err := proto.Unmarshal([]byte(data), pipeline); err != nil {
			return nil, errors.Wrap(err, "error deserializing pipeline")
		}

		pipelines[pipeline.Version] = pipeline
	}

	return pipelines, nil
}

func (s *Store) StorePipelineHistory(ctx context.Context, pipeline *protos.Pipeline) error {
	for _, step := range pipeline.Steps {
		step.XWasmBytes = nil
	}

	data, err := proto.Marshal(pipeline)
	if err != nil {
		return errors.Wrap(err, "error serializing pipeline to protobuf")
	}

	key := RedisPipelineHistoryKey(pipeline.Id)
	if err := s.options.RedisBackend.HSet(ctx, key, pipeline.Version, data).Err(); err != nil {
		return errors.Wrap(err, "error saving pipeline to store")
	}

	s.log.Debugf("stored pipeline history for pipeline '%s' version '%d'", pipeline.Id, pipeline.Version)

	return nil
}

// GetActivePipelineUsage gets *ACTIVE* pipeline usage on the CURRENT node
// NOTE: This method is a helper for GetPipelineUsage().
func (s *Store) GetActivePipelineUsage(ctx context.Context, pipelineID string) ([]*PipelineUsage, error) {
	usage, err := s.GetPipelineUsage(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "error getting pipeline usage")
	}

	active := make([]*PipelineUsage, 0)

	for _, u := range usage {
		if !u.Active {
			continue
		}

		if u.NodeName != s.options.NodeName {
			continue
		}

		if u.PipelineId != pipelineID {
			continue
		}

		active = append(active, u)
	}

	return active, nil
}

// WatchKeys will watch for key changes for given key pattern; every time key
// is updated, it will send a message on the return channel.
func (s *Store) WatchKeys(quitCtx context.Context, key string) chan struct{} {
	s.log.Infof("launching dedicated WatchKeys goroutine for key '%s'", key)

	ps := s.options.RedisBackend.PSubscribe(s.options.ShutdownCtx, RedisKeyWatchPrefix+key)

	// OK to buffer since the contents are so small
	keyChan := make(chan struct{}, 100_000)

	go func() {
		defer func() {
			if err := ps.Unsubscribe(s.options.ShutdownCtx, RedisKeyWatchPrefix+key); err != nil {
				s.log.Errorf("error unsubscribing from key '%s': %s", key, err)
			}
		}()

		for {
			select {
			case <-s.options.ShutdownCtx.Done():
				s.log.Debug("shutdown detected - exiting WatchKeys()")
				return
			case <-ps.Channel():
				keyChan <- struct{}{}
			case <-quitCtx.Done():
				s.log.Debug("quit detected - exiting WatchKeys()")
				return
			}
		}
	}()

	s.log.Infof("dedicated WatchKeys goroutine for key '%s' exiting", key)

	return keyChan
}

func (s *Store) GetSchema(ctx context.Context, aud *protos.Audience) (*protos.Schema, error) {
	audStr := util.AudienceToStr(aud)

	data, err := s.options.RedisBackend.Get(ctx, RedisSchemaKey(audStr)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			// No schema yet, generate empty one
			return &protos.Schema{
				XVersion:   0,
				JsonSchema: []byte(`{"$schema":"http://json-schema.org/draft-07/schema#","properties":{},"required":[],"type":"object"}`),
			}, nil
		}
		return nil, errors.Wrapf(err, "error fetching schema for audience '%s'", audStr)
	}

	schema := &protos.Schema{}
	if err := proto.Unmarshal([]byte(data), schema); err != nil {
		return nil, errors.Wrapf(err, "error unmarshaling schema for audience '%s'", audStr)
	}

	return schema, nil
}

func (s *Store) AddSchema(ctx context.Context, req *protos.SendSchemaRequest) error {
	// Save to K/V
	schemaData, err := proto.Marshal(req.Schema)
	if err != nil {
		return errors.Wrap(err, "error serializing schema to protobuf")
	}

	key := RedisSchemaKey(util.AudienceToStr(req.Audience))
	if err := s.options.RedisBackend.Set(ctx, key, schemaData, 0).Err(); err != nil {
		return errors.Wrap(err, "error saving schema to store")
	}

	return nil
}

func (s *Store) GetAudiencesBySessionID(ctx context.Context, sessionID string) ([]*protos.Audience, error) {
	// Get audiences by session ID
	prefix := RedisLivePrefix + ":" + sessionID
	keys, err := s.options.RedisBackend.Keys(ctx, prefix+":*").Result()
	if err != nil {
		return nil, errors.Wrap(err, "error fetching live keys from store")
	}

	// Looking for format: <sessionID>:<nodeName>:<<service>:<operation_type>:<operation_name>:<component_name>>
	// Ignoring format: <sessionID>:<nodeName>:register

	live := make([]*protos.Audience, 0)
	for _, key := range keys {
		// Trim <sessionID>:<nodeName>: from key
		// Remaining key should just be <service>:<operation_type>:<operation_name>:<component_name>
		audStr := strings.TrimPrefix(key, prefix+":"+s.options.NodeName+":")
		parts := strings.Split(audStr, ":")
		if len(parts) != 4 {
			// This is a :registerKey, ignore these, we only want audiences
			continue
		}

		aud := util.AudienceFromStr(audStr)
		if aud == nil {
			return nil, errors.Errorf("invalid live key '%s'", key)
		}

		live = append(live, aud)
	}

	return live, nil
}

func (s *Store) GetInstallID(ctx context.Context) (string, error) {
	// Check cache first
	if s.options.InstallID != "" {
		return s.options.InstallID, nil
	}

	v, err := s.options.RedisBackend.Get(ctx, InstallIDKey).Result()
	if errors.Is(err, redis.Nil) {
		id, setErr := s.setStreamdalID(ctx)
		if setErr != nil {
			return "", setErr
		}
		return id, nil
	} else if err != nil {
		return "", errors.Wrap(err, "unable to query cluster ID")
	}

	return v, nil
}

func (s *Store) setStreamdalID(ctx context.Context) (string, error) {
	id, err := s.options.RedisBackend.Get(ctx, InstallIDKey).Result()
	if errors.Is(err, redis.Nil) {
		// Create new ID
		id := util.GenerateUUID()

		s.options.InstallID = id

		err := s.options.RedisBackend.Set(ctx, InstallIDKey, id, 0).Err()
		if err != nil {
			return "", errors.Wrap(err, "unable to set cluster ID")
		}

		s.log.Debugf("Set cluster ID '%s'", id)
		return id, nil
	} else if err != nil {
		return "", errors.Wrap(err, "unable to query cluster ID")
	}

	// Streamdal cluster ID already set
	return id, nil
}

func (s *Store) GetActiveTailCommandsByService(ctx context.Context, serviceName string) ([]*protos.Command, error) {
	tailCommands := make([]*protos.Command, 0)

	activeTailKeys, err := s.options.RedisBackend.Keys(ctx, RedisActiveTailPrefix+":"+serviceName+":*").Result()
	// No keys in redis
	if errors.Is(err, redis.Nil) {
		s.log.Debug("resume: no active tails found")
		return tailCommands, nil
	} else if err != nil {
		s.log.Errorf("resume: error fetching active tail keys from store: %s", err)
		return nil, errors.Wrap(err, "error fetching active tail keys from store")
	}

	s.log.Debugf("resume: found '%d' active tails", len(activeTailKeys))

	// Each key is an active tail - fetch each + decode
	for _, key := range activeTailKeys {
		encodedTailReq, err := s.options.RedisBackend.Get(ctx, key).Result()
		if err != nil {
			return nil, errors.Wrapf(err, "error fetching key '%s' from store", key)
		}

		// Attempt to decode
		tailRequest := &protos.TailRequest{}

		if err := proto.Unmarshal([]byte(encodedTailReq), tailRequest); err != nil {
			return nil, errors.Wrap(err, "error unmarshaling tail request")
		}

		if err := validate.StartTailRequest(tailRequest); err != nil {
			return nil, errors.Wrap(err, "error validating tail request")
		}

		// We shouldn't normally hit this - if we do, it's a bug - it means that
		// we are storing a tail command that is not actually a start command.
		if tailRequest.Type != protos.TailRequestType_TAIL_REQUEST_TYPE_START {
			s.log.Warningf("bug? active tail command for key '%s' should contain a start request type but contains '%s'",
				serviceName, tailRequest.Type.String())

			continue
		}

		tailCommands = append(tailCommands, &protos.Command{
			Audience: tailRequest.Audience,
			Command: &protos.Command_Tail{
				Tail: &protos.TailCommand{
					Request: tailRequest,
				},
			},
		})
	}

	return tailCommands, nil
}

func (s *Store) AddActiveTailRequest(ctx context.Context, req *protos.TailRequest) (string, error) {
	encodedReq, err := proto.Marshal(req)
	if err != nil {
		return "", errors.Wrap(err, "unable to marshal tail request")
	}

	tailKey := RedisActiveTailKey(req.Id)

	if err := s.options.RedisBackend.Set(ctx, tailKey, encodedReq, RedisActiveTailTTL).Err(); err != nil {
		return "", errors.Wrap(err, "unable to save tail request")
	}

	return tailKey, nil
}

func (s *Store) GetTailRequestById(ctx context.Context, tailID string) (*protos.TailRequest, error) {
	encodedReq, err := s.options.RedisBackend.Get(ctx, RedisActiveTailKey(tailID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errors.Errorf("no tail request found for tail ID '%s'", tailID)
		}
		return nil, errors.Wrapf(err, "error fetching tail request '%s' from store", tailID)
	}

	// Attempt to decode
	tailRequest := &protos.TailRequest{}

	if err := proto.Unmarshal([]byte(encodedReq), tailRequest); err != nil {
		return nil, errors.Wrap(err, "error unmarshaling tail request")
	}

	return tailRequest, nil
}

func (s *Store) GetPausedTailRequestById(ctx context.Context, tailID string) (*protos.TailRequest, error) {
	encodedReq, err := s.options.RedisBackend.Get(ctx, RedisPausedTailKey(tailID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errors.Errorf("no tail request found for tail ID '%s'", tailID)
		}
		return nil, errors.Wrapf(err, "error fetching tail request '%s' from store", tailID)
	}

	// Attempt to decode
	tailRequest := &protos.TailRequest{}

	if err := proto.Unmarshal([]byte(encodedReq), tailRequest); err != nil {
		return nil, errors.Wrap(err, "error unmarshaling tail request")
	}

	return tailRequest, nil
}

func (s *Store) PauseTailRequest(ctx context.Context, req *protos.PauseTailRequest) (*protos.TailRequest, error) {
	// Fetch tail request
	tailRequest, err := s.GetTailRequestById(ctx, req.TailId)
	if err != nil {
		return nil, errors.Wrap(err, "error fetching tail request")
	}

	// Delete tail request from active
	if err := s.options.RedisBackend.Del(ctx, RedisActiveTailKey(tailRequest.Id)).Err(); err != nil {
		return nil, errors.Wrap(err, "error deleting tail request")
	}

	data, err := proto.Marshal(tailRequest)
	if err != nil {
		return nil, errors.Wrap(err, "error marshaling tail request")
	}

	// Save to paused
	if err := s.options.RedisBackend.Set(ctx, RedisPausedTailKey(tailRequest.Id), data, 0).Err(); err != nil {
		return nil, errors.Wrap(err, "error saving paused tail request")
	}

	return tailRequest, nil
}

func (s *Store) ResumeTailRequest(ctx context.Context, req *protos.ResumeTailRequest) (*protos.TailRequest, error) {
	// Fetch tail request
	tailRequest, err := s.GetPausedTailRequestById(ctx, req.TailId)
	if err != nil {
		return nil, errors.Wrap(err, "error fetching tail request")
	}

	// Delete tail request from paused
	if err := s.options.RedisBackend.Del(ctx, RedisPausedTailKey(tailRequest.Id)).Err(); err != nil {
		return nil, errors.Wrap(err, "error deleting tail request")
	}

	// Save to active
	if _, err := s.AddActiveTailRequest(ctx, tailRequest); err != nil {
		return nil, errors.Wrap(err, "error saving active tail request")
	}

	return tailRequest, nil
}

func applyPipelineDefaults(pipeline *protos.Pipeline) {
	// Loop over steps and ensure negate is set for detective steps if it's nil
	// Rust will panic if it's not set
	for _, step := range pipeline.Steps {
		if d := step.GetDetective(); d != nil {
			if d.Negate == nil {
				d.Negate = proto.Bool(false)
			}
		}
	}
}

func (s *Store) GetCreationDate(ctx context.Context) (int64, error) {
	created := s.options.RedisBackend.Get(ctx, RedisCreationDateKey).Val()
	if created == "" {
		return 0, nil
	}

	createdTS, err := strconv.ParseInt(created, 10, 64)
	if err != nil {
		return 0, errors.Wrap(err, "unable to convert creation date to int")
	}

	return createdTS, nil
}

func (s *Store) SetCreationDate(ctx context.Context, ts int64) error {
	if err := s.options.RedisBackend.Set(ctx, RedisCreationDateKey, ts, 0).Err(); err != nil {
		return errors.Wrap(err, "unable to set creation date in store")
	}

	return nil
}
