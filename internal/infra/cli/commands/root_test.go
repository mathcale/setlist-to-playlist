package commands

// import (
// 	"errors"
// 	"testing"
// 	"time"

// 	"github.com/stretchr/testify/mock"
// 	"github.com/stretchr/testify/suite"

// 	"github.com/mathcale/goexpert-stresstest-cli-challenge/internal/pkg/httpclient"
// 	"github.com/mathcale/goexpert-stresstest-cli-challenge/internal/tests/mocks"
// 	"github.com/mathcale/goexpert-stresstest-cli-challenge/internal/tests/utils"
// 	report_dto "github.com/mathcale/goexpert-stresstest-cli-challenge/internal/usecases/report/dto"
// 	stress_dto "github.com/mathcale/goexpert-stresstest-cli-challenge/internal/usecases/stress/dto"
// )

// type StressTestCmdTestSuite struct {
// 	suite.Suite
// 	StressTestUseCaseMock *mocks.StressTestUseCaseMock
// 	ReportUseCaseMock     *mocks.ReportUseCaseMock
// 	StressTestCmd         StressTestCmdInterface
// }

// func (s *StressTestCmdTestSuite) SetupTest() {
// 	s.StressTestUseCaseMock = new(mocks.StressTestUseCaseMock)
// 	s.ReportUseCaseMock = new(mocks.ReportUseCaseMock)

// 	s.StressTestCmd = NewStressTestCmd(s.StressTestUseCaseMock, s.ReportUseCaseMock)
// }

// func (s *StressTestCmdTestSuite) cleanMocks() {
// 	s.StressTestUseCaseMock.ExpectedCalls = nil
// 	s.ReportUseCaseMock.ExpectedCalls = nil
// }

// func TestStressTestCmd(t *testing.T) {
// 	suite.Run(t, new(StressTestCmdTestSuite))
// }

// func (s *StressTestCmdTestSuite) TestBuild() {
// 	s.Run("Should build a new stress test command", func() {
// 		cmd := s.StressTestCmd.Build()

// 		s.NotNil(cmd)
// 	})

// 	s.Run("Should have the correct short description", func() {
// 		cmd := s.StressTestCmd.Build()

// 		s.Equal("Stress test a given URL", cmd.Short)
// 	})

// 	s.Run("Should have the correct long description", func() {
// 		cmd := s.StressTestCmd.Build()

// 		s.Equal("Executes a stress test on a given URL with a given number of requests and concurrency.", cmd.Long)
// 	})

// 	s.Run("Should have the correct run function", func() {
// 		cmd := s.StressTestCmd.Build()

// 		s.NotNil(cmd.RunE)
// 	})

// 	s.Run("Should have the correct flags", func() {
// 		cmd := s.StressTestCmd.Build()

// 		flags := cmd.Flags()

// 		s.NotNil(flags.Lookup("url"))
// 		s.NotNil(flags.Lookup("requests"))
// 		s.NotNil(flags.Lookup("concurrency"))
// 	})

// 	s.Run("Should run command with success", func() {
// 		defer s.cleanMocks()

// 		s.StressTestUseCaseMock.On("Execute", mock.Anything).Return(&stress_dto.StressTestOutput{
// 			Duration: time.Duration(500),
// 			Results: []*httpclient.HttpClientResponse{
// 				{
// 					StatusCode: utils.IntPtr(200),
// 					Duration:   time.Duration(100),
// 					Error:      nil,
// 				},
// 				{
// 					StatusCode: utils.IntPtr(200),
// 					Duration:   time.Duration(200),
// 					Error:      nil,
// 				},
// 				{
// 					StatusCode: utils.IntPtr(200),
// 					Duration:   time.Duration(300),
// 					Error:      nil,
// 				},
// 			},
// 		}, nil)

// 		s.ReportUseCaseMock.On("Execute", mock.Anything).Return(&report_dto.ReportOutput{
// 			Duration:       time.Duration(500),
// 			SuccessfulReqs: 3,
// 			FailedReqs:     0,
// 			StatusCount: map[int]uint64{
// 				200: 3,
// 			},
// 			LatencyPercentiles: map[int]time.Duration{
// 				50: time.Duration(200),
// 				75: time.Duration(300),
// 				90: time.Duration(300),
// 				95: time.Duration(300),
// 				99: time.Duration(300),
// 			},
// 		})

// 		cmd := s.StressTestCmd.Build()
// 		err := cmd.RunE(cmd, []string{
// 			"--url", "http://example.com",
// 			"--requests", "3",
// 			"--concurrency", "1",
// 		})

// 		s.NoError(err)
// 	})

// 	s.Run("Should return an error when something goes wrong on stress test use-case", func() {
// 		defer s.cleanMocks()

// 		s.StressTestUseCaseMock.On("Execute", mock.Anything).Return(&stress_dto.StressTestOutput{}, errors.New("any-error"))
// 		s.ReportUseCaseMock.On("Execute", mock.Anything).Return(&report_dto.ReportOutput{})

// 		cmd := s.StressTestCmd.Build()
// 		err := cmd.RunE(cmd, []string{})

// 		s.Error(err)
// 	})
// }
