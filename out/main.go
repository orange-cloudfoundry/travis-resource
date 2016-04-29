package main

import (
	"os"
	"github.com/ArthurHlt/travis-resource/common"
	"github.com/ArthurHlt/travis-resource/model"
	"encoding/json"
	"errors"
	"github.com/ArthurHlt/travis-resource/travis"
	"strconv"
	"time"
)

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
	buildParam := ""
	if buildParamNumber, ok := request.OutParams.Build.(float64); ok {
		buildParam = strconv.FormatFloat(buildParamNumber, 'f', 0, 64)
	}
	if buildParamString, ok := request.OutParams.Build.(string); ok {
		buildParam = buildParamString
	}
	if buildParam == "latest" || (request.OutParams.Repository != "" && request.OutParams.Build == "" && request.OutParams.Branch == "") {
		build, err = travisClient.Builds.GetFirstFinishedBuild(repository)
		common.FatalIf("can't get build", err)
	} else if buildParam != "" {
		build, err = travisClient.Builds.GetFirstBuildFromBuildNumber(repository, buildParam)
		common.FatalIf("can't get build", err)
	} else if request.OutParams.Branch != "" {
		build, err = travisClient.Builds.GetFirstFinishedBuildWithBranch(repository, request.OutParams.Branch)
		common.FatalIf("can't get build", err)
	} else {
		build, err = travisClient.Builds.GetFirstBuildFromBuildNumber(repository, request.Version.BuildNumber)
		common.FatalIf("can't get build", err)
	}

	travisClient.Builds.Restart(build.Id)
	if !request.OutParams.UnWaitBuild {
		waitBuild(travisClient, repository, build.Number, request)
	}
	build, err = travisClient.Builds.GetFirstBuildFromBuildNumber(repository, build.Number)
	common.FatalIf("can't get build after restart", err)
	response := model.InResponse{common.GetMetadatasFromBuild(build), model.Version{build.Number}}
	json.NewEncoder(os.Stdout).Encode(response)
}
func waitBuild(travisClient *travis.Client, repository, buildNumber string, request model.OutRequest) {
	var build travis.Build
	var err error
	var travisUrl string
	if request.Source.Url != "" {
		travisUrl = request.Source.Url
	} else {
		travisUrl = common.GetTravisUrl(request.Source.Pro)
	}
	travisUrl = common.GetTravisDashboardUrl(travisUrl)
	for {
		build, err = travisClient.Builds.GetFirstBuildFromBuildNumber(repository, buildNumber)
		common.FatalIf("can't get build after restart", err)
		if !stringInSlice(build.State, travis.RUNNING_STATE) {
			break
		}
		time.Sleep(5 * time.Second)
	}
	if build.State != travis.SUCCEEDED_STATE {
		common.FatalIf("Build '" + build.Number + "' in errored, see",
			errors.New(travisUrl + repository + "/builds/" + strconv.Itoa(int(build.Id))))
	}
}
func stringInSlice(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}