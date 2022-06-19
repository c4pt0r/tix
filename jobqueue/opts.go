package jobqueue

type GetOpt struct {
	Limit  int
	Tp     string
	Assign bool
	Status JobStatus
}

func NewGetOpt() *GetOpt {
	return &GetOpt{
		Limit:  1,
		Tp:     "",
		Assign: true,
		Status: JobStatusPending,
	}
}

func (o *GetOpt) SetLimit(n int) *GetOpt {
	o.Limit = n
	return o
}

func (o *GetOpt) SetTp(s string) *GetOpt {
	o.Tp = s
	return o
}

func (o *GetOpt) SetStatus(st JobStatus) *GetOpt {
	o.Status = st
	return o
}

func (o *GetOpt) SetAssign(b bool) *GetOpt {
	o.Assign = b
	return o
}
