package job

import (
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gammazero/workerpool"
	"github.com/google/uuid"
	"github.com/ictsc/rstate/pkg/notifications"
	"github.com/ictsc/rstate/pkg/terraform"
	"github.com/ictsc/rstate/pkg/utils"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
)

type Worker struct {
	workerPool      *workerpool.WorkerPool
	adminWorkerPool *workerpool.WorkerPool
	terraformPath   string
	workDir         string
	env             []string
	logger          *zap.SugaredLogger
	AddTaskChannel  chan Job
	m               sync.Mutex //job list lock
	jobs            map[uuid.UUID]*Job
	counter         int
	c               *cache.Cache
	s               sync.Mutex //save file lock
	tokenCache      *cache.Cache
	stop            bool
}

type Job struct {
	Id          uuid.UUID `json:"id"`
	State       State     `json:"state"`
	CreatedTime time.Time `json:"created_time"`
	EndTime     time.Time `json:"end_time"`
	Priority    int64     `json:"priority"`
	TeamID      string    `json:"team_id"`
	ProbID      string    `json:"prob_id"`
	WorkDir     string    `json:"-"`
}

func NewWorker(maxThread int, terraformPath, workDir string, env []string, logger *zap.SugaredLogger) *Worker {
	c := cache.New(cache.NoExpiration, 5*time.Minute)
	jw := &Worker{
		workerPool:      workerpool.New(maxThread),
		adminWorkerPool: workerpool.New(1),
		terraformPath:   terraformPath,
		workDir:         workDir,
		env:             env,
		logger:          logger,
		AddTaskChannel:  make(chan Job, 100),
		jobs:            map[uuid.UUID]*Job{},
		c:               c,
		tokenCache:      cache.New(24*time.Hour, 10*time.Minute),
		stop:            false,
	}
	return jw
}

func (j *Worker) LoadJob() {
	j.m.Lock()
	defer j.m.Unlock()
	gob.Register(Job{})
	err := j.c.LoadFile(j.workDir + "/job.state")
	if err != nil {
		log.Println(err)
	}

	for _, obj := range j.c.Items() {
		job := obj.Object.(Job)
		if job.State == StateWait {
			j.AddTaskChannel <- job
		}
		if job.State == StateRunning {
			j.c.Set(job.Id.String(), &job, 24*3*time.Hour)
		}
	}
}

func (j *Worker) StopWait() {
	j.stop = true
	j.workerPool.StopWait()
}

func (j *Worker) GetState(id uuid.UUID) State {
	if job, err := j.GetJobWithID(id); err == nil {
		return job.State
	}
	return StateUnknown
}

func (j *Worker) SetState(id uuid.UUID, state State) error {
	j.s.Lock()
	defer j.s.Unlock()
	if _, ok := j.jobs[id]; !ok {
		return errors.New("set jobState Failed")
	}
	j.jobs[id].State = state

	switch state {
	case StateError:
	case StateSuccess:
		j.counter--
		j.jobs[id].EndTime = time.Now()
		break
	}
	j.c.Set(id.String(), *j.jobs[id], 24*14*time.Hour)
	j.c.SaveFile(j.workDir + "/job.state")
	return nil
}

func (j *Worker) GetJobWithID(id uuid.UUID) (*Job, error) {
	obj, ok := j.c.Get(id.String())
	if !ok {
		return nil, errors.New("get job Failed")
	}
	job := obj.(Job)
	return &job, nil
}

func (j *Worker) Run() {
	for {
		job := <-j.AddTaskChannel
		if j.stop { //signal
			return
		}
		if !j.IsJobExist(job.TeamID, job.ProbID) {
			str := fmt.Sprintf("チーム:　%s\n問題:　　%s\n", job.TeamID, job.ProbID)
			log.Println(str)
			job.State = StateWait
			job.CreatedTime = time.Now()

			j.m.Lock()
			// 追加
			j.counter++
			j.jobs[job.Id] = &job
			j.c.Set(job.Id.String(), job, 24*14*time.Hour)
			j.m.Unlock()

			//Submit Task
			wo := j.workerPool
			if utils.IsAdminTeam(job.TeamID) {
				wo = j.adminWorkerPool
			}
			wo.Submit(func() {
				j.SetState(job.Id, StateRunning)

				//terraform init
				tfclient := terraform.NewClient(j.terraformPath, j.workDir, job.TeamID, "10", true, j.env)
				j.logger.Info("Recreate Problem Start", "TeamID", job.TeamID, "ProbID", job.ProbID)
				_, targetCount, err := tfclient.RecreateFromProblemId(job.ProbID, true)
				if err != nil {
					j.logger.Errorw("Recreate Problem Error", "TeamID", job.TeamID, "ProbID", job.ProbID, "targetCount", targetCount, "error", err)
					j.SetState(job.Id, StateError)
				}
				j.SetState(job.Id, StateSuccess)
				if utils.IsAdminTeam(job.TeamID) {
					notifications.NewNotifications(job.ProbID+" - 完了", str).SendAll()
				}
			})

		} else {
			job.State = StateTaskLimit
		}
		j.s.Lock()
		j.c.Set(job.Id.String(), job, 24*14*time.Hour)
		j.c.SaveFile(j.workDir + "/job.state")
		j.s.Unlock()
	}
}

func (j *Worker) IsJobExist(teamid, probid string) bool {
	j.m.Lock()
	defer j.m.Unlock()

	if utils.IsAdminTeam(teamid) {
		// 運営チームは、admin-worker-poolでworker 1で利用する
		for _, obj := range j.c.Items() {
			job := obj.Object.(*Job)
			if job.TeamID == teamid && job.ProbID == probid && (job.State == StateWait || job.State == StateRunning) {
				return true
			}
		}
		return false
	}
	// 競技時間外はLimit
	if !utils.IsCompetitionTime(teamid) {
		return true
	}

	// 参加者1チームにつき1問リクエストできる。
	for _, obj := range j.c.Items() {
		job := obj.Object.(*Job)
		if job.TeamID == teamid && (job.State == StateWait || job.State == StateRunning) {
			return true
		}
	}
	return false
}
