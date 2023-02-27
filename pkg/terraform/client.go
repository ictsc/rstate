package terraform

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
	_ "time/tzdata"
)

type Client struct {
	parallelism string
	workDir     string
	options     string
	path        string
	stdout      bool
	env         []string
	IsDebug     bool
}

func NewClient(path, workDir, workspace, parallelism string, stdout bool, envs []string) *Client {
	mergeEnv := append(os.Environ(), envs...)
	mergeEnv = append(mergeEnv, "TF_WORKSPACE="+workspace)
	mergeEnv = append(mergeEnv, "TF_LOG=ERROR")
	logPath := fmt.Sprintf("TF_LOG_PATH=/app/recreate-logs/%s-%s.json", workspace, time.Now().Format("2006-01-02-15-04-05"))
	mergeEnv = append(mergeEnv, logPath)
	c := &Client{
		parallelism: parallelism,
		workDir:     workDir,
		path:        path,
		stdout:      stdout,
		env:         mergeEnv,
	}
	log.Printf("%#v", c.env)
	str, err := c.init()
	if err != nil {
		panic(str)
	}
	return c
}

func (c *Client) init() (string, error) {

	cmd := exec.Command(c.path, "init")
	fmt.Println(cmd.String())
	cmd.Env = c.env
	cmd.Dir = c.workDir
	stdoutStderr, err := cmd.CombinedOutput()
	return string(stdoutStderr), err
}

func (c *Client) validate() (string, error) {

	cmd := exec.Command(c.path, "validate")
	fmt.Println(cmd.String())
	cmd.Env = c.env
	cmd.Dir = c.workDir
	stdoutStderr, err := cmd.CombinedOutput()
	return string(stdoutStderr), err
}

func (c *Client) plan(opt string) (string, error) {
	args := strings.Fields("plan -input=false -detailed-exitcode -lock-timeout=0s -lock=true -parallelism=" + c.parallelism + " -refresh=false " + opt)
	cmd := exec.Command(c.path, args...)
	fmt.Println(cmd.String())
	cmd.Env = c.env
	cmd.Dir = c.workDir

	return c.do(cmd)
}

func (c *Client) apply(opt string) (string, error) {
	args := strings.Fields("apply -auto-approve  -no-color -input=false -lock-timeout=0s -lock=true -parallelism=" + c.parallelism + " -refresh=false " + opt)
	cmd := exec.Command(c.path, args...)
	fmt.Println(cmd.String())
	cmd.Env = c.env
	cmd.Dir = c.workDir

	return c.do(cmd)
}

func (c Client) do(cmd *exec.Cmd) (string, error) {
	//Console stdout
	if c.stdout {
		stdout, err := cmd.StdoutPipe()
		cmd.Stderr = os.Stderr
		if err != nil {
			fmt.Println(err)
			return "", err
		}

		cmd.Start()

		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
		cmd.Wait()
		return "", err
	}

	out, err := cmd.CombinedOutput()
	return string(out), err
}

func (c *Client) GetResourceTargetId(moduleName string) (string, int, error) {
	resource := ""
	resourceCount := 0

	cmd := exec.Command(c.path, c.options, "state", "list", moduleName)
	cmd.Env = c.env
	cmd.Dir = c.workDir
	stdoutStderr, err := cmd.CombinedOutput()

	stdoutStr := string(stdoutStderr)

	stdOutSplit := strings.Split(stdoutStr, "\n")
	for _, resourceId := range stdOutSplit {
		// Remove data source
		if !strings.Contains(resourceId, "data.sakuracloud") &&
			!strings.Contains(resourceId, "sakuracloud_switch") && resourceId != "" {
			resource += " -replace=" + resourceId
			resourceCount++
		}
	}

	if err != nil {
		log.Println(stdoutStr)
	}

	return resource, resourceCount, err
}
