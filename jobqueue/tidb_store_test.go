package jobqueue

import (
	"math/rand"
	"testing"
	"time"

	"github.com/c4pt0r/tix"
	_ "github.com/go-sql-driver/mysql"
)

func TestCreateChannelTable(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	s, err := OpenStore(tix.DefaultConfig[Config]())
	if err != nil {
		t.Fatal(err)
	}
	ch, err := s.OpenJobChannel("jobqueue_test")
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		err = ch.SubmitJob(&Job{
			Name:     tix.RandomString("job-", 10),
			Type:     "test",
			AssignTo: "worker-1",
			Data:     []byte("test_data"),
		})
		if err != nil {
			t.Error(err)
		}
	}

	time.Sleep(time.Second)

	jobs, err := ch.FetchJobs("worker-1", NewGetOpt().SetLimit(10))
	if err != nil {
		t.Error(err)
	}

	if len(jobs) != 10 {
		t.Errorf("expected 10 jobs, got %d", len(jobs))
	}
}
