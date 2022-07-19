package tix

import (
	"testing"

	"github.com/c4pt0r/log"
)

func TestConfig(t *testing.T) {
	jqConfig, err := ParseConfig("config_test.toml")
	if err != nil {
		t.Fatal(err)
	}
	log.Errorf("jqConfig: %+v", jqConfig)
}
