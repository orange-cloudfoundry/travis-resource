package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"

	"github.com/alphagov/travis-resource/common"
	"github.com/alphagov/travis-resource/messager"
	"github.com/alphagov/travis-resource/model"
	. "github.com/alphagov/travis-resource/out/command"
	"github.com/shuheiktgw/go-travis"
)

func main() {
	ctx := context.Background()
	mes := messager.GetMessager()
	var request model.OutRequest
	err := json.NewDecoder(os.Stdin).Decode(&request)
	mes.FatalIf("failed to read request", err)
	if request.Source.Repository == "" {
		mes.FatalIf("can't get build", errors.New("there is no repository set"))
	}
	travisClient, err := common.MakeTravisClient(ctx, request.Source)
	mes.FatalIf("failed to create travis client", err)

	var build *travis.Build
	outCommand := &OutCommand{travisClient, request, "", mes}
	outCommand.LoadRepository()

	buildParam := outCommand.GetBuildParam()

	build, err = outCommand.GetBuild(ctx, buildParam)
	mes.FatalIf("fetch build error", err)

	build, err = outCommand.Restart(ctx, build)
	mes.FatalIf("can't get build after restart", err)

	outCommand.SendResponse(build)
}
