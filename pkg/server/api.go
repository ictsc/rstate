package server

import (
	_ "embed"
	"fmt"
	"math/rand"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ictsc/rstate/pkg/server/handler"
	"github.com/ictsc/rstate/pkg/server/job"
	"github.com/ictsc/rstate/pkg/utils"
)

func Launch(port string, jobWorker *job.Worker, basicPass string) {

	rand.Seed(time.Now().UnixNano())
	runtime.GOMAXPROCS(runtime.NumCPU())
	//ProbCode
	utils.Init()
	//Gin Option
	gin.SetMode(gin.ReleaseMode)

	go jobWorker.Run()

	g := gin.Default()

	//admin
	admin := g.Group("/admin", gin.BasicAuth(gin.Accounts{
		"ictsc": basicPass,
	}))

	handler.NewAdminHandler(admin, jobWorker)

	//User

	handler.NewStatusHandler(g.Group("/backend"), jobWorker)

	//Prometheus

	g.GET("/metrics", func(c *gin.Context) {
		result := fmt.Sprintf("# HELP terraform_job 走ってるJob数\n# TYPE terraform_job counter\nterraform_job %d\n", jobWorker.GetJobCount())
		c.String(200, result)
	})

	//以前のJobを読み込む

	jobWorker.LoadJob()

	g.Run(":" + port)
}
