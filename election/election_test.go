package election

import (
	"math/rand"
	"testing"
	"time"

	"github.com/c4pt0r/tix"
	_ "github.com/go-sql-driver/mysql"
)

func TestElection(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	cfg := tix.DefaultConfig[Config]()

	campaign, err := NewCampaign(cfg, "test", "worker-1")
	err = campaign.Init()
	if err != nil {
		t.Fatal(err)
	}
	if err != nil {
		t.Fatal(err)
	}

	campaign.Elect()
}
