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
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/mgutz/ansi"
	"github.com/pkg/errors"
)

const (
	minimumDockerAPIVersion = "1.12"
	dockerVersionVarName    = "DOCKER_API_VERSION"
	sinceTimeSeconds        = 600
	rfc3339LocalFormat      = "2006-01-02T15:04:05"
	shortIDLength           = 12
	maxContainerNameLen     = 25
)

type application struct {
	ctx        context.Context
	docker     *client.Client
	containers map[string]container
	viewState  viewState
}

type container struct {
	id    string
	color string
}

type viewState struct {
	colorsTable   []string
	currColorNum  int
	maxNameLength int
}

func (vs *viewState) getNextColor() string {
	color := vs.colorsTable[vs.currColorNum]
	vs.currColorNum++
	if vs.currColorNum > len(vs.colorsTable)-1 {
		vs.currColorNum = 0
	}

	return color
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

func newApplication() (*application, error) {
	ctx := context.Background()

	app := application{
		ctx:        ctx,
		containers: map[string]container{},
		viewState: viewState{
			colorsTable: []string{
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
			},
		},
	}

	if err := app.initDockerClient(); err != nil {
		return nil, err
	}

	containers, err := app.docker.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "init docker client failed")
	}

	randomSetSeed(containers)
	rand.Shuffle(len(app.viewState.colorsTable), func(i int, j int) {
		app.viewState.colorsTable[i], app.viewState.colorsTable[j] = app.viewState.colorsTable[j], app.viewState.colorsTable[i]
	})

	for _, item := range containers {
		containerName := getContainerName(item)

		log.Printf("found container: %s (%s) %s", containerName, item.ID[:shortIDLength], item.Image)

		app.containers[containerName] = container{
			color: app.viewState.getNextColor(),
			id:    item.ID,
		}

		if len(containerName) > app.viewState.maxNameLength && len(containerName) <= maxContainerNameLen {
			app.viewState.maxNameLength = len(containerName)
		}
	}

	return &app, nil
}

func randomSetSeed(list []types.Container) {
	h := fnv.New64()
	for _, item := range list {
		_, _ = io.WriteString(h, item.ID)
	}

	rand.Seed(int64(h.Sum64()))
}

func (a *application) initDockerClient() error {
	if os.Getenv(dockerVersionVarName) == "" {
		if err := os.Setenv(dockerVersionVarName, minimumDockerAPIVersion); err != nil {
			return errors.Wrap(err, "set env failed")
		}
	}

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return errors.Wrap(err, "new docker client failed")
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

func (a *application) getContainerColor(name string) string {
	if container, ok := a.containers[name]; ok {
		return container.color
	}

	return a.viewState.colorsTable[0]
}

func (a *application) printLogLine(line logLine) {
	fmt.Printf("%s%*s%s %s\n", ansi.ColorCode(a.getContainerColor(line.containerName)), -a.viewState.maxNameLength, line.containerName, ansi.Reset, line.log)
}

func (a *application) showDockerLogs() error {
	logsCh := make(chan logLine, 10)

	for containerName, container := range a.containers {
		multiplexedLogReader, err := a.docker.ContainerLogs(a.ctx, container.id, types.ContainerLogsOptions{
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
				log.Printf("do close %q %s logs", containerName, outType(outTypeStdErr))
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

			log.Printf("do close %q %s logs", containerName, outType(outTypeStdOut))
			if err := dstOutWriter.Close(); err != nil {
				log.Printf("close stdout for %q failed: %s", containerName, err)
			}
			closeStdErrFn()
		}()

		go processLogLines(logsCh, dstOutReader, containerName, outTypeStdOut)
		go processLogLines(logsCh, dstErrReader, containerName, outTypeStdErr)
	}

	for nextLine := range logsCh {
		a.printLogLine(nextLine)
	}

	return nil
}

func main() {
	app, err := newApplication()
	if err != nil {
		log.Printf("init application failed: %s", err)
		return
	}

	if err := app.showDockerLogs(); err != nil {
		log.Print(err)
		return
	}
}
