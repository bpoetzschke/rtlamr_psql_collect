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
	Run(ctx context.Context, startedChan chan struct{}) error
}

func NewClient() Client {
	c := client{
		done:    make(chan struct{}),
		errChan: make(chan error),
	}

	return &c
}

type client struct {
	done    chan struct{}
	errChan chan error
}

func (c *client) Run(ctx context.Context, startedChan chan struct{}) error {
	rtlTCPCmd := execCmd("rtl_tcp")

	stdOutReader, err := rtlTCPCmd.StdoutPipe()
	if err != nil {
		return err
	}

	go c.processStdOutPipe(stdOutReader)

	stdErrReader, err := rtlTCPCmd.StderrPipe()
	if err != nil {
		return err
	}

	go c.processStdErrPipe(stdErrReader)

	err = rtlTCPCmd.Start()
	if err != nil {
		return err
	}

	// Signal that we started the process
	startedChan <- struct{}{}

	go func() {
		err = rtlTCPCmd.Wait()
		if err != nil {
			log.Errorf("Error while waiting for command %s to finish.", rtlTCPCmd.String())
			c.errChan <- err
			close(c.errChan)
		} else {
			close(c.done)
		}
	}()

	for {
		select {
		case <-c.done:
			log.Infof("Command %s finished runnig.", rtlTCPCmd.String())
			return nil
		case <-ctx.Done():
			log.Infof("Context was terminated. Stopping %s process.", rtlTCPCmd.String())
			killErr := rtlTCPCmd.Process.Kill()
			if killErr != nil {
				log.Errorf("Error while killing process: %s. %s", rtlTCPCmd.Path, err)
			}
		case err := <-c.errChan:
			return err
		}
	}
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
