package rtlamrclient

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"testing"
	"time"

	"github.com/bpoetzschke/rtlamr_psql_collect/repositories"
	"github.com/stretchr/testify/suite"
)

var testShouldRun = flag.Bool("GO_WANT_HELPER_PROCESS", false, "")
var fakeExecTimeout = flag.Int64("FAKE_EXEC_TIMEOUT", int64(2*time.Second), "")

type RtlAmrClientSuite struct {
	suite.Suite
	clnt     client
	repoMock repositories.RTLAMRRepoMock
}

func TestRtlAmrClient(t *testing.T) {
	suite.Run(t, &RtlAmrClientSuite{})
}

func (s *RtlAmrClientSuite) SetupTest() {
	s.repoMock = repositories.RTLAMRRepoMock{}
	s.clnt = client{
		repo:    &s.repoMock,
		done:    make(chan struct{}),
		cmdArgs: nil,
	}
}

func (s *RtlAmrClientSuite) TearDownTest() {
	s.Require().NoError(os.Setenv("RTLAMR_SERVER", ""))
	s.Require().NoError(os.Setenv("RTLAMR_FILTERID", ""))

	execCmd = exec.Command

	s.repoMock.AssertExpectations(s.T())
}

func (s *RtlAmrClientSuite) xTestClientSetupParsesEnvVarsDefault() {
	err := s.clnt.setup()
	s.Require().NoError(err)
	s.Require().EqualValues([]string{"-format=json", "-unique=true"}, s.clnt.cmdArgs)
}

func (s *RtlAmrClientSuite) xTestClientSetupParsesEnvVarForServer() {
	s.Require().NoError(os.Setenv("RTLAMR_SERVER", "RTLAMR_SERVER_VALUE"))

	err := s.clnt.setup()
	s.Require().NoError(err)
	s.Require().EqualValues([]string{"-format=json", "-unique=true", "-server=RTLAMR_SERVER_VALUE"}, s.clnt.cmdArgs)
}

func (s *RtlAmrClientSuite) xTestClientSetupParsesEnvVarForFilterIDs() {
	s.Require().NoError(os.Setenv("RTLAMR_FILTERID", "RTLAMR_FILTERID_VALUE"))

	err := s.clnt.setup()
	s.Require().NoError(err)
	s.Require().EqualValues([]string{"-format=json", "-unique=true", "-filterid=RTLAMR_FILTERID_VALUE"}, s.clnt.cmdArgs)
}

func (s *RtlAmrClientSuite) xTestClientRun() {
	execCmd = s.generateFakeExecCmd(2*time.Second, "rtlamr", "-format=json", "-unique=true")

	ctx := context.Background()

	err := s.clnt.setup()
	s.Require().NoError(err)

	err = s.clnt.Run(ctx)
	s.Require().NoError(err)

	channelClosed := false
	select {
	case <-s.clnt.done:
		channelClosed = true
	default:
	}

	s.Require().True(channelClosed)
}

func (s *RtlAmrClientSuite) TestClientRunCtxCanceled() {
	execCmd = s.generateFakeExecCmd(2*time.Minute, "rtlamr", "-format=json", "-unique=true")

	ctx := context.Background()
	cancelCtx, cancelFunc := context.WithCancel(ctx)

	err := s.clnt.setup()
	s.Require().NoError(err)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		err = s.clnt.Run(cancelCtx)
		s.Require().NoError(err)
		wg.Done()
	}()

	<-time.After(2 * time.Second)

	cancelFunc()

	channelClosed := false
	select {
	case <-s.clnt.done:
		channelClosed = true
	default:
	}

	s.Require().False(channelClosed)
}

// generateFakeExecCmd method generates a valid implementation of the exec.Command method. generateFakeExecCmd accepts
// a timeout parameter which is used inside the generated method as flag. When the generated method
// is called as part of the current test suite we start the current go test binary again but this time we only provide
// one test which should be run. This test is also part of this file and is only run when the GO_WANT_HELPER_PROCESS
// flag is set.
func (s *RtlAmrClientSuite) generateFakeExecCmd(timeout time.Duration, expectedCmd string, expectedArgs ...string) func(string, ...string) *exec.Cmd {
	return func(command string, args ...string) *exec.Cmd {
		s.Require().EqualValues(expectedCmd, command)
		s.Require().EqualValues(expectedArgs, args)
		cs := []string{
			"-test.run=TestRtlAmrClientHelperProcess",
			"-GO_WANT_HELPER_PROCESS=true",
			fmt.Sprintf("-FAKE_EXEC_TIMEOUT=%d", timeout),
			"--",
			command,
		}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		return cmd
	}
}

func TestRtlAmrClientHelperProcess(t *testing.T) {
	if testShouldRun != nil && !*testShouldRun {
		return
	}

	timeout := 2 * time.Second
	if fakeExecTimeout != nil {
		timeout = time.Duration(*fakeExecTimeout)
	}

	<-time.After(timeout)

	os.Exit(0)
}
