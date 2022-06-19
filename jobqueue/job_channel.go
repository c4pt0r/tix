package jobqueue

import (
	"fmt"
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

	go func() {
		for {
			maxScheduleAt, err := jc.getMaxScheduleAt()
			if err != nil {
				log.E("getMaxScheduleAt", err)
				time.Sleep(time.Second)
				continue
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
		return maxScheduleAt, err
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

	stmt := fmt.Sprintf(`
		SELECT 
			id, name, data, status, type, schedule_at, progress_data, result_code, result_data, error_message, created_at, updated_at
		FROM %s
		WHERE 
			status = ? AND (assign_to = ? OR assign_to = '') AND (schedule_at <= ? AND schedule_at >= ?)
		ORDER BY 
			schedule_at 
		ASC
		LIMIT 1
		FOR UPDATE
		`, jc.tblNameForJobs())
	log.D("FetchJobs", stmt)
	rows, err := txn.Query(stmt, JobStatusPending, workerID, time.Now(), jc.maxDispatchedScheduleAt, opt.Limit)
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
				worker_id = ?,
				updated_at = ?
			WHERE 
				id = ?
			`, jc.tblNameForJobs())
		log.D("FetchJobs", stmt)
		_, err := txn.Exec(stmt, JobStatusRunning, workerID, time.Now(), job.ID)
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

func (jc *JobChannel) UpdateJobForWorker(workerID string, job *Job) error {
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
	log.D("UpdateJob", stmt)
	_, err := jc.store.DB().Exec(stmt,
		job.Status,
		job.ProgressData,
		job.ResultCode,
		job.ResultData,
		job.ErrorMessage,
		time.Now(),
		job.ID, workerID)
	if err != nil {
		return err
	}
	return nil
}
