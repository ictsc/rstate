package job

import (
	"time"

	"github.com/google/uuid"
)

type ResponseJob struct {
	Id          uuid.UUID `json:"id"`
	State       string    `json:"state"`
	CreatedTime int64     `json:"created_time"`
	EndTime     int64     `json:"end_time"`
	TeamID      string    `json:"team_id"`
	ProbID      string    `json:"prob_id"`
}

func (j *Worker) GetStatusToken(teamId string) string {
	obj, ok := j.tokenCache.Get(teamId)
	if ok {
		return obj.(string)
	}
	//generate
	uuidObj, err := uuid.NewUUID()
	if err != nil {
		return ""
	}
	id := uuidObj.String()
	j.tokenCache.Set(teamId, id, 72*time.Hour)
	j.tokenCache.Set(id, teamId, 72*time.Hour)
	return id
}

func (j *Worker) GetTeamIdFromToken(token string) string {
	obj, ok := j.tokenCache.Get(token)
	if ok {
		return obj.(string)
	}
	return ""
}

func (j *Worker) GetJobList(teamId string) []*ResponseJob {

	var result []*ResponseJob
	for _, object := range j.c.Items() {
		job := object.Object.(Job)
		var stateString string
		switch job.State {
		case StateWait:
			stateString = StateStringWait
		case StateRunning:
			stateString = StateStringRunning
		case StateSuccess:
			stateString = StateStringSuccess
		case StateError:
			stateString = StateStringError
		case StateTaskLimit:
			stateString = StateStringTaskLimit
		default:
			stateString = StateStringUnknown
		}
		var createdtime, endtime int64

		if job.CreatedTime == nil {
			createdtime = time.Now().UnixNano()
		} else {
			createdtime = job.CreatedTime.UnixNano()
		}

		if job.EndTime == nil {
			endtime = time.Now().UnixNano()
		} else {
			endtime = job.EndTime.UnixNano()
		}

		res := &ResponseJob{
			Id:          job.Id,
			State:       stateString,
			CreatedTime: createdtime,
			EndTime:     endtime,
			TeamID:      job.TeamID[4:], // Trim "team" prefix
			ProbID:      job.ProbID,
		}
		if job.TeamID == teamId || teamId == "" {
			result = append(result, res)
		}
	}

	return result
}

func (j *Worker) FailedRunningJobs() {
	for _, object := range j.c.Items() {
		job := object.Object.(Job)
		switch job.State {
		case StateRunning:
			j.SetState(job.Id, StateError)
		}
	}
	j.c.SaveFile(j.workDir + "/job.state")
}

func (j *Worker) GetJobCount() int {
	j.m.Lock()
	defer j.m.Unlock()
	return j.counter
}
