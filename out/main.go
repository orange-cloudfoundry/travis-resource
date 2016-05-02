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
	mes := messager.GetMessager()
	var request model.OutRequest
	err := json.NewDecoder(os.Stdin).Decode(&request)
	mes.FatalIf("failed to read request", err)
	if request.Source.Repository == "" {
		mes.FatalIf("can't get build", errors.New("there is no repository set"))
	}
	travisClient, err := common.MakeTravisClient(request.Source)
	mes.FatalIf("failed to create travis client", err)

	var build travis.Build
	outCommand := &OutCommand{travisClient, request, "", mes}
	outCommand.LoadRepository()

	buildParam := outCommand.GetBuildParam()

	build, err = outCommand.GetBuild(buildParam)
	mes.FatalIf("fetch build error", err)

	build, err = outCommand.Restart(build)
	mes.FatalIf("can't get build after restart", err)

	commit, _, err := travisClient.Commits.GetFromBuild(build.Id)
	mes.FatalIf("can't get commit", err)

	outCommand.SendResponse(build, *commit)
}
