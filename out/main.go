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
	if len(os.Args) <= 1 {
		common.FatalIf("error in command argument", errors.New("you must pass a folder as a first argument"))
	}
	destinationFolder := os.Args[1]
	var request model.OutRequest
	err := json.NewDecoder(os.Stdin).Decode(&request)
	common.FatalIf("failed to read request", err)
	if request.Source.Repository == "" {
		common.FatalIf("can't get build", errors.New("there is no repository set"))
	}
	travisClient, err := common.MakeTravisClient(request.Source)

	common.FatalIf("failed to create travis client", err)
	file, err := os.Open(destinationFolder + "/" + common.FILENAME_BUILD_INFO)
	common.FatalIf("can't open file", err)
	defer file.Close()

	var listBuilds travis.ListBuildsResponse
	err = json.NewDecoder(file).Decode(&listBuilds)
	common.FatalIf("failed to read builds informations", err)

	build := listBuilds.Builds[0]
	repository := request.Source.Repository
	if request.OutParams.Repository != "" {
		repository = request.OutParams.Repository
	}
	buildParam := ""
	if buildParamInt, ok := request.OutParams.Build.(int); ok {
		buildParam = strconv.Itoa(buildParamInt)
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
	}

	travisClient.Builds.Restart(build.Id)
	buildNumberInt, err := strconv.Atoi(build.Number)
	common.FatalIf("build number invalid", err)
	response := model.InResponse{common.GetMetadatasFromBuild(build), model.Version{buildNumberInt}}
	json.NewEncoder(os.Stdout).Encode(response)
}
