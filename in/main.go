package main

import (
	"errors"
	"os"
	"github.com/Orange-OpenSource/travis-resource/model"
	"encoding/json"
	"github.com/Orange-OpenSource/travis-resource/common"
	"github.com/Orange-OpenSource/travis-resource/travis"
	"path/filepath"
	"fmt"
)

const (
	LOGS_FOLDER string = "travis-logs"
	LOGS_FILENAME_PATTERN = "job-%d.log"
)

type InCommand struct {
	travisClient      *travis.Client
	request           model.InRequest
	destinationFolder string
}

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
	build, listBuild, err := inCommand.getBuildInfo()
	common.FatalIf("can't get build", err)

	err = inCommand.writeInBuildInfoFile(listBuild)
	common.FatalIf("can't create file build info", err)

	err = inCommand.downloadLogs(build)
	common.FatalIf("can't download logs", err)

	response := model.InResponse{
		Metadata: common.GetMetadatasFromBuild(build),
		Version: model.Version{request.Version.BuildNumber},
	}
	json.NewEncoder(os.Stdout).Encode(response)
}
func (this *InCommand) getBuildInfo() (travis.Build, travis.ListBuildsResponse, error) {
	var build travis.Build
	var err error
	var listBuild travis.ListBuildsResponse

	builds, jobs, commits, _, err := this.travisClient.Builds.ListFromRepository(this.request.Source.Repository, &travis.BuildListOptions{
		Number: this.request.Version.BuildNumber,
	})
	if err != nil {
		return build, listBuild, err
	}
	common.FatalIf("can't get build", err)
	if len(builds) == 0 {
		return build, listBuild, errors.New("there is no builds in travis")
	}
	build = builds[0]
	listBuild = travis.ListBuildsResponse{
		Builds: builds,
		Jobs: jobs,
		Commits: commits,
	}
	return build, listBuild, nil
}
func (this *InCommand) writeInBuildInfoFile(listBuild travis.ListBuildsResponse) error {
	file, err := os.Create(filepath.Join(this.destinationFolder, common.FILENAME_BUILD_INFO))
	if err != nil {
		return err
	}
	defer file.Close()
	listBuildJson, err := json.MarshalIndent(listBuild, "", "\t")
	if err != nil {
		return err
	}
	file.Write(listBuildJson)
	return nil
}
func (this *InCommand) downloadLogs(build travis.Build) error {
	if !this.request.InParams.DownloadLogs {
		return nil
	}
	err := os.MkdirAll(filepath.Join(this.destinationFolder, LOGS_FOLDER), 0755)
	if err != nil {
		return err
	}
	for _, jobId := range build.JobIds {
		err = this.downloadLogFromJob(jobId)
		if err != nil {
			return err
		}
	}
	return nil
}
func (this *InCommand) downloadLogFromJob(jobId uint) error {
	file, err := os.Create(filepath.Join(this.destinationFolder, LOGS_FOLDER, fmt.Sprintf(LOGS_FILENAME_PATTERN, jobId)))
	if err != nil {
		return err
	}
	defer file.Close()
	logs, _, err := this.travisClient.Jobs.RawLog(jobId)
	if err != nil {
		return err
	}
	file.Write(logs)
	return nil
}