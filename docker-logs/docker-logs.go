/*
Utility for show docker logs from multiple containers in "follow" mode (like "docker-compose logs -f").

Install:
	go get -u github.com/msoap/etc/docker-logs

*/
package main

import (
	"bufio"
	"context"
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/mgutz/ansi"
)

const (
	minimumDockerAPIVersion = "1.12"
	dockerVersionVarName    = "DOCKER_API_VERSION"
	timeout                 = 60
	sinceTimeSeconds        = 600
	rfc3339LocalFormat      = "2006-01-02T15:04:05"
	shortIDLength           = 12
)

type application struct {
	docker *client.Client
}

type outType int

const (
	outTypeStdOut = iota
	outTypeStdErr
)

func (o outType) String() string {
	switch o {
	case outTypeStdOut:
		return "STDOUT"
	case outTypeStdErr:
		return "STDERR"
	default:
		return "UNKNOWN"
	}
}

type logLine struct {
	containerName string
	log           string
	outType       outType
}

func (ll logLine) String() string {
	return fmt.Sprintf("%s %s", ansi.Color(ll.containerName, getColorByHash(ll.containerName)), ll.log)
}

var colors = [...]string{
	"red",
	"green",
	"yellow",
	"blue",
	"magenta",
	"cyan",
	"red+h",
	"green+h",
	"yellow+h",
	"blue+h",
	"magenta+h",
	"cyan+h",
	"white+h",
}

func getColorByHash(in string) string {
	h := fnv.New64()
	io.WriteString(h, in)
	i := int(h.Sum(nil)[0]) % len(colors)
	return colors[i]
}

func (a *application) initDockerClient() error {
	if os.Getenv(dockerVersionVarName) == "" {
		if err := os.Setenv(dockerVersionVarName, minimumDockerAPIVersion); err != nil {
			return err
		}
	}

	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	a.docker = cli

	return nil
}

func processLogLines(logsCh chan logLine, reader io.Reader, containerName string, outType outType) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		nextLine := scanner.Text()
		logsCh <- logLine{
			containerName: containerName,
			log:           nextLine,
			outType:       outType,
		}
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		log.Printf("%s: failed to read %s log line: %s", containerName, outType, err)
	}

	log.Printf("closed %s log for container %s", outType, containerName)
}

func getContainerName(container types.Container) string {
	if len(container.Names) > 0 {
		return strings.TrimPrefix(container.Names[0], "/")
	} else if len(container.ID) >= shortIDLength {
		return container.ID[:shortIDLength]
	}

	return container.ID
}

func (a *application) showDockerLogs() error {
	ctx := context.Background()

	containers, err := a.docker.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return fmt.Errorf("init docker client failed: %s", err)
	}

	logsCh := make(chan logLine, 10)

	for _, container := range containers {
		containerName := getContainerName(container)
		log.Printf("found container: %s (%s) %s", containerName, container.ID[:shortIDLength], container.Image)

		multiplexedLogReader, err := a.docker.ContainerLogs(ctx, container.ID, types.ContainerLogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Follow:     true,
			Since:      time.Now().Add(-sinceTimeSeconds * time.Second).Format(rfc3339LocalFormat),
		})
		if err != nil {
			return err
		}

		dstOutReader, dstOutWriter := io.Pipe()
		dstErrReader, dstErrWriter := io.Pipe()

		go func() {
			closeStdErrFn := func() {
				log.Printf("do close %q stderr logs", containerName)
				if err := dstErrWriter.Close(); err != nil {
					log.Printf("close stderr for %q failed: %s", containerName, err)
				}
			}

			if _, err := stdcopy.StdCopy(dstOutWriter, dstErrWriter, multiplexedLogReader); err != nil {
				if err != io.EOF {
					log.Printf("demultiplex %q via StdCopy failed: %q, try parse simple stdout", containerName, err)
					closeStdErrFn()

					// try parse out without headers
					if _, err := io.Copy(dstOutWriter, multiplexedLogReader); err != nil {
						if err == io.EOF {
							log.Printf("%q logs closed", containerName)
						} else {
							log.Printf("copy %q logs failed: %s", containerName, err)
						}
					}
				}
			}

			log.Printf("do close %q stdout logs", containerName)
			if err := dstOutWriter.Close(); err != nil {
				log.Printf("close stdout for %q failed: %s", containerName, err)
			}
			closeStdErrFn()
		}()

		go processLogLines(logsCh, dstOutReader, containerName, outTypeStdOut)
		go processLogLines(logsCh, dstErrReader, containerName, outTypeStdErr)
	}

	for nextLine := range logsCh {
		fmt.Println(nextLine)
	}

	return nil
}

func main() {
	app := application{}

	if err := app.initDockerClient(); err != nil {
		log.Printf("init docker client failed: %s", err)
		return
	}

	if err := app.showDockerLogs(); err != nil {
		log.Print(err)
		return
	}
}
