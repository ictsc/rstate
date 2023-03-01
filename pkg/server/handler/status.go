package handler

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ictsc/rstate/pkg/server/job"
	"github.com/ictsc/rstate/pkg/utils"
)

type StatusHandler struct {
	rg *gin.RouterGroup

	jobWorker *job.Worker
}

type RecreateProblemStatusResponse struct {
	Available      bool       `json:"available"`
	Created_time   *time.Time `json:"created_time"`
	Completed_time *time.Time `json:"completed_time"`
}

func NewStatusHandler(rg *gin.RouterGroup, jobw *job.Worker) *StatusHandler {
	sh := &StatusHandler{
		rg:        rg,
		jobWorker: jobw,
	}
	//	sh.rg.GET("/:token", sh.statusHtml)
	//	sh.rg.GET("/:token/list", sh.jobListAPI)
	sh.rg.GET("/:teamid/:probcode", sh.GetStatusWithParam)
	sh.rg.POST("/:teamid/:probcode", sh.postJob)
	return sh
}

func (sh *StatusHandler) statusHtml(c *gin.Context) {
	token := c.Param("token")
	teamId := sh.jobWorker.GetTeamIdFromToken(token)
	if !strings.HasPrefix(teamId, "team") {
		c.Status(403)
		return
	}
	if utils.IsAdminTeam(teamId) {
		c.Data(http.StatusOK, "text/html", teamHtml)
		return
	}
	// 準備期間中はページ(ダミーデータ)を返す。
	if utils.IsPreparatoryPhase(teamId) {
		c.Data(http.StatusOK, "text/html", teamHtml)
		return
	}
	// 競技時間外はページを返さない。
	if !utils.IsCompetitionTime(teamId) {
		c.AbortWithStatus(403)
		return
	}
	c.Data(http.StatusOK, "text/html", teamHtml)
}

func (sh *StatusHandler) jobListAPI(c *gin.Context) {
	token := c.Param("token")
	teamId := sh.jobWorker.GetTeamIdFromToken(token)
	if !strings.HasPrefix(teamId, "team") {
		c.AbortWithStatus(403)
		return
	}
	if utils.IsAdminTeam(teamId) {
		c.JSON(200, sh.jobWorker.GetJobList(teamId))
		return
	}
	// 準備期間中はダミーデータを返す。
	if utils.IsPreparatoryPhase(teamId) {
		var res = make([]job.ResponseJob, 0)
		res = append(res, job.ResponseJob{
			Id:          uuid.New(),
			State:       "開始待ち",
			CreatedTime: time.Now().Add(3 * time.Minute).UnixNano(),
			TeamID:      teamId,
			ProbID:      "dummy code",
		})
		res = append(res, job.ResponseJob{
			Id:          uuid.New(),
			State:       "実行中",
			CreatedTime: time.Now().Add(2 * time.Minute).UnixNano(),
			TeamID:      teamId,
			ProbID:      "dummy code",
		})
		res = append(res, job.ResponseJob{
			Id:          uuid.New(),
			State:       "Limit",
			CreatedTime: time.Now().Add(1 * time.Minute).UnixNano(),
			TeamID:      teamId,
			ProbID:      "dummy code",
		})
		res = append(res, job.ResponseJob{
			Id:          uuid.New(),
			State:       "終了",
			CreatedTime: time.Now().UnixNano(),
			TeamID:      teamId,
			ProbID:      "dummy code",
		})

		c.JSON(200, res)
		return
	}
	// 競技時間外はデータを返さない。
	if !utils.IsCompetitionTime(teamId) {
		c.AbortWithStatus(403)
		return
	}
	c.JSON(200, sh.jobWorker.GetJobList(teamId))
}

func (sh *StatusHandler) postJob(c *gin.Context) {
	teamId := c.Param("teamid")
	probId := c.Param("probcode")

	if !strings.HasPrefix(teamId, "team") || len(probId) < 3 {
		c.String(400, "BadRequest!")
		return
	}

	id, err := uuid.NewUUID()
	if err != nil {
		log.Println(err)
	}

	now := time.Now()
	priority := now.UnixNano()
	if utils.IsAdminTeam(teamId) {
		priority = 1
	}

	job := job.Job{
		TeamID:      teamId,
		ProbID:      probId,
		CreatedTime: utils.ToTimePtr(time.Now()),
		Priority:    priority,
		Id:          id,
	}

	if _, exist := sh.jobWorker.IsJobExist(teamId, probId); exist {
		c.String(http.StatusConflict, "")
		return
	}

	sh.jobWorker.AddTaskChannel <- job
	c.JSONP(http.StatusOK, job)
}

func (sh *StatusHandler) GetStatusWithParam(c *gin.Context) {
	teamId := c.Param("teamid")
	probCode := c.Param("probcode")
	job, exist := sh.jobWorker.IsJobExist(teamId, probCode)
	if exist {
		response := RecreateProblemStatusResponse{
			Available:      false,
			Created_time:   job.CreatedTime,
			Completed_time: job.EndTime,
		}
		c.JSONP(200, response)
		return
	}
	//Available
	if job != nil {
		response := RecreateProblemStatusResponse{
			Available:      true,
			Created_time:   job.CreatedTime,
			Completed_time: job.EndTime,
		}
		c.JSONP(200, response)
		return
	}
	response := RecreateProblemStatusResponse{
		Available:      true,
		Created_time:   nil,
		Completed_time: nil,
	}
	c.JSONP(200, response)

}
