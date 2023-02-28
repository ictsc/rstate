package handler

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ictsc/rstate/pkg/server/job"
	"github.com/ictsc/rstate/pkg/utils"
)

type AdminHandler struct {
	rg *gin.RouterGroup

	jobWorker *job.Worker
}

func NewAdminHandler(rg *gin.RouterGroup, jobw *job.Worker) *AdminHandler {

	ah := &AdminHandler{
		rg:        rg,
		jobWorker: jobw,
	}
	rg.GET("/", ah.index)
	rg.GET("/list", ah.getAll)
	rg.GET("/list/:id", ah.get)
	rg.GET("/token/:teamid", ah.team)
	rg.POST("/postJob", ah.postJob)
	return ah
}

func (ah *AdminHandler) index(c *gin.Context) {
	c.Data(http.StatusOK, "text/html", teamHtml)
}

func (ah *AdminHandler) getAll(c *gin.Context) {
	c.JSON(200, ah.jobWorker.GetJobList(""))
}

func (ah *AdminHandler) get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.AbortWithStatus(400)
		return
	}

	uid, err := uuid.Parse(id)
	if err != nil {
		c.AbortWithStatus(400)
		return
	}
	job, err := ah.jobWorker.GetJobWithID(uid)
	if err != nil {
		c.AbortWithStatus(400)
	}
	c.JSON(200, job)
}

func (ah *AdminHandler) team(c *gin.Context) {
	teamid := c.Param("teamid")
	if !strings.HasPrefix(teamid, "team") {
		c.Status(400)
		return
	}
	token := ah.jobWorker.GetStatusToken(teamid)
	url := fmt.Sprintf("%s%s%s", os.Getenv("root_url"), "/status/", token)
	c.String(200, url)
}

func (ah *AdminHandler) postJob(c *gin.Context) {
	teamId := c.PostForm("team_id")
	probId := c.PostForm("prob_id")

	if !strings.HasPrefix(teamId, "team") || len(probId) < 3 {
		c.String(400, "BadRequest!")
		return
	}
	token := ah.jobWorker.GetStatusToken(teamId)
	url := fmt.Sprintf("%s%s%s", os.Getenv("root_url"), "/status/", token)

	// 準備期間中はURLだけを返す, 運営用のチームは関係ない。
	if utils.IsPreparatoryPhase(teamId) {
		c.String(400, url)
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
		CreatedTime: time.Now(),
		Priority:    priority,
		Id:          id,
	}

	if _, exist := ah.jobWorker.IsJobExist(teamId, probId); exist {
		c.String(400, url)
		return
	}

	ah.jobWorker.AddTaskChannel <- job
	c.JSONP(200, job)
}
