package jobqueue

import (
	"testing"

	"github.com/alecthomas/assert"
	"github.com/c4pt0r/log"
)

func TestSomething(t *testing.T) {
	assert := assert.New(t)

	for i := 0; i < 10; i++ {
		job := NewJob("demo", "type-default", "content")
		log.I(job)
		assert.Equal("demo", job.Name)
	}
}
