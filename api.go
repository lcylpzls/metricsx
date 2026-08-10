package metricsx

import "github.com/lcylpzls/metricsx/internal/core"

const Version = core.Version

const (
	CodeInvalidConfig     = core.CodeInvalidConfig
	CodeAlreadyRegistered = core.CodeAlreadyRegistered
)

type (
	Option         = core.Option
	Metrics        = core.Metrics
	Sink           = core.Sink
	SnapshotMetric = core.SnapshotMetric
	Snapshot       = core.Snapshot
	Snapshotter    = core.Snapshotter
)

func New(opts ...Option) (*Metrics, error) { return core.New(opts...) }
func WithSink(sink Sink) Option            { return core.WithSink(sink) }
