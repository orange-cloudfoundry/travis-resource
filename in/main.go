package main

import (
	"errors"
	"os"
	"github.com/Orange-OpenSource/travis-resource/model"
	"encoding/json"
	"github.com/Orange-OpenSource/travis-resource/common"
	. "github.com/Orange-OpenSource/travis-resource/in/command"
	"github.com/Orange-OpenSource/travis-resource/messager"
)

func main() {
	mes := messager.GetMessager()
	if len(os.Args) <= 1 {
		mes.FatalIf("error in command argument", errors.New("you must pass a folder as a first argument"))
	}
	destinationFolder := os.Args[1]
	err := os.MkdirAll(destinationFolder, 0755)
	if err != nil {
		mes.FatalIf("creating destination", err)
	}
	var request model.InRequest
	err = json.NewDecoder(os.Stdin).Decode(&request)
	mes.FatalIf("failed to read request ", err)

	if request.Source.Repository == "" {
		mes.FatalIf("can't get build", errors.New("there is no repository set"))
	}

	travisClient, err := common.MakeTravisClient(request.Source)
	mes.FatalIf("failed to create travis client", err)

	inCommand := &InCommand{travisClient, request, destinationFolder, mes}
	build, listBuild, err := inCommand.GetBuildInfo()
	mes.FatalIf("can't get build", err)

	err = inCommand.WriteInBuildInfoFile(listBuild)
	mes.FatalIf("can't create file build info", err)

	err = inCommand.DownloadLogs(build)
	mes.FatalIf("can't download logs", err)

	commit, _, err := travisClient.Commits.GetFromBuild(build.Id)
	mes.FatalIf("can't get commit", err)

	inCommand.SendResponse(build, *commit)
}
