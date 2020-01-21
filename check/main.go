package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"

	. "github.com/alphagov/travis-resource/check/command"
	"github.com/alphagov/travis-resource/common"
	"github.com/alphagov/travis-resource/messager"
	"github.com/alphagov/travis-resource/model"
)

func main() {
	ctx := context.Background()
	var request model.CheckRequest
	err := json.NewDecoder(os.Stdin).Decode(&request)
	mes := messager.GetMessager()
	mes.FatalIf("failed to read request", err)
	if request.Source.Repository == "" {
		mes.FatalIf("can't get build", errors.New("there is no repository set"))
	}
	travisClient, err := common.MakeTravisClient(ctx, request.Source)
	mes.FatalIf("failed to create travis client", err)
	checkCommand := CheckCommand{travisClient, request, mes}

	buildId, err := checkCommand.GetBuildId()
	mes.FatalIf("can't get build", err)

	checkCommand.SendResponse(buildId)
}
