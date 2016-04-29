package main

import (
	"os"
	"github.com/Orange-OpenSource/travis-resource/model"
	"encoding/json"
	"errors"
	"github.com/Orange-OpenSource/travis-resource/common"
	"github.com/Orange-OpenSource/travis-resource/travis"
	"io"
)

type CheckCommand struct {
	travisClient *travis.Client
	request      model.CheckRequest
}

func main() {
	var request model.CheckRequest
	err := json.NewDecoder(os.Stdin).Decode(&request)
	common.FatalIf("failed to read request", err)
	if request.Source.Repository == "" {
		common.FatalIf("can't get build", errors.New("there is no repository set"))
	}
	travisClient, err := common.MakeTravisClient(request.Source)
	common.FatalIf("failed to create travis client", err)
	checkCommand := CheckCommand{travisClient, request}

	buildNumber, err := checkCommand.GetBuildNumber()
	common.FatalIf("can't get build", err)

	checkCommand.SendResponse(buildNumber, os.Stdout)
}
func (this *CheckCommand) SendResponse(buildNumber string, w io.Writer) {
	response := model.CheckResponse{model.Version{buildNumber}}
	json.NewEncoder(w).Encode(response)
}
func (this *CheckCommand) GetBuildNumber() (string, error) {
	var builds []travis.Build
	var err error
	if this.request.Source.Branch != "" {
		if this.request.Source.CheckAllBuilds {
			builds, _, _, _, err = this.travisClient.Builds.ListFromRepositoryWithBranch(this.request.Source.Repository, this.request.Source.Branch, nil)
		} else {
			builds, _, _, _, err = this.travisClient.Builds.ListSucceededFromRepositoryWithBranch(this.request.Source.Repository, this.request.Source.Branch, nil)
		}
	} else {
		if this.request.Source.CheckAllBuilds {
			builds, _, _, _, err = this.travisClient.Builds.ListFromRepository(this.request.Source.Repository, nil)
		} else {
			builds, _, _, _, err = this.travisClient.Builds.ListSucceededFromRepository(this.request.Source.Repository, nil)
		}
	}
	if err != nil {
		return "", err
	}
	if len(builds) == 0 {
		return "", errors.New("there is no builds in travis")
	}
	return builds[0].Number, nil
}