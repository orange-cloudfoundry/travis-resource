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
	repository := request.Source.Repository
	if request.OutParams.Repository != "" {
		repository = request.OutParams.Repository
	}
	outCommand := &OutCommand{travisClient, request, repository}
	buildParam := ""
	if buildParamNumber, ok := request.OutParams.Build.(float64); ok {
		buildParam = strconv.FormatFloat(buildParamNumber, 'f', 0, 64)
	}
	if buildParamString, ok := request.OutParams.Build.(string); ok {
		buildParam = buildParamString
	}

	build, err = outCommand.getBuild(buildParam)
	common.FatalIf("fetch build error", err)

	travisClient.Builds.Restart(build.Id)
	if !request.OutParams.SkipWait {
		outCommand.waitBuild(build.Number)
	}
	build, err = travisClient.Builds.GetFirstBuildFromBuildNumber(repository, build.Number)
	common.FatalIf("can't get build after restart", err)
	response := model.InResponse{common.GetMetadatasFromBuild(build), model.Version{build.Number}}
	json.NewEncoder(os.Stdout).Encode(response)
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
		if !stringInSlice(build.State, travis.RUNNING_STATE) {
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
func stringInSlice(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}