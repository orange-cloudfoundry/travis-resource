package main

import (
	"os"
	"github.com/Orange-OpenSource/travis-resource/model"
	"encoding/json"
	"errors"
	"github.com/Orange-OpenSource/travis-resource/common"
	. "github.com/Orange-OpenSource/travis-resource/check/command"
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
	checkCommand := CheckCommand{travisClient, request}

	buildNumber, err := checkCommand.GetBuildNumber()
	common.FatalIf("can't get build", err)

	checkCommand.SendResponse(buildNumber, os.Stdout)
}
