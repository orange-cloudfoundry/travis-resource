package main

import (
	"os"
	"github.com/Orange-OpenSource/travis-resource/common"
	"github.com/Orange-OpenSource/travis-resource/model"
	"encoding/json"
	"errors"
	"github.com/Orange-OpenSource/travis-resource/travis"
	. "github.com/Orange-OpenSource/travis-resource/out/command"
	"github.com/Orange-OpenSource/travis-resource/messager"
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
	outCommand := &OutCommand{travisClient, request, "", messager.GetMessager()}
	outCommand.LoadRepository()

	buildParam := outCommand.GetBuildParam()

	build, err = outCommand.GetBuild(buildParam)
	common.FatalIf("fetch build error", err)

	build, err = outCommand.Restart(build)
	common.FatalIf("can't get build after restart", err)

	outCommand.SendResponse(build)
}
