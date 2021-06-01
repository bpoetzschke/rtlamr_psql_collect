package rtlamrclient

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/bpoetzschke/rtlamr_psql_collect/repositories"
	"github.com/caarlos0/env/v6"
	log "github.com/sirupsen/logrus"
)

type Client interface {
	Run() error
}

type rtlAMRConfig struct {
	FilterIDs string `env:"RTLAMR_FILTERID"`
	Server    string `env:"RTLAMR_SERVER"`
}

type client struct {
	repo    repositories.RTLAMRRepo
	cmdArgs []string
}

func New(repo repositories.RTLAMRRepo) (Client, error) {
	rtl := client{repo: repo}
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

	if cfg.Server != "" {
		log.Debugf("Using rtlamr with server: %s", cfg.Server)
		r.cmdArgs = append(r.cmdArgs, fmt.Sprintf("-server=%s", cfg.Server))
	}

	return nil
}

func (r client) Run() error {
	rtlAMRCmd := exec.Command("rtlamr", r.cmdArgs...)
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

	err = rtlAMRCmd.Wait()
	if err != nil {
		return err
	}

	return nil
}

func (r client) processStdoutPipe(reader io.ReadCloser) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		rawData := scanner.Text()
		log.Debugf("Received data: %s", rawData)
		var data ClientData
		err := json.Unmarshal([]byte(rawData), &data)
		if err != nil {
			log.Errorf("Error parsing client data: %s", err)
			continue
		}

		err = r.processData(&data)
		if err != nil {
			log.Errorf("Error processing data %#v: %s", data, err)
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
			log.Errorln(msg)
		}
	}
}

func (r client) processData(data *ClientData) error {
	rtlAMRData := data.ToRTLAMRData()

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
		log.Infof("Did not store reading of %d for meter %s.", rtlAMRData.CurrentReading, rtlAMRData.MeterID)
	}

	return nil
}
