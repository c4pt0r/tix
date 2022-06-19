package jobqueue

import (
	"math/rand"
	"testing"
	"time"

	"github.com/c4pt0r/log"
	"github.com/c4pt0r/tix"
	_ "github.com/go-sql-driver/mysql"
)

func TestCreateChannelTable(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	s, err := OpenStore(tix.DefaultConfig[Config]())
	if err != nil {
		t.Fatal(err)
	}
	jc, err := s.OpenJobChannel("test")
	if err != nil {
		t.Fatal(err)
	}

	err = jc.SubmitJob(&Job{
		Name:       tix.RandomString("job-", 10),
		Type:       "test",
		Data:       []byte("test_data"),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		ScheduleAt: time.Now(),
	})
	if err != nil {
		t.Error(err)
	}

	jobs, err := jc.FetchJobs("worker-1", DefaultGetOpt().
		SetLimit(1).
		SetAssign(true))

	if err != nil {
		t.Error(err)
	}

	for _, job := range jobs {
		log.I(job)
	}
}
