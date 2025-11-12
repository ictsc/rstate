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

type SStateHandler struct {
	jobWorker *job.Worker
}

func NewSStateHandler(jobw *job.Worker) *SStateHandler {
	return &SStateHandler{
		jobWorker: jobw,
	}
}

// GetQueueStatus implements ServerInterface
func (h *SStateHandler) GetQueueStatus(c *gin.Context) {
	jobs := h.jobWorker.GetJobList("")
	inQueue := make([]string, 0)

	for _, j := range jobs {
		if j.State == job.StateStringWait || j.State == job.StateStringRunning {
			// TeamID is already without "team" prefix in ResponseJob
			// Convert probID to uppercase for API response
			key := j.TeamID + "_" + strings.ToUpper(j.ProbID)
			inQueue = append(inQueue, key)
		}
	}

	response := QueueStatus{
		InQueue: &inQueue,
	}

	c.JSON(http.StatusOK, response)
}

// PostRedeploy implements ServerInterface
func (h *SStateHandler) PostRedeploy(c *gin.Context) {
	var req RedeployRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "無効なリクエストフォーマットです")
		return
	}

	// Validate team_id format (2 digits)
	if len(req.TeamId) != 2 {
		respondError(c, http.StatusBadRequest, "team_idは2桁の数字である必要があります")
		return
	}

	// Validate problem_id (minimum 3 characters)
	if len(req.ProblemId) < 3 {
		respondError(c, http.StatusBadRequest, "無効なproblem_idです")
		return
	}

	teamID := "team" + req.TeamId
	probID := strings.ToLower(req.ProblemId)

	// Check if job already exists
	if _, exist := h.jobWorker.IsJobExist(teamID, probID); exist {
		respondError(c, http.StatusTooManyRequests, "同じチームの問題のリクエストは既にキューに存在します")
		return
	}

	// Create new job
	id, err := uuid.NewUUID()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "内部エラーが発生しました")
		return
	}

	now := time.Now()
	priority := now.UnixNano()
	if utils.IsAdminTeam(teamID) {
		priority = 1
	}

	newJob := job.Job{
		TeamID:      teamID,
		ProbID:      probID,
		CreatedTime: utils.ToTimePtr(now),
		Priority:    priority,
		Id:          id,
	}

	// Check queue capacity (assuming channel size of 100)
	select {
	case h.jobWorker.AddTaskChannel <- newJob:
		respondSuccess(c, http.StatusCreated, "再展開リクエストを受け付けました")
	default:
		respondError(c, http.StatusTooManyRequests, "リクエストキューが満杯です")
	}
}

// GetTeamStatus implements ServerInterface
func (h *SStateHandler) GetTeamStatus(c *gin.Context, teamID string) {
	teamIDWithPrefix := "team" + teamID
	jobs := h.jobWorker.GetJobList(teamIDWithPrefix)

	statuses := make(map[string]RedeployStatus)

	for _, j := range jobs {
		status := responseJobToRedeployStatus(j)
		// Convert probID to uppercase for API response
		statuses[strings.ToUpper(j.ProbID)] = status
	}

	c.JSON(http.StatusOK, statuses)
}

// GetProblemStatus implements ServerInterface
func (h *SStateHandler) GetProblemStatus(c *gin.Context, teamID string, problemID string) {
	teamIDWithPrefix := "team" + teamID
	// Convert problemID to lowercase for internal comparison
	probIDLower := strings.ToLower(problemID)

	jobs := h.jobWorker.GetJobList(teamIDWithPrefix)

	for _, j := range jobs {
		if j.ProbID == probIDLower {
			status := responseJobToRedeployStatus(j)
			c.JSON(http.StatusOK, status)
			return
		}
	}

	respondError(c, http.StatusNotFound, "指定されたチームIDと問題IDの状態は見つかりません")
}

// Helper functions

func responseJobToRedeployStatus(j *job.ResponseJob) RedeployStatus {
	var statusStr RedeployStatusStatus
	var message string

	var updatedAt *time.Time
	switch j.State {
	case job.StateStringWait:
		statusStr = RedeployStatusStatusQueuing
		message = "再展開リクエストを受け取りました"
	case job.StateStringRunning:
		statusStr = RedeployStatusStatusCreating
		message = "再展開を実行中です"
	case job.StateStringSuccess:
		statusStr = RedeployStatusStatusRunning
		// Convert probID to uppercase for display
		message = "チーム " + j.TeamID + " の問題 " + strings.ToUpper(j.ProbID) + " のリソースが正常に再展開されました"

		if j.EndTime > 0 {
			updatedAt = utils.ToTimePtr(time.Unix(0, j.EndTime))
		}
	case job.StateStringError:
		statusStr = RedeployStatusStatusError
		message = "再展開中にエラーが発生しました"
	default:
		statusStr = RedeployStatusStatusError
		message = "不明な状態です"
	}

	// Convert Unix nanoseconds to time.Time
	if updatedAt == nil && j.CreatedTime > 0 {
		t := time.Unix(0, j.CreatedTime)
		updatedAt = &t
	}

	return RedeployStatus{
		Status:    &statusStr,
		Message:   &message,
		UpdatedAt: updatedAt,
	}
}

func respondError(c *gin.Context, statusCode int, message string) {
	status := "error"
	c.JSON(statusCode, Error{
		Status:  &status,
		Message: &message,
	})
}

func respondSuccess(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, gin.H{
		"status":  "accepted",
		"message": message,
	})
}
