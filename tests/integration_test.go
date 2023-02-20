//go:build integration

package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	dc "github.com/ory/dockertest/v3/docker"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"

	"github.com/wimspaargaren/go-lambda-localstack-example/internal/api"
)

const (
	localStackPort = "4566"
	host           = "localhost"
)

type LocalstackIntegrationSuite struct {
	suite.Suite

	pool       *dockertest.Pool
	localStack *dockertest.Resource

	startTime time.Time

	lambdaBaseURL string
	httpClient    *http.Client

	logger *log.Logger
}

func (s *LocalstackIntegrationSuite) SetupSuite() {
	logger := log.New()
	logger.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "15:04:05",
	})
	s.logger = logger
	s.httpClient = http.DefaultClient

	s.startTime = time.Now()
	pool, err := dockertest.NewPool("")
	s.NoError(err)

	s.pool = pool

	// Start localstack container & apply terraform
	s.startLocalStack()
	s.applyTerraform()

	s.logger.Infof("initialised dependencies in %s", time.Since(s.startTime))
}

func (s *LocalstackIntegrationSuite) startLocalStack() {
	res := map[dc.Port][]dc.PortBinding{
		dc.Port(fmt.Sprintf("%s/tcp", localStackPort)): {{HostIP: "", HostPort: localStackPort}},
	}
	// Future self; In case you find a better way to do this, please improve.
	for i := 4510; i <= 4559; i++ {
		res[dc.Port(fmt.Sprintf("%d/tcp", i))] = []dc.PortBinding{{HostIP: "", HostPort: fmt.Sprintf("%d", i)}}
	}

	localStack, err := s.pool.RunWithOptions(&dockertest.RunOptions{
		Repository:   "localstack/localstack",
		Tag:          "latest",
		PortBindings: res,
	})
	s.NoError(err)
	s.localStack = localStack

	err = s.pool.Retry(func() error {
		req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, fmt.Sprintf("http://%s:%s/_localstack/health", host, localStackPort), nil)
		s.NoError(err)
		resp, err := s.httpClient.Do(req)
		if err != nil {
			s.logger.WithError(err).Infof("unable to perform http request to localstack")
			return err
		}
		if resp.StatusCode != http.StatusOK {
			s.logger.WithError(err).Infof("localstack not healthy, try again...")
			return fmt.Errorf("local stack not healthy")
		}
		return nil
	})
	s.NoError(err)
	s.logger.Info("localStack is healthy")
	go func() {
		ctx := context.Background()
		opts := dc.LogsOptions{
			Context: ctx,

			Stderr:      true,
			Stdout:      true,
			Follow:      true,
			Timestamps:  true,
			RawTerminal: true,

			Container: localStack.Container.ID,

			OutputStream: s.logger.Writer(),
		}
		err := s.pool.Client.Logs(opts)
		s.NoError(err)
	}()
}

func (s *LocalstackIntegrationSuite) applyTerraform() {
	s.logger.Info("setting up Lambda using Terraform...")
	err := exec.Command("tflocal", "-chdir=../infra/terraform", "init").Run()
	s.NoError(err)
	err = exec.Command("tflocal", "-chdir=../infra/terraform", "apply", "--auto-approve").Run()
	s.NoError(err)
	output, err := exec.Command("tflocal", "-chdir=../infra/terraform", "output", "-raw", "api_gw_id").Output()
	s.NoError(err)

	s.lambdaBaseURL = fmt.Sprintf("http://%s:%s/restapis/%s/test/_user_request_", host, localStackPort, strings.TrimSpace(string(output)))
	s.logger.Infof("base url: %s", s.lambdaBaseURL)
}

type Response struct {
	Message string `json:"message"`
}

func (s *LocalstackIntegrationSuite) TestHelloWorld() {
	url := fmt.Sprintf("%s/hello-world", s.lambdaBaseURL)
	s.logger.Infof("making request to: %s", url)
	req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, fmt.Sprintf("%s/hello-world", s.lambdaBaseURL), nil)
	s.NoError(err)
	resp, err := s.httpClient.Do(req)
	s.NoError(err)

	s.logger.Infof("response status code: %d", resp.StatusCode)
	s.Equal(http.StatusOK, resp.StatusCode)

	b, err := io.ReadAll(resp.Body)
	s.NoError(err)

	response := Response{}
	err = json.Unmarshal(b, &response)
	s.NoError(err)

	s.logger.Infof("response body: %s", string(b))
	s.Equal("Hello World!", response.Message)
}

func (s *LocalstackIntegrationSuite) TestYourName() {
	tests := []struct {
		Name               string
		Body               api.YourNameRequest
		ExpectedStatusCode int
		ExpectedMsg        string
	}{
		{
			Name: "Happy flow",
			Body: api.YourNameRequest{
				Name: "Wim",
			},
			ExpectedStatusCode: http.StatusOK,
			ExpectedMsg:        "your name is: Wim", // useful in case I forget!
		},
		{
			Name:               "No name provided",
			Body:               api.YourNameRequest{},
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedMsg:        "if you don't tell me I don't know your name",
		},
	}
	for _, test := range tests {
		s.Run(test.Name, func() {
			reqBytes, err := json.Marshal(test.Body)
			s.NoError(err)

			url := fmt.Sprintf("%s/your-name", s.lambdaBaseURL)
			s.logger.Infof("making request to: %s", url)

			req, err := http.NewRequestWithContext(context.TODO(), http.MethodPost, url, bytes.NewReader(reqBytes))
			s.NoError(err)
			resp, err := s.httpClient.Do(req)
			s.NoError(err)

			s.logger.Infof("response status code: %d", resp.StatusCode)
			s.Equal(test.ExpectedStatusCode, resp.StatusCode)

			b, err := io.ReadAll(resp.Body)
			s.NoError(err)

			response := Response{}
			err = json.Unmarshal(b, &response)
			s.NoError(err)

			s.logger.Infof("response body: %s", string(b))
			s.Equal(test.ExpectedMsg, response.Message)
		})
	}
}

func (s *LocalstackIntegrationSuite) TearDownSuite() {
	s.NoError(s.localStack.Close())
	s.logger.Infof("localstack integration test completed in: %s", time.Since(s.startTime))
}

func TestLocalstackIntegrationSuite(t *testing.T) {
	suite.Run(t, new(LocalstackIntegrationSuite))
}
