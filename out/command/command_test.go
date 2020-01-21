package command_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/alphagov/travis-resource/common"
	"github.com/alphagov/travis-resource/messager"
	"github.com/alphagov/travis-resource/model"
	. "github.com/alphagov/travis-resource/out/command"
	"github.com/alphagov/travis-resource/travis/mock_travis"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/shuheiktgw/go-travis"
)

type GinkgoTestReporter struct{}

func (g GinkgoTestReporter) Errorf(format string, args ...interface{}) {
	Fail(fmt.Sprintf(format, args...))
}

func (g GinkgoTestReporter) Fatalf(format string, args ...interface{}) {
	Fail(fmt.Sprintf(format, args...))
}

var _ = Describe("OutCommand", func() {
	var (
		t               GinkgoTestReporter
		mockCtrl        *gomock.Controller
		mockBuilds      *mock_travis.MockBuildsInterface
		mockJobs        *mock_travis.MockJobsInterface
		travisClient    *travis.Client
		outCommand      *OutCommand
		responseBuffer  bytes.Buffer
		responseWritter *bufio.Writer
		logBuffer       bytes.Buffer
		logWritter      *bufio.Writer
		tempDir         string
		outRequest      model.OutRequest
		repository      string       = "myrepo"
		jobId1          uint         = uint(1)
		jobId2          uint         = uint(2)
		buildId         uint         = 1234
		buildNumber                  = "12"
		build           travis.Build = travis.Build{
			Id:        buildId,
			State:     travis.SUCCEEDED_STATE,
			Duration:  60,
			StartedAt: "now",
			Number:    buildNumber,
			JobIds:    []uint{jobId1, jobId2},
		}
		commit travis.Commit = travis.Commit{
			Sha:         "ref",
			AuthorName:  "arthurh",
			CommittedAt: "now",
			Message:     "message",
		}
	)

	BeforeEach(func() {
		var err error
		tempDir, err = ioutil.TempDir(os.TempDir(), "travis-resource-in")
		Expect(err).To(BeNil())
		responseBuffer.Reset()
		logWritter = bufio.NewWriter(&logBuffer)
		responseWritter = bufio.NewWriter(&responseBuffer)
		mockCtrl = gomock.NewController(t)
		mockBuilds = mock_travis.NewMockBuildsInterface(mockCtrl)
		mockJobs = mock_travis.NewMockJobsInterface(mockCtrl)
		travisClient, err = common.MakeTravisClient(model.Source{})
		Expect(err).To(BeNil())
		travisClient.Builds = mockBuilds
		travisClient.Jobs = mockJobs
		outRequest = model.OutRequest{
			Source: model.Source{
				Repository: repository,
			},
			OutParams: model.OutParams{},
			Version:   model.Version{buildNumber},
		}
		outCommand = &OutCommand{travisClient, outRequest, repository, messager.NewMessager(logWritter, responseWritter)}
	})
	Describe("When loading repository", func() {
		Context("With no repository set in params", func() {
			It("Should fill the repository field from configuration", func() {
				outCommand.LoadRepository()
				Expect(outCommand.Repository).To(Equal(repository))
			})
		})
		Context("With repository set in params", func() {
			It("Should fill the repository field from configuration", func() {
				otherRepo := "myotherRepo"
				outCommand.Request.OutParams.Repository = otherRepo
				outCommand.LoadRepository()
				Expect(outCommand.Repository).To(Equal(otherRepo))
			})
		})
	})
	Describe("When getting the build parameter", func() {
		Context("With a number passed", func() {
			It("Should return the number as a string", func() {
				outCommand.Request.OutParams.Build = float64(1)
				Expect(outCommand.GetBuildParam()).To(Equal("1"))
			})
		})
		Context("With a string passed", func() {
			It("Should return the number as a string", func() {
				outCommand.Request.OutParams.Build = "mytext"
				Expect(outCommand.GetBuildParam()).To(Equal("mytext"))
			})
		})
		Context("With a empty or unknown type passed", func() {
			It("Should return an empty string", func() {
				Expect(outCommand.GetBuildParam()).To(Equal(""))

				outCommand.Request.OutParams.Build = int(1)
				Expect(outCommand.GetBuildParam()).To(Equal(""))
			})
		})
	})
	Describe("Getting build url", func() {
		Context("with a travis pro account", func() {
			It("should give the right travis url with build id", func() {
				expectedTravisUrl := common.GetTravisDashboardUrl(common.GetTravisUrl(true)) + repository + "/builds/" + strconv.Itoa(int(buildId))
				outCommand.Request.Source.Pro = true
				Expect(outCommand.GetBuildUrl(build)).To(Equal(expectedTravisUrl))
			})
		})
		Context("with a travis default account", func() {
			It("should give the right travis url with build id", func() {
				expectedTravisUrl := common.GetTravisDashboardUrl(common.GetTravisUrl(false)) + repository + "/builds/" + strconv.Itoa(int(buildId))
				outCommand.Request.Source.Pro = false
				Expect(outCommand.GetBuildUrl(build)).To(Equal(expectedTravisUrl))
			})
		})
		Context("with a custom travis api endpoint", func() {
			It("should give the right travis url with build id", func() {
				api := "https://api.mytravis.com/"
				expectedTravisUrl := common.GetTravisDashboardUrl(api) + repository + "/builds/" + strconv.Itoa(int(buildId))
				outCommand.Request.Source.Url = api
				Expect(outCommand.GetBuildUrl(build)).To(Equal(expectedTravisUrl))
			})
		})
	})
	Describe("Getting build from travis", func() {
		var callLatest *gomock.Call
		var callLatestFromRepository *gomock.Call
		var callFirstFromParamBuildNumber *gomock.Call
		var callFirstFromBuildNumber *gomock.Call
		var callFirstFromBranch *gomock.Call
		var buildNumberParams string = "123"
		var branchParams string = "mybranch2"
		var repoParam string = "myrepo2"
		BeforeEach(func() {
			callLatest = mockBuilds.EXPECT().GetFirstFinishedBuild(repository).Return(build, nil).AnyTimes()
			callLatestFromRepository = mockBuilds.EXPECT().GetFirstFinishedBuild(repoParam).Return(build, nil).AnyTimes()
			callFirstFromParamBuildNumber = mockBuilds.EXPECT().GetFirstBuildFromBuildNumber(repository, buildNumberParams).Return(build, nil).AnyTimes()
			callFirstFromBuildNumber = mockBuilds.EXPECT().GetFirstBuildFromBuildNumber(repository, buildNumber).Return(build, nil).AnyTimes()
			callFirstFromBranch = mockBuilds.EXPECT().GetFirstFinishedBuildWithBranch(repository, branchParams).Return(build, nil).AnyTimes()
		})
		Context("With params build set to latest", func() {
			It("should give the last build found", func() {
				build, err := outCommand.GetBuild("latest")
				Expect(build).To(BeEquivalentTo(build))
				Expect(err).To(BeNil())
				callLatest.Times(1)
			})
		})
		Context("With params build set to a number", func() {
			It("should give the build with this number", func() {
				build, err := outCommand.GetBuild(buildNumberParams)
				Expect(build).To(BeEquivalentTo(build))
				Expect(err).To(BeNil())
				callFirstFromParamBuildNumber.Times(1)
			})
		})
		Context("With only params repository set", func() {
			It("should give the last build found in this repository", func() {
				outCommand.Repository = repoParam
				outCommand.Request.OutParams.Repository = repoParam
				outCommand.Request.OutParams.Build = ""
				outCommand.Request.OutParams.Branch = ""
				build, err := outCommand.GetBuild("")
				Expect(build).To(BeEquivalentTo(build))
				Expect(err).To(BeNil())
				callLatestFromRepository.Times(1)
			})
		})
		Context("With params branch set", func() {
			It("should give the last finished build found related to this branch", func() {
				outCommand.Request.OutParams.Branch = branchParams
				build, err := outCommand.GetBuild("")
				Expect(build).To(BeEquivalentTo(build))
				Expect(err).To(BeNil())
				callFirstFromBranch.Times(1)
			})
		})
		Context("Without params set", func() {
			It("should give the last build found in the repository set in source", func() {
				build, err := outCommand.GetBuild("")
				Expect(build).To(BeEquivalentTo(build))
				Expect(err).To(BeNil())
				callFirstFromBuildNumber.Times(1)
			})
		})
	})
	Describe("When restarting a build", func() {
		var callBuild *gomock.Call
		var callRestart *gomock.Call
		BeforeEach(func() {
			callRestart = mockBuilds.EXPECT().Restart(buildId).AnyTimes()
			callBuild = mockBuilds.EXPECT().GetFirstBuildFromBuildNumber(repository, buildNumber).Return(build, nil).AnyTimes()
		})
		Context("Without waiting the build to finish", func() {
			It("Should return the build with new informations", func() {

				outCommand.Request.OutParams.SkipWait = true
				build, err := outCommand.Restart(build)
				Expect(build).To(BeEquivalentTo(build))
				Expect(err).To(BeNil())
				callRestart.Times(1)
				callBuild.Times(1)

			})
		})
		Context("And waiting the build to finish", func() {
			buildStarted := travis.Build{
				Id:        buildId,
				State:     travis.STATE_STARTED,
				Duration:  60,
				StartedAt: "now",
				Number:    buildNumber,
				JobIds:    []uint{jobId1, jobId2},
			}
			It("Should return the build with new informations", func() {
				outCommand.Request.OutParams.SkipWait = false
				build, err := outCommand.Restart(buildStarted)
				Expect(build).To(BeEquivalentTo(build))
				Expect(err).To(BeNil())
				callRestart.Times(1)
				callBuild.Times(2)
			})
		})
	})
	Describe("When send the version and metadata to output", func() {
		Context("With build and commit given", func() {
			It("should give correct metadata and version", func() {
				outCommand.SendResponse(build, commit)
				responseWritter.Flush()
				var reponseJson model.InResponse
				err := json.Unmarshal(responseBuffer.Bytes(), &reponseJson)
				Expect(err).To(BeNil())
				Expect(reponseJson).To(BeEquivalentTo(model.InResponse{
					Metadata: common.GetMetadatasFromBuild(build, commit),
					Version:  model.Version{build.Number},
				}))
			})
		})
	})
	AfterEach(func() {
		os.Remove(tempDir)
		mockCtrl.Finish()
	})
})
