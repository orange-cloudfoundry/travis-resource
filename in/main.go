package main

import (
	"errors"
	"os"
	"github.com/Orange-OpenSource/travis-resource/model"
	"encoding/json"
	"github.com/Orange-OpenSource/travis-resource/common"
	. "github.com/Orange-OpenSource/travis-resource/in/command"
)

func main() {
	if len(os.Args) <= 1 {
		common.FatalIf("error in command argument", errors.New("you must pass a folder as a first argument"))
	}
	destinationFolder := os.Args[1]
	err := os.MkdirAll(destinationFolder, 0755)
	if err != nil {
		common.FatalIf("creating destination", err)
	}
	var request model.InRequest
	err = json.NewDecoder(os.Stdin).Decode(&request)
	common.FatalIf("failed to read request ", err)
	if request.Source.Repository == "" {
		common.FatalIf("can't get build", errors.New("there is no repository set"))
	}
	travisClient, err := common.MakeTravisClient(request.Source)
	common.FatalIf("failed to create travis client", err)

	inCommand := &InCommand{travisClient, request, destinationFolder}
	build, listBuild, err := inCommand.GetBuildInfo()
	common.FatalIf("can't get build", err)

	err = inCommand.WriteInBuildInfoFile(listBuild)
	common.FatalIf("can't create file build info", err)

	err = inCommand.DownloadLogs(build)
	common.FatalIf("can't download logs", err)

	inCommand.SendResponse(build, os.Stdout)
}
