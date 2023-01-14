package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ictsc/rstate/pkg/server"
	"github.com/ictsc/rstate/pkg/server/job"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

var (
	configPath string
	config     Config
	loggers    *zap.SugaredLogger
)

func init() {

	flag.StringVar(&configPath, "config", "config.yaml", "config path")
	flag.Parse()
	f, err := os.Open(configPath)
	if err != nil {
		log.Fatalf("Failed to open config file `%s`.", configPath)
	}

	if err := yaml.NewDecoder(f).Decode(&config); err != nil {
		f.Close()
		log.Fatalf("Failed to decode config.")
	}
	log.Println(config.Terraform.Options.WorkingDirectory)
	var cfg = zap.Config{}
	atomicLevel := zap.NewAtomicLevel()

	err = atomicLevel.UnmarshalText([]byte(config.LogLevel))
	if err != nil {
		return
	}

	cfg.Level = atomicLevel
	cfg.Encoding = "json"
	cfg.OutputPaths = []string{"stdout"}

	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	loggers = logger.Sugar()
}

func main() {
	terraformOpt := config.Terraform.Options
	terraformEnv := config.Terraform.Secrets
	env := []string{
		"SAKURACLOUD_ACCESS_TOKEN=" + terraformEnv.SakuraCloudAccessToken,
		"SAKURACLOUD_ACCESS_TOKEN_SECRET=" + terraformEnv.SakuraCloudAccessTokenSecret,
		"AWS_ACCESS_KEY_ID=" + terraformEnv.AwsAccessKeyID,
		"AWS_SECRET_ACCESS_KEY=" + terraformEnv.AwsSecretAccessKey,
		"CLOUDFLARE_API_TOKEN=" + terraformEnv.CloudflareAPIToken,
		"TF_VAR_infra_password=" + config.AdminPass,
	}
	worker := job.NewWorker(config.Worker.MaxThread, terraformOpt.Path, terraformOpt.WorkingDirectory, env, loggers)
	go server.Launch("8089", worker, config.AdminPass)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)

	<-quit
	loggers.Sync()
	worker.StopWait()
	worker.FailedRunningJobs()
}
