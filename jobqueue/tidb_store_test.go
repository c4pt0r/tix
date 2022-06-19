package jobqueue

import (
	"database/sql"
	"testing"
	"time"

	"github.com/c4pt0r/tix"
	_ "github.com/go-sql-driver/mysql"
)

func TestCreateChannelTable(t *testing.T) {
	db, err := sql.Open("mysql", "root:@tcp(localhost:4000)/test")
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	s := OpenStoreWithDB(db, tix.DefaultConfig[Config]())
	jc, err := s.OpenJobChannel("test")
	if err != nil {
		t.Error(err)
	}

	err = jc.SubmitJob(&Job{
		Name:       "test",
		Type:       "test",
		Data:       "test_data",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		ScheduleAt: time.Now(),
	})
	if err != nil {
		t.Error(err)
	}
}
