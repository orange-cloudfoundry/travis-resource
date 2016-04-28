package main

import (
	"os"
	"github.com/ArthurHlt/travis-resource/common"
	"github.com/ArthurHlt/travis-resource/model"
	"encoding/json"
	"errors"
	"github.com/ArthurHlt/travis-resource/travis"
	"strconv"
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
		buildParam = strconv.FormatFloat(buildParamNumber, 'f', 6, 64)
	}
	if buildParamString, ok := request.OutParams.Build.(string); ok {
		buildParam = buildParamString
	}
	common.FatalIf("build param", errors.New(buildParam))
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
		builds, _, _, _, err := travisClient.Builds.ListFromRepository(request.Source.Repository, &travis.BuildListOptions{
			Number: request.Version.BuildNumber,
		})
		common.FatalIf("can't get build", err)
		if len(builds) == 0 {
			common.FatalIf("can't get build", errors.New("there is no builds in travis"))
		}
		build = builds[0]
	}

	travisClient.Builds.Restart(build.Id)
	response := model.InResponse{common.GetMetadatasFromBuild(build), model.Version{build.Number}}
	json.NewEncoder(os.Stdout).Encode(response)
}
