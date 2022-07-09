package tagstore

import "github.com/c4pt0r/tix"

// Config is the configuration for the TagStore service
type Config struct {
	// DSN is the data source name.
	DSN string `toml:"dsn" env:"DSN" env-default:"root:@tcp(localhost:4000)/test?charset=utf8&parseTime=True&loc=Local"`
	// TagTable is the name of the table to store tags in.
	TagTable string `toml:"tag_table" env:"TAG_TABLE" env-default:"tix_tag_table"`

	// TagMapTable is the name of the table to store tag maps in.
	TagMapTable string `toml:"tag_map_table" env:"TAG_MAP_TABLE" env-default:"tix_tag_map_table"`
}

var _ tix.IConfig = (*Config)(nil)

func (c *Config) Name() string {
	return "tagstore"
}
