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
	. "github.com/Orange-OpenSource/travis-resource/check/command"
	"github.com/Orange-OpenSource/travis-resource/messager"
	"bytes"
	"bufio"
	"encoding/json"
)

type GinkgoTestReporter struct{}

func (g GinkgoTestReporter) Errorf(format string, args ...interface{}) {
	Fail(fmt.Sprintf(format, args...))
}

func (g GinkgoTestReporter) Fatalf(format string, args ...interface{}) {
	Fail(fmt.Sprintf(format, args...))
}

var _ = Describe("CheckCommand", func() {
	var (
		t GinkgoTestReporter
		mockCtrl *gomock.Controller
		mockBuilds *mock_travis.MockBuildsInterface
		travisClient *travis.Client
		checkCommand *CheckCommand
		responseBuffer bytes.Buffer
		responseWritter *bufio.Writer
		logBuffer bytes.Buffer
		logWritter *bufio.Writer
	)

	BeforeEach(func() {

		responseBuffer.Reset()
		logWritter = bufio.NewWriter(&logBuffer)
		responseWritter = bufio.NewWriter(&responseBuffer)
		mockCtrl = gomock.NewController(t)
		mockBuilds = mock_travis.NewMockBuildsInterface(mockCtrl)
		var err error
		travisClient, err = common.MakeTravisClient(model.Source{})
		Expect(err).To(BeNil())
		travisClient.Builds = mockBuilds
		checkCommand = &CheckCommand{travisClient, model.CheckRequest{}, messager.NewMessager(logWritter, responseWritter)}
	})
	Describe("When check send the version to output", func() {
		Context("With no build number given", func() {
			It("should give an empty version list", func() {
				checkCommand.SendResponse("")
				responseWritter.Flush()
				var reponseJson model.CheckResponse
				err := json.Unmarshal(responseBuffer.Bytes(), &reponseJson)
				Expect(err).To(BeNil())
				Expect(reponseJson).To(BeEquivalentTo(model.CheckResponse{}))
			})
		})
		Context("With build number given", func() {
			It("should give an empty version list", func() {
				checkCommand.SendResponse("1")
				responseWritter.Flush()
				var reponseJson model.CheckResponse
				err := json.Unmarshal(responseBuffer.Bytes(), &reponseJson)
				Expect(err).To(BeNil())
				Expect(reponseJson).To(BeEquivalentTo(model.CheckResponse{model.Version{"1"}}))
			})
		})
	})
	Describe("When user want to have the last build number", func() {
		Context("without branch set and check only finished builds", func() {
			var call *gomock.Call
			BeforeEach(func() {
				checkCommand.Request.Source = model.Source{
					Repository: "myrepo",
					CheckAllBuilds: false,
				}
				call = mockBuilds.EXPECT().ListFromRepositoryWithInfos(gomock.Any(), "", travis.STATE_PASSED, gomock.Nil()).AnyTimes()
			})
			It("expect to run the right travis call", func() {
				checkCommand.GetBuildNumber()
				call.Times(1)
			})
		})
		Context("without branch set and check only started builds", func() {
			var call *gomock.Call
			BeforeEach(func() {
				checkCommand.Request.Source = model.Source{
					Repository: "myrepo",
					CheckAllBuilds: false,
					CheckOnState: travis.STATE_STARTED,
				}
				call = mockBuilds.EXPECT().ListFromRepositoryWithInfos(gomock.Any(), "", travis.STATE_STARTED, gomock.Nil()).AnyTimes()
			})
			It("expect to run the right travis call", func() {
				checkCommand.GetBuildNumber()
				call.Times(1)
			})
		})
		Context("without branch set and check all builds", func() {
			var call *gomock.Call
			BeforeEach(func() {
				checkCommand.Request.Source = model.Source{
					Repository: "myrepo",
					CheckAllBuilds: true,
				}
				call = mockBuilds.EXPECT().ListFromRepositoryWithInfos(gomock.Any(), "", "", gomock.Nil()).AnyTimes()
			})
			It("expect to run the right travis call", func() {
				checkCommand.GetBuildNumber()
				call.Times(1)
			})
		})
		Context("with branch set and check only finished builds", func() {
			var call *gomock.Call
			BeforeEach(func() {
				checkCommand.Request.Source = model.Source{
					Repository: "myrepo",
					CheckAllBuilds: false,
					Branch: "mybranch",
				}
				call = mockBuilds.EXPECT().ListFromRepositoryWithInfos(gomock.Any(), "mybranch", travis.STATE_PASSED, gomock.Nil()).AnyTimes()
			})
			It("expect to run the right travis call", func() {
				checkCommand.GetBuildNumber()
				call.Times(1)
			})
		})
		Context("with branch set and check only started builds", func() {
			var call *gomock.Call
			BeforeEach(func() {
				checkCommand.Request.Source = model.Source{
					Repository: "myrepo",
					CheckAllBuilds: false,
					Branch: "mybranch",
					CheckOnState: travis.STATE_STARTED,
				}
				call = mockBuilds.EXPECT().ListFromRepositoryWithInfos(gomock.Any(), "mybranch", travis.STATE_STARTED, gomock.Nil()).AnyTimes()
			})
			It("expect to run the right travis call", func() {
				checkCommand.GetBuildNumber()
				call.Times(1)
			})
		})
		Context("with branch set and check all builds", func() {
			var call *gomock.Call
			BeforeEach(func() {
				checkCommand.Request.Source = model.Source{
					Repository: "myrepo",
					CheckAllBuilds: true,
					Branch: "mybranch",
				}
				call = mockBuilds.EXPECT().ListFromRepositoryWithInfos(gomock.Any(), "mybranch", "", gomock.Nil()).AnyTimes()
			})
			It("expect to run the right travis call", func() {
				checkCommand.GetBuildNumber()
				call.Times(1)
			})
		})
	})
	AfterEach(func() {
		mockCtrl.Finish()
	})
})
