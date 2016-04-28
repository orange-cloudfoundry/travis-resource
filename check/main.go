package main

import (
	"os"
	"github.com/ArthurHlt/travis-resource/model"
	"encoding/json"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/ArthurHlt/travis-resource/common"
	"github.com/ArthurHlt/travis-resource/travis"
)

func main() {
	var request model.CheckRequest
	err := json.NewDecoder(os.Stdin).Decode(&request)
	common.FatalIf("failed to read request", err)
	if request.Source.Repository == "" {
		common.FatalIf("can't get build", errors.New("there is no repository set"))
	}
	travisClient, err := common.MakeTravisClient(request.Source)
	common.FatalIf("failed to create travis client", err)

	var buildNumber string
	var builds []travis.Build
	if request.Source.Branch != "" {
		if request.Source.CheckAllBuilds {
			builds, _, _, _, err = travisClient.Builds.ListFromRepositoryWithBranch(request.Source.Repository, request.Source.Branch, nil)
		} else {
			builds, _, _, _, err = travisClient.Builds.ListSucceededFromRepositoryWithBranch(request.Source.Repository, request.Source.Branch, nil)
		}
	} else {
		if request.Source.CheckAllBuilds {
			builds, _, _, _, err = travisClient.Builds.ListFromRepository(request.Source.Repository, nil)
		} else {
			builds, _, _, _, err = travisClient.Builds.ListSucceededFromRepository(request.Source.Repository, nil)
		}
	}
	common.FatalIf("can't get build", err)
	if len(builds) == 0 {
		common.FatalIf("can't get build", errors.New("there is no builds in travis"))
	}
	buildNumber = builds[0].Number
	response := model.CheckResponse{model.Version{buildNumber}}
	json.NewEncoder(os.Stdout).Encode(response)
}
