package rtlamrclient

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/bpoetzschke/rtlamr_psql_collect/repositories"
	"github.com/caarlos0/env/v6"
	log "github.com/sirupsen/logrus"
)

// We store the exec.Command method in a variable to allow us later on to replace this method while testing.
// This makes the life easier as we can control what this method is doing.
var execCmd = exec.Command

type Client interface {
	Run(ctx context.Context) error
}

type rtlAMRConfig struct {
	FilterIDs string `env:"RTLAMR_FILTERID"`
}

type client struct {
	repo    repositories.RTLAMRRepo
	cmdArgs []string
	done    chan struct{}
}

func New(repo repositories.RTLAMRRepo) (Client, error) {
	rtl := client{
		repo: repo,
		done: make(chan struct{}),
	}
	err := rtl.setup()
	if err != nil {
		return nil, err
	}

	return &rtl, nil
}

func (r *client) setup() error {
	r.cmdArgs = []string{
		"-format=json",
		"-unique=true",
	}

	cfg := rtlAMRConfig{}
	if err := env.Parse(&cfg); err != nil {
		return err
	}

	if cfg.FilterIDs != "" {
		log.Debugf("Using rltamr with filter ids: %s", cfg.FilterIDs)
		r.cmdArgs = append(r.cmdArgs, fmt.Sprintf("-filterid=%s", cfg.FilterIDs))
	}

	return nil
}

func (r client) Run(ctx context.Context) error {
	rtlAMRCmd := execCmd("rtlamr", r.cmdArgs...)
	rtlAMRCmd.Env = []string{}
	cmdReader, err := rtlAMRCmd.StdoutPipe()
	if err != nil {
		return err
	}

	go r.processStdoutPipe(cmdReader)

	cmdErrReader, err := rtlAMRCmd.StderrPipe()
	if err != nil {
		return err
	}

	go r.processStderrPipe(cmdErrReader)

	log.Debugf("Run command: %s", rtlAMRCmd.String())

	err = rtlAMRCmd.Start()
	if err != nil {
		return err
	}

	go func() {
		err = rtlAMRCmd.Wait()
		if err != nil {
			log.Errorf("Error while waiting for command %s to finish", rtlAMRCmd.String())
		}

		close(r.done)
	}()

	for {
		select {
		case <-r.done:
			log.Infof("Command %s finished running.", rtlAMRCmd.String())
			return nil
		case <-ctx.Done():
			log.Infof("Context was terminated. Stopping %s process.", rtlAMRCmd.Path)
			killErr := rtlAMRCmd.Process.Kill()
			if killErr != nil {
				log.Errorf("Error while killing process %s. %s", rtlAMRCmd.Path, err)
			}
			return nil
		}
	}
}

func (r client) processStdoutPipe(reader io.ReadCloser) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		rawData := scanner.Text()
		log.Infof("Received data: %s", rawData)
		var data ClientData
		err := json.Unmarshal([]byte(rawData), &data)
		if err != nil {
			log.Errorf("rtl_amr: Error parsing client data: %s", err)
			continue
		}

		err = r.processData(&data)
		if err != nil {
			log.Errorf("rtl_amr: Error processing data %#v: %s", data, err)
		}
	}
}

func (r client) processStderrPipe(reader io.ReadCloser) {
	errScanner := bufio.NewScanner(reader)
	for errScanner.Scan() {
		msg := errScanner.Text()
		// Some debug logs are written to stderr
		// This is filtering them out to keep the error log clean
		if !strings.Contains(msg, "decode.go") &&
			!strings.Contains(msg, "GainCount: ") {
			log.Errorf("rtl_amr: %s\n", msg)
		}
	}
}

func (r client) processData(data *ClientData) error {
	rtlAMRData, err := data.ToRTLAMRData()
	if err != nil {
		return err
	}

	lastData, found, err := r.repo.GetLastReading(rtlAMRData.MeterID)
	if err != nil {
		return err
	}
	if found {
		rtlAMRData.Difference = rtlAMRData.CurrentReading - lastData.CurrentReading
	}

	stored, err := r.repo.StoreReading(rtlAMRData)
	if err != nil {
		return err
	}

	if !stored {
		log.Infof("rtl_tcp: Did not store reading of %d for meter %s.", rtlAMRData.CurrentReading, rtlAMRData.MeterID)
	}

	return nil
}
