package tix

import (
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/mcuadros/go-defaults"
)

type (
	Config struct {
		DSN        string `toml:"dsn" default:"root:@tcp(localhost:4000)/test?charset=utf8&parseTime=True&loc=Local"`
		MaxTxnSize int    `toml:"max_transaction_size" default:"1000"`

		JobQueueConfig JobQueueConfig `toml:"job_queue"`
		PubSubConfig   PubSubConfig   `toml:"pubsub"`
		ElectionConfig ElectionConfig `toml:"election"`
	}

	JobQueueConfig struct {
		TablePrefix  string `toml:"table_prefix" default:"tix_job_queue"`
		PollInterval string `toml:"poll_interval" default:"1s"`
		EnableGC     bool   `toml:"enable_gc" default:"true"`
		GCKeepItems  int    `toml:"gc_keep_items" default:"10000"`
		GCInterval   string `toml:"gc_interval" default:"1m"`
	}

	PubSubConfig struct {
		TablePrefix  string `toml:"table_prefix" default:"tix_pubsub"`
		PollInterval string `toml:"poll_interval" default:"1s"`
		EnableGC     bool   `toml:"enable_gc" default:"true"`
		GCKeepItems  int    `toml:"gc_keep_items" default:"10000"`
		GCInterval   string `toml:"gc_interval" default:"1m"`
	}

	ElectionConfig struct {
		TableName    string `toml:"table_name" default:"tix_election"`
		PollInterval string `toml:"poll_interval" default:"1s"`
		TermTimeout  string `toml:"term_timeout" default:"1m"`
	}
)

func (pubsubConfig PubSubConfig) StreamTblName(streamName string) string {
	return pubsubConfig.TablePrefix + "_" + streamName
}

func (pubsubConfig PubSubConfig) MetaTblName() string {
	return pubsubConfig.TablePrefix
}

func NewDefaultConfig() *Config {
	var config Config
	defaults.SetDefaults(&config)
	return &config
}

func SampleConfig() string {
	config := NewDefaultConfig()
	var buf strings.Builder
	toml.NewEncoder(&buf).Encode(config)
	return buf.String()
}

func ParseConfig(path string) (*Config, error) {
	config := NewDefaultConfig()
	fp, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	if _, err := toml.DecodeReader(fp, &config); err != nil {
		return nil, err
	}
	return config, nil
}

func ParseConfigFromContent(content string) (*Config, error) {
	config := NewDefaultConfig()
	if _, err := toml.Decode(content, &config); err != nil {
		return nil, err
	}
	return config, nil
}
