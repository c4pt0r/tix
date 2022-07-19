# JobQueue


## Usage

```

Job Disptcher:

s, _ := jobqueue.OpenStore(tix.NewDefaultConfig())
ch, _ := s.OpenJobChannel(channalName)

ch.SubmitJob(&jobqueue.Job{
    Name: str,
    Type: str,
    AssignTo: workerID,
    Data: []byte{},
})



Job Worker:

s, _ := jobqueue.OpenStore(tix.NewDefaultConfig())
ch, _ := s.OpenJobChannel(channalName)


jobs := ch.FetchJobs(workerID, NewGetOpt().SetLimit(10))
...


jobs[0].ProgressData = []byte("progress_data")
jobs.[0].

err = ch.UpdateJobsForWorker("worker-1", []*Job{
	jobs[0],
})

```
