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
			stateString = "開始待ち"
			break
		case StateRunning:
			stateString = "実行中"
			break
		case StateSuccess:
			stateString = "終了"
			break
		case StateError:
			stateString = "エラー"
			break
		case StateTaskLimit:
			stateString = "Limit"
			break
		default:
			stateString = "Unknown State"
			break
		}
		var createdtime, endtime int64
		if job.CreatedTime == nil {
			createdtime = time.Now().UnixNano()
		}
		if job.EndTime == nil {
			endtime = time.Now().UnixNano()
		}
		res := &ResponseJob{
			Id:          job.Id,
			State:       stateString,
			CreatedTime: createdtime,
			EndTime:     endtime,
			TeamID:      job.TeamID[4:],
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
			break
		default:
			break
		}
	}
	j.c.SaveFile(j.workDir + "/job.state")
}

func (j *Worker) GetJobCount() int {
	j.m.Lock()
	defer j.m.Unlock()
	return j.counter
}
