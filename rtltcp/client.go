package rtltcp

import (
	"bufio"
	"context"
	"io"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

// We store the exec.Command method in a variable to allow us later on to replace this method while testing.
// This makes the life easier as we can control what this method is doing.
var execCmd = exec.Command

type Client interface {
	Run(ctx context.Context) chan error
}

func NewClient() Client {
	c := client{
		done: make(chan struct{}),
	}

	return &c
}

type client struct {
	done    chan struct{}
	errChan chan error
}

func (c *client) Run(ctx context.Context) chan error {
	errChan := make(chan error)
	rtlTCPCmd := execCmd("rtl_tcp")

	stdOutReader, err := rtlTCPCmd.StdoutPipe()
	if err != nil {
		errChan <- err
		return errChan
	}

	go c.processStdOutPipe(stdOutReader)

	stdErrReader, err := rtlTCPCmd.StderrPipe()
	if err != nil {
		errChan <- err
		return errChan
	}

	go c.processStdErrPipe(stdErrReader)

	err = rtlTCPCmd.Start()
	if err != nil {
		errChan <- err
		return errChan
	}

	go func() {
		err = rtlTCPCmd.Wait()
		if err != nil {
			log.Errorf("Error while waiting for command %s to finish.", rtlTCPCmd.String())
		}

		close(c.done)
	}()

	go func() {
		for {
			select {
			case <-c.done:
				log.Infof("Command %s finished runnig.", rtlTCPCmd.String())
				return
			case <-ctx.Done():
				log.Infof("Context was terminated. Stopping %s process.", rtlTCPCmd.String())
				killErr := rtlTCPCmd.Process.Kill()
				if killErr != nil {
					log.Errorf("Error while killing process: %s. %s", rtlTCPCmd.Path, err)
				}
				return
			}
		}
	}()

	return errChan
}

func (c *client) processStdOutPipe(reader io.ReadCloser) {
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		msg := scanner.Text()

		log.Debug(msg)
	}
}

func (c *client) processStdErrPipe(reader io.ReadCloser) {
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		msg := scanner.Text()

		log.Error(msg)
	}
}
