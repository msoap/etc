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
	"sort"
	"strings"
	"sync/atomic"
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
	autoAttachTTL           = 10 // in seconds
)

type application struct {
	ctx            context.Context
	docker         *client.Client
	containers     map[string]container
	viewState      viewState
	shuffledColors bool
}

type container struct {
	id     string
	color  string
	outNum *int32
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

	if err := app.searchNewContainers(); err != nil {
		return nil, err
	}

	return &app, nil
}

func (app *application) searchNewContainers() error {
	containers, err := app.docker.ContainerList(app.ctx, types.ContainerListOptions{})
	if err != nil {
		return errors.Wrap(err, "init docker client failed")
	}

	if !app.shuffledColors {
		randomSetSeed(containers)
		rand.Shuffle(len(app.viewState.colorsTable), func(i int, j int) {
			app.viewState.colorsTable[i], app.viewState.colorsTable[j] = app.viewState.colorsTable[j], app.viewState.colorsTable[i]
		})

		app.shuffledColors = true
	}

	for _, item := range containers {
		containerName := getContainerName(item)

		if _, ok := app.containers[containerName]; ok {
			continue
		}

		log.Printf("found container: %s (%s) %s", containerName, item.ID[:shortIDLength], item.Image)

		var zero int32
		app.containers[containerName] = container{
			color:  app.viewState.getNextColor(),
			id:     item.ID,
			outNum: &zero,
		}
	}

	return nil
}

func randomSetSeed(list []types.Container) {
	names := []string{}
	for _, item := range list {
		names = append(names, getContainerName(item))
	}
	sort.Strings(names)

	h := fnv.New64()
	_, _ = io.WriteString(h, strings.Join(names, ""))

	rand.Seed(int64(h.Sum64()))
}

func (app *application) initDockerClient() error {
	if os.Getenv(dockerVersionVarName) == "" {
		if err := os.Setenv(dockerVersionVarName, minimumDockerAPIVersion); err != nil {
			return errors.Wrap(err, "set env failed")
		}
	}

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return errors.Wrap(err, "new docker client failed")
	}

	app.docker = cli

	return nil
}

func (app *application) processLogLines(logsCh chan logLine, reader io.Reader, containerName string, outType outType) {
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

	if atomic.AddInt32(app.containers[containerName].outNum, -1) == 0 {
		delete(app.containers, containerName)
		log.Printf("container %q was removed", containerName)
	}
}

func getContainerName(container types.Container) string {
	if len(container.Names) > 0 {
		return strings.TrimPrefix(container.Names[0], "/")
	} else if len(container.ID) >= shortIDLength {
		return container.ID[:shortIDLength]
	}

	return container.ID
}

func (app *application) getContainerColor(name string) string {
	if container, ok := app.containers[name]; ok {
		return container.color
	}

	return app.viewState.colorsTable[0]
}

func (app *application) printLogLine(line logLine) {
	length := len(line.containerName)
	if length > app.viewState.maxNameLength && length <= maxContainerNameLen {
		app.viewState.maxNameLength = length
	}

	fmt.Printf("%s%*s%s %s\n", ansi.ColorCode(app.getContainerColor(line.containerName)), -app.viewState.maxNameLength, line.containerName, ansi.Reset, line.log)
}

func (app *application) showDockerLogs() error {
	logsCh := make(chan logLine, 10)
	hasNewContainers := true

	for {
		if hasNewContainers {
			for containerName, container := range app.containers {
				if atomic.LoadInt32(container.outNum) > 0 {
					continue
				}

				multiplexedLogReader, err := app.docker.ContainerLogs(app.ctx, container.id, types.ContainerLogsOptions{
					ShowStdout: true,
					ShowStderr: true,
					Follow:     true,
					Since:      time.Now().Add(-sinceTimeSeconds * time.Second).Format(rfc3339LocalFormat),
				})
				if err != nil {
					return err
				}
				atomic.StoreInt32(container.outNum, 2)

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

				go app.processLogLines(logsCh, dstOutReader, containerName, outTypeStdOut)
				go app.processLogLines(logsCh, dstErrReader, containerName, outTypeStdErr)
			}

			hasNewContainers = false
		}

		select {
		case nextLine := <-logsCh:
			app.printLogLine(nextLine)
		case <-time.After(autoAttachTTL * time.Second):
			log.Print("try search new containers...")

			if err := app.searchNewContainers(); err != nil {
				return errors.Wrap(err, "failed to search new containers")
			}
			hasNewContainers = true
		}
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
