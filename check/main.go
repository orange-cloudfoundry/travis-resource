package main

import (
	"os"
	"github.com/Orange-OpenSource/travis-resource/model"
	"encoding/json"
	"errors"
	"github.com/Orange-OpenSource/travis-resource/common"
	. "github.com/Orange-OpenSource/travis-resource/check/command"
	"github.com/Orange-OpenSource/travis-resource/messager"
)

func main() {
	var request model.CheckRequest
	err := json.NewDecoder(os.Stdin).Decode(&request)
	mes := messager.GetMessager()
	mes.FatalIf("failed to read request", err)
	if request.Source.Repository == "" {
		mes.FatalIf("can't get build", errors.New("there is no repository set"))
	}
	travisClient, err := common.MakeTravisClient(request.Source)
	mes.FatalIf("failed to create travis client", err)
	checkCommand := CheckCommand{travisClient, request, mes}

	buildNumber, err := checkCommand.GetBuildNumber()
	mes.FatalIf("can't get build", err)

	checkCommand.SendResponse(buildNumber)
}
