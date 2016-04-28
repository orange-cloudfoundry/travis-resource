package main

import (
	"errors"
	"os"
	"github.com/ArthurHlt/travis-resource/model"
	"encoding/json"
	"github.com/ArthurHlt/travis-resource/common"
	"github.com/ArthurHlt/travis-resource/travis"
	"strconv"
)

func main() {
	if len(os.Args) <= 1 {
		common.FatalIf("error in command argument", errors.New("you must pass a folder as a first argument"))
	}
	destinationFolder := os.Args[1]
	var request model.InRequest
	err := json.NewDecoder(os.Stdin).Decode(&request)
	common.FatalIf("failed to read request", err)
	if request.Source.Repository == "" {
		common.FatalIf("can't get build", errors.New("there is no repository set"))
	}
	travisClient, err := common.MakeTravisClient(request.Source)
	common.FatalIf("failed to create travis client", err)
	builds, jobs, commits, _, err := travisClient.Builds.ListFromRepository(request.Source.Repository, &travis.BuildListOptions{
		Number: strconv.Itoa(request.Version.BuildNumber),
	})
	common.FatalIf("can't get build", err)
	if len(builds) == 0 {
		common.FatalIf("can't get build", errors.New("there is no builds in travis"))
	}
	build := builds[0]
	listBuild := travis.ListBuildsResponse{
		Builds: builds,
		Jobs: jobs,
		Commits: commits,
	}

	file, err := os.Create(destinationFolder + "/" + common.FILENAME_BUILD_INFO)
	common.FatalIf("can't create file", err)
	defer file.Close()

	listBuildJson, err := json.MarshalIndent(listBuild, "", "\t")
	common.FatalIf("error during marshall", err)
	file.Write(listBuildJson)

	response := model.InResponse{common.GetMetadatasFromBuild(build), model.Version{request.Version.BuildNumber}}
	json.NewEncoder(os.Stdout).Encode(response)
}
