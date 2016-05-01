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
)

type GinkgoTestReporter struct{}

func (g GinkgoTestReporter) Errorf(format string, args ...interface{}) {
	Fail(fmt.Sprintf(format, args))
}

func (g GinkgoTestReporter) Fatalf(format string, args ...interface{}) {
	Fail(fmt.Sprintf(format, args))
}

var _ = Describe("Check", func() {
	var (
		t GinkgoTestReporter
		mockCtrl *gomock.Controller
		mockBuilds *mock_travis.MockBuildsInterface
		travisClient *travis.Client
		checkCommand *CheckCommand
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(t)
		mockBuilds = mock_travis.NewMockBuildsInterface(mockCtrl)
		var err error
		travisClient, err = common.MakeTravisClient(model.Source{})
		Expect(err).To(BeNil())
		travisClient.Builds = mockBuilds
		checkCommand = &CheckCommand{travisClient, model.CheckRequest{}, messager.GetMessager()}
	})
	Describe("When user want to have the last build number", func() {

		Context("without branch set and check only finished builds", func() {
			var call *gomock.Call
			BeforeEach(func() {
				checkCommand.Request.Source = model.Source{
					Repository: "myrepo",
					CheckAllBuilds: false,
				}
				call = mockBuilds.EXPECT().ListSucceededFromRepository(gomock.Any(), gomock.Nil()).AnyTimes()
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
				call = mockBuilds.EXPECT().ListFromRepository(gomock.Any(), gomock.Nil()).AnyTimes()
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
