package main

import (
	"os"
	"github.com/Orange-OpenSource/travis-resource/common"
	"github.com/Orange-OpenSource/travis-resource/model"
	"encoding/json"
	"errors"
	"github.com/Orange-OpenSource/travis-resource/travis"
	"strconv"
	"time"
	"io"
)

type OutCommand struct {
	travisClient *travis.Client
	request      model.OutRequest
	repository   string
}

func main() {

	var request model.OutRequest
	err := json.NewDecoder(os.Stdin).Decode(&request)
	common.FatalIf("failed to read request", err)
	if request.Source.Repository == "" {
		common.FatalIf("can't get build", errors.New("there is no repository set"))
	}
	travisClient, err := common.MakeTravisClient(request.Source)
	common.FatalIf("failed to create travis client", err)

	var build travis.Build
	outCommand := &OutCommand{travisClient, request, ""}
	outCommand.loadRepository()

	buildParam := outCommand.getBuildParam()

	build, err = outCommand.getBuild(buildParam)
	common.FatalIf("fetch build error", err)

	build, err = outCommand.restart(build)
	common.FatalIf("can't get build after restart", err)

	outCommand.sendResponse(build, os.Stdout)
}
func (this *OutCommand) loadRepository() {
	this.repository = this.request.Source.Repository
	if this.request.OutParams.Repository != "" {
		this.repository = this.request.OutParams.Repository
	}
}
func (this *OutCommand) getBuildParam() string {
	buildParam := ""
	if buildParamNumber, ok := this.request.OutParams.Build.(float64); ok {
		buildParam = strconv.FormatFloat(buildParamNumber, 'f', 0, 64)
	}
	if buildParamString, ok := this.request.OutParams.Build.(string); ok {
		buildParam = buildParamString
	}
	return buildParam
}
func (this *OutCommand) sendResponse(build travis.Build, w io.Writer) {
	response := model.InResponse{
		Metadata: common.GetMetadatasFromBuild(build),
		Version: model.Version{this.request.Version.BuildNumber},
	}
	json.NewEncoder(w).Encode(response)
}
func (this *OutCommand) restart(build travis.Build) (travis.Build, error) {
	this.travisClient.Builds.Restart(build.Id)
	if !this.request.OutParams.SkipWait {
		this.waitBuild(build.Number)
	}
	return this.travisClient.Builds.GetFirstBuildFromBuildNumber(this.repository, build.Number)
}
func (this *OutCommand) waitBuild(buildNumber string) {
	var build travis.Build
	var err error
	var travisUrl string
	if this.request.Source.Url != "" {
		travisUrl = this.request.Source.Url
	} else {
		travisUrl = common.GetTravisUrl(this.request.Source.Pro)
	}
	travisUrl = common.GetTravisDashboardUrl(travisUrl)
	for {
		build, err = this.travisClient.Builds.GetFirstBuildFromBuildNumber(this.repository, buildNumber)
		common.FatalIf("can't get build after restart", err)
		if !common.StringInSlice(build.State, travis.RUNNING_STATE) {
			break
		}
		time.Sleep(5 * time.Second)
	}
	if build.State != travis.SUCCEEDED_STATE {
		common.FatalIf("Build '" + build.Number + "' failed",
			errors.New("\n\tstate: " + build.State + "\n\tsee: " + travisUrl + this.repository + "/builds/" + strconv.Itoa(int(build.Id))))
	}
}
func (this *OutCommand) getBuild(buildParam string) (travis.Build, error) {
	var build travis.Build
	var err error
	if buildParam == "latest" || (this.request.OutParams.Repository != "" && this.request.OutParams.Build == "" && this.request.OutParams.Branch == "") {
		build, err = this.travisClient.Builds.GetFirstFinishedBuild(this.repository)
		if err != nil {
			return build, errors.New("can't get build for repository " + this.repository + " with latest build " + err.Error())
		}
		return build, nil
	} else if buildParam != "" {
		build, err = this.travisClient.Builds.GetFirstBuildFromBuildNumber(this.repository, buildParam)
		if err != nil {
			return build, errors.New("can't get build for repository " + this.repository + " with build " + buildParam + " " + err.Error())
		}
		return build, nil
	} else if this.request.OutParams.Branch != "" {
		build, err = this.travisClient.Builds.GetFirstFinishedBuildWithBranch(this.repository, this.request.OutParams.Branch)
		if err != nil {
			return build, errors.New("can't get build for repository " + this.repository + " with branch " + this.request.OutParams.Branch + " " + err.Error())
		}
		return build, nil
	}
	build, err = this.travisClient.Builds.GetFirstBuildFromBuildNumber(this.repository, this.request.Version.BuildNumber)
	if err != nil {
		return build, errors.New("can't get build for repository " + this.repository + " with build " + this.request.Version.BuildNumber + " " + err.Error())
	}
	return build, nil

}
