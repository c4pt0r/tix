package jobqueue

import (
	"fmt"

	"github.com/c4pt0r/tix"
)

// Config is the configuration for the job queue.
type Config struct {
	// DSN is the data source name.
	DSN string `toml:"dsn" env:"DSN" env-default:"root:@tcp(localhost:4000)/test?charset=utf8&parseTime=True&loc=Local"`
	// MaxBatchSize is the maximum number of messages to batch a transaction.
	MaxBatchSize int `toml:"max_batch_size" env:"MAX_BATCH_SIZE" env-default:"1000"`
	// PollIntervalInMs is the interval to poll the database.
	PollIntervalInMs int `toml:"poll_interval_in_ms" env:"POLL_INTERVAL_IN_MS" env-default:"100"`
	// TablePrefix
	TablePrefix string `toml:"table_prefix" env:"TABLE_PREFIX" env-default:"tix_jobqueue"`
}

func (c *Config) String() string {
	return fmt.Sprintf("%+v", *c)
}

var _ tix.IConfig = (*Config)(nil)

func (c *Config) Name() string {
	return "jobqueue"
}
