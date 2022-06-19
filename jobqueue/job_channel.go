package jobqueue

import (
	"database/sql"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/c4pt0r/log"
)

type JobChannel struct {
	store                   *TiDBStore
	channName               string
	maxDispatchedScheduleAt atomic.Value
}

func OpenJobChannel(s *TiDBStore, channelName string) (*JobChannel, error) {
	jc := &JobChannel{
		store:                   s,
		channName:               channelName,
		maxDispatchedScheduleAt: atomic.Value{},
	}
	if err := jc.createChannelTable(); err != nil {
		return nil, err
	}

	jc.maxDispatchedScheduleAt.Store(time.Time{})

	go func() {
		for {
			maxScheduleAt, err := jc.getMaxScheduleAt()
			if err != nil {
				log.E("getMaxScheduleAt", err)
			}
			jc.maxDispatchedScheduleAt.Store(maxScheduleAt)
			time.Sleep(time.Second * 5)
		}
	}()

	return jc, nil
}

func (jc *JobChannel) getMaxScheduleAt() (time.Time, error) {
	stmt := fmt.Sprintf(`
		SELECT MAX(schedule_at)
		FROM %s
		WHERE status != ?
	`, jc.tblNameForJobs())
	log.D("getMaxScheduleAt", stmt)
	var maxScheduleAt time.Time
	err := jc.store.DB().QueryRow(stmt, JobStatusPending).Scan(&maxScheduleAt)
	if err != nil {
		return time.Time{}, err
	}
	return maxScheduleAt, nil
}

func (jc *JobChannel) tblNameForJobs() string {
	return fmt.Sprintf("%s_jobs_%s", jc.store.GetTablePrefix(), jc.channName)
}

func (jc *JobChannel) createChannelTable() error {
	stmt := Job{}.CreateTableSQL(jc.tblNameForJobs())
	log.D("createChannelTable", stmt)
	_, err := jc.store.DB().Exec(stmt)
	if err != nil {
		return err
	}
	return nil
}

func (jc *JobChannel) SubmitJob(job *Job) error {
	if job.CreatedAt.IsZero() {
		job.CreatedAt = time.Now()
	}
	if job.UpdatedAt.IsZero() {
		job.UpdatedAt = time.Now()
	}
	if job.ScheduleAt.IsZero() {
		job.ScheduleAt = time.Now()
	}
	stmt := fmt.Sprintf(`
		INSERT INTO %s
			(name, data, status, type, assign_to, schedule_at, created_at, updated_at)
		VALUES
			(?, ?, ?, ?, ?, ?, ?, ?)
	`, jc.tblNameForJobs())
	log.D("SubmitJob", stmt)
	ret, err := jc.store.DB().Exec(stmt,
		job.Name,
		job.Data,
		job.Status,
		job.Type,
		job.AssignTo,
		job.ScheduleAt,
		job.CreatedAt,
		job.UpdatedAt)
	if err != nil {
		return err
	}
	id, err := ret.LastInsertId()
	if err != nil {
		return err
	}
	job.ID = id
	return nil
}

func (jc *JobChannel) FetchJobs(workerID string, opt *GetOpt) ([]*Job, error) {
	txn, err := jc.store.DB().Begin()
	if err != nil {
		return nil, err
	}
	defer txn.Rollback()

	predicates := []string{}
	params := []any{}

	predicates = append(predicates, "status = ?")
	params = append(params, JobStatusPending)

	predicates = append(predicates, "(assign_to = ? OR assign_to = '')")
	params = append(params, workerID)

	if opt.Tp != "" {
		predicates = append(predicates, "type = ?")
		params = append(params, opt.Tp)
	}

	predicates = append(predicates, "schedule_at <= ?")
	params = append(params, time.Now())

	if !jc.maxDispatchedScheduleAt.Load().(time.Time).IsZero() {
		predicates = append(predicates, "schedule_at >= ?")
		params = append(params, jc.maxDispatchedScheduleAt.Load().(time.Time))
	}

	predicatesClause := strings.Join(predicates, " AND ")
	params = append(params, opt.Limit)

	stmt := fmt.Sprintf(`
		SELECT 
			id, name, data, status, type, 
			schedule_at, progress_data, result_code, result_data, error_message, created_at, updated_at
		FROM %s
		WHERE 
			%s
		ORDER BY 
			schedule_at ASC
		LIMIT ?
		FOR UPDATE
		`, jc.tblNameForJobs(), predicatesClause)
	var rows *sql.Rows
	log.D("FetchJobs", stmt)

	rows, err = txn.Query(stmt, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	jobs := []*Job{}
	for rows.Next() {
		job := &Job{}
		err := rows.Scan(
			&job.ID,
			&job.Name,
			&job.Data,
			&job.Status,
			&job.Type,
			&job.ScheduleAt,
			&job.ProgressData,
			&job.ResultCode,
			&job.ResultData,
			&job.ErrorMessage,
			&job.CreatedAt,
			&job.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}

	for _, job := range jobs {
		stmt := fmt.Sprintf(`
			UPDATE %s
			SET 
				status = ?,
				owner = ?,
				updated_at = ?
			WHERE 
				id = ?
			`, jc.tblNameForJobs())
		_, err := txn.Exec(stmt, JobStatusDispatched, workerID, time.Now(), job.ID)
		if err != nil {
			return nil, err
		}
	}
	err = txn.Commit()
	if err != nil {
		return nil, err
	}
	return jobs, nil
}

func (jc *JobChannel) UpdateJobsForWorker(workerID string, jobs []*Job) error {
	txn, err := jc.store.DB().Begin()
	if err != nil {
		return err
	}
	defer txn.Rollback()

	stmt := fmt.Sprintf(`
		UPDATE %s
		SET 
			status = ?,
			progress_data = ?,
			result_code = ?,
			result_data = ?,
			error_message = ?,
			updated_at = ?
		WHERE 
			id = ? AND worker_id = ?
		`, jc.tblNameForJobs())
	s, err := txn.Prepare(stmt)
	if err != nil {
		return err
	}

	for _, j := range jobs {
		log.D("UpdateJob", stmt)
		_, err = s.Exec(
			j.Status,
			j.ProgressData,
			j.ResultCode,
			j.ResultData,
			j.ErrorMessage,
			time.Now(),
			j.ID,
			workerID)
		if err != nil {
			return err
		}
	}
	return txn.Commit()
}
