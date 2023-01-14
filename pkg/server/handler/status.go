package handler

import (
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

func NewStatusHandler(rg *gin.RouterGroup, jobw *job.Worker) *StatusHandler {
	sh := &StatusHandler{
		rg:        rg,
		jobWorker: jobw,
	}
	sh.rg.GET("/:token", sh.statusHtml)
	sh.rg.GET("/:token/list", sh.jobListAPI)
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
