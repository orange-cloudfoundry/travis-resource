package command_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/golang/mock/gomock"
	"fmt"
	"github.com/Orange-OpenSource/travis-resource/travis/mock_travis"
	"github.com/Orange-OpenSource/travis-resource/travis"
	"github.com/Orange-OpenSource/travis-resource/common"
	"github.com/Orange-OpenSource/travis-resource/model"
	. "github.com/Orange-OpenSource/travis-resource/in/command"
	"github.com/Orange-OpenSource/travis-resource/messager"
	"bytes"
	"bufio"
	"io/ioutil"
	"os"
	"path/filepath"
	"encoding/json"
	"net/http"
)

type GinkgoTestReporter struct{}

func (g GinkgoTestReporter) Errorf(format string, args ...interface{}) {
	Fail(fmt.Sprintf(format, args...))
}

func (g GinkgoTestReporter) Fatalf(format string, args ...interface{}) {
	Fail(fmt.Sprintf(format, args...))
}

var _ = Describe("InCommand", func() {
	var (
		t GinkgoTestReporter
		mockCtrl *gomock.Controller
		mockBuilds *mock_travis.MockBuildsInterface
		mockJobs *mock_travis.MockJobsInterface
		travisClient *travis.Client
		inCommand *InCommand
		responseBuffer bytes.Buffer
		responseWritter *bufio.Writer
		logBuffer bytes.Buffer
		logWritter *bufio.Writer
		tempDir string
		inRequest model.InRequest
		repository string = "myrepo"
		jobId1 uint = uint(1)
		jobId2 uint = uint(2)
		build travis.Build = travis.Build{
			State: travis.SUCCEEDED_STATE,
			Duration: uint(60),
			StartedAt: "now",
			JobIds: []uint{jobId1, jobId2},
		}
		commit travis.Commit = travis.Commit{
			Sha: "ref",
			AuthorName: "arthurh",
			CommittedAt: "now",
			Message: "message",
		}
		job travis.Job = travis.Job{
			Id: jobId1,
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
		inRequest = model.InRequest{
			Source: model.Source{
				Repository: repository,
			},
		}
		inCommand = &InCommand{travisClient, inRequest, tempDir, messager.NewMessager(logWritter, responseWritter)}
	})
	Describe("When getting build information", func() {
		Context("with build found", func() {
			var call *gomock.Call
			BeforeEach(func() {
				call = mockBuilds.EXPECT().ListFromRepository(repository, gomock.Any()).Return(
					[]travis.Build{build},
					[]travis.Job{job},
					[]travis.Commit{commit},
					nil,
					nil,
				).AnyTimes()
			})
			It("should give the build and the complete information about build", func() {
				buildFound, listBuilds, err := inCommand.GetBuildInfo()
				Expect(err).To(BeNil())
				Expect(buildFound).To(BeEquivalentTo(build))
				Expect(listBuilds).To(BeEquivalentTo(travis.ListBuildsResponse{
					Builds: []travis.Build{build},
					Jobs: []travis.Job{job},
					Commits: []travis.Commit{commit},
				}))
			})
		})
		Context("with build not found", func() {
			var call *gomock.Call
			BeforeEach(func() {
				call = mockBuilds.EXPECT().ListFromRepository(repository, gomock.Any()).Return(
					[]travis.Build{},
					[]travis.Job{job},
					[]travis.Commit{commit},
					nil,
					nil,
				).AnyTimes()
			})
			It("should raise an error", func() {
				_, _, err := inCommand.GetBuildInfo()
				Expect(err).NotTo(BeNil())
			})
		})
	})
	Describe("write build informations in designated file", func() {
		It("should have a file created and have good informations inside", func() {
			listBuilds := travis.ListBuildsResponse{
				Builds: []travis.Build{build},
				Jobs: []travis.Job{job},
				Commits: []travis.Commit{commit},
			}
			inCommand.WriteInBuildInfoFile(listBuilds)
			expectedFile := filepath.Join(tempDir, common.FILENAME_BUILD_INFO)
			Expect(expectedFile).Should(BeAnExistingFile())
			var listBuildsFound travis.ListBuildsResponse
			bytes, err := ioutil.ReadFile(expectedFile)
			Expect(err).To(BeNil())
			err = json.Unmarshal(bytes, &listBuildsFound)
			Expect(err).To(BeNil())
			Expect(listBuildsFound).To(BeEquivalentTo(listBuilds))
		})
	})
	Describe("download log in logs folder", func() {
		jobLogString1 := "log 1"
		jobLogString2 := "log 2"
		jobLogBytes1 := []byte(jobLogString1)
		jobLogBytes2 := []byte(jobLogString2)

		jobResponse1 := &http.Response{
			ContentLength: int64(len(jobLogBytes1)),
			Body:  ioutil.NopCloser(bytes.NewReader(jobLogBytes1)),
		}
		jobResponse2 := &http.Response{
			ContentLength: int64(len(jobLogBytes2)),
			Body:  ioutil.NopCloser(bytes.NewReader(jobLogBytes2)),
		}

		BeforeEach(func() {
			inCommand.Request.InParams.DownloadLogs = true
			mockJobs.EXPECT().RawLogOnlyResponse(jobId1).Return(jobResponse1, nil).AnyTimes()
			mockJobs.EXPECT().RawLogOnlyResponse(jobId2).Return(jobResponse2, nil).AnyTimes()
		})
		Context("with download logs set to true", func() {
			It("should have logs in the folder with right data inside", func() {
				err := inCommand.DownloadLogs(build)
				Expect(err).To(BeNil())

				fileJob1 := filepath.Join(tempDir, LOGS_FOLDER, fmt.Sprintf(LOGS_FILENAME_PATTERN, jobId1))
				fileJob2 := filepath.Join(tempDir, LOGS_FOLDER, fmt.Sprintf(LOGS_FILENAME_PATTERN, jobId2))
				Expect(fileJob1).Should(BeAnExistingFile())
				Expect(fileJob2).Should(BeAnExistingFile())

				bytes, err := ioutil.ReadFile(fileJob1)
				Expect(err).To(BeNil())
				Expect(string(bytes)).To(Equal(jobLogString1))

				bytes, err = ioutil.ReadFile(fileJob2)
				Expect(err).To(BeNil())
				Expect(string(bytes)).To(Equal(jobLogString2))
			})
		})

	})
	AfterEach(func() {
		d, err := os.Open(tempDir)
		if err != nil {
			panic(err)
		}
		defer d.Close()
		names, err := d.Readdirnames(-1)
		if err != nil {
			panic(err)
		}
		for _, name := range names {
			err = os.RemoveAll(filepath.Join(tempDir, name))
			if err != nil {
				panic(err)
			}
		}
		os.Remove(tempDir)
		mockCtrl.Finish()
	})
})
