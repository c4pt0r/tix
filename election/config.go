package election

import "github.com/c4pt0r/tix"

// Config is the configuration for the election service
type Config struct {
	// DSN is the data source name.
	DSN string `toml:"dsn" env:"DSN" env-default:"root:@tcp(localhost:4000)/test?charset=utf8&parseTime=True&loc=Local"`
	// TablePrefix
	TermTable string `toml:"table_prefix" env:"TABLE_PREFIX" env-default:"tix_leader_election_campaigns"`
	// TermTimeout
	TermTimeoutInSec int64 `toml:"term_timeout_in_sec" env:"TERM_TIMEOUT_IN_SEC" env-default:"10"`
	// PollIntervalInSec
	PollIntervalInSec int64 `toml:"poll_interval_in_sec" env:"POLL_TIMEOUT_IN_SEC" env-default:"3"`
}

var _ tix.IConfig = (*Config)(nil)

func (c *Config) Name() string {
	return "election"
}
