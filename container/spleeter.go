package container

import (
	"bufio"
	"context"
	"fmt"
	"github.com/ahmetalpbalkan/dexec"
	docker "github.com/fsouza/go-dockerclient"
	"os"
	"strings"
	"time"
)

const (
	SPLEETER_IMAGE = "researchdeezer/spleeter"
	SPLEETER_CONTAINER_MODEL_PATH = "MODEL_PATH=/model"
)

// Container is a spleeter docker container that can run commamnds
// to get stems
type Container struct {
	// Client for executing commands to docker
	client dexec.Docker

	// The underlying representation of the Spleeter docker container
	containerCommandExecution dexec.Execution
}

// NewContainer returns a new spleeter container which can run spleeter commands
func NewContainer(ctx context.Context) (*Container, error) {
	currentDirectory, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %v", err)
	}

	spleeterContainer, err := dexec.ByCreatingContainer(
		docker.CreateContainerOptions{
			Name: fmt.Sprintf("spleeter-for-wav-%v", time.Now().UnixNano()),
			Config: &docker.Config{
				Image: SPLEETER_IMAGE,
				Env:   []string{SPLEETER_CONTAINER_MODEL_PATH},
			},
			// Binds the contents of the data folder into the container
			// with volumes
			HostConfig: &docker.HostConfig{
				Binds: []string{
					fmt.Sprintf("%s/data/input:/input", currentDirectory),
					fmt.Sprintf("%s/data/output:/output", currentDirectory),
					fmt.Sprintf("%s/data/model:/model", currentDirectory),
				},
			},
		},
	)

	dockerClient, err := docker.NewClientFromEnv()
	if err != nil {
		return nil, err
	}

	dockerExecClient := dexec.Docker{dockerClient}

	return &Container{
		client:                    dockerExecClient,
		containerCommandExecution: spleeterContainer,
	}, nil
}

// Run splits a file in the data/input directory into
// stems using the Spleeter container
// Todo: Support specifying the model to use for splits
// Todo: Validate that songs were split in expected directory
// Todo: Upload songs to a CDN and return link to download songs to the client
func (c *Container) Run(fileName string) error {
	cmd := c.client.Command(c.containerCommandExecution, "spleeter",
		"separate",
		"-o",
		"/output",
		"-i",
		sanitizeFileName(fileName),
		"--verbose",
	)

	cmdOutputPipe, _ := cmd.StdoutPipe()
	cmdErrPipe, _ := cmd.StderrPipe()
	cmdOutputScanner := bufio.NewScanner(cmdOutputPipe)
	cmdErrScanner := bufio.NewScanner(cmdErrPipe)

	// Start command in the background
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start command: %v", err)
	}

	// Start a goroutine to read from the command's output pipe
	go func() {
		for cmdOutputScanner.Scan() {
			fmt.Println(cmdOutputScanner.Text())
		}

		if err := cmdOutputScanner.Err(); err != nil {
			fmt.Printf("Error occurred while reading command pipe output: %v", err)
		}
	}()

	// Start a goroutine to read from the command's error pipe
	go func() {
		for cmdErrScanner.Scan() {
			fmt.Println(cmdErrScanner.Text())
		}

		if err := cmdErrScanner.Err(); err != nil {
			fmt.Printf("Error occurred while reading command errror pipe output: %v", err)
		}
	}()

	// Block until command finishes
	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("failed to split with spleeter: %v", err)
	}

	return  nil
}

// sanitizeFileName replaces the data/ prefix in the tmp filename
// The data/input directory on the filesystem is mounted as /input
// in the container
func sanitizeFileName(fileName string) string {
	return strings.ReplaceAll(fileName, "data/", "")
}
