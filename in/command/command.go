package command

import (
	"errors"
	"os"
	"github.com/Orange-OpenSource/travis-resource/model"
	"encoding/json"
	"github.com/Orange-OpenSource/travis-resource/common"
	"github.com/Orange-OpenSource/travis-resource/travis"
	"path/filepath"
	"fmt"
	"io"
)

const (
	LOGS_FOLDER string = "travis-logs"
	LOGS_FILENAME_PATTERN = "job-%d.log"
)

type InCommand struct {
	TravisClient      *travis.Client
	Request           model.InRequest
	DestinationFolder string
}

func (this *InCommand) SendResponse(build travis.Build, w io.Writer) {
	response := model.InResponse{
		Metadata: common.GetMetadatasFromBuild(build),
		Version: model.Version{this.Request.Version.BuildNumber},
	}
	json.NewEncoder(w).Encode(response)
}
func (this *InCommand) GetBuildInfo() (travis.Build, travis.ListBuildsResponse, error) {
	var build travis.Build
	var err error
	var listBuild travis.ListBuildsResponse

	builds, jobs, commits, _, err := this.TravisClient.Builds.ListFromRepository(this.Request.Source.Repository, &travis.BuildListOptions{
		Number: this.Request.Version.BuildNumber,
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
func (this *InCommand) WriteInBuildInfoFile(listBuild travis.ListBuildsResponse) error {
	file, err := os.Create(filepath.Join(this.DestinationFolder, common.FILENAME_BUILD_INFO))
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
func (this *InCommand) DownloadLogs(build travis.Build) error {
	if !this.Request.InParams.DownloadLogs {
		return nil
	}
	err := os.MkdirAll(filepath.Join(this.DestinationFolder, LOGS_FOLDER), 0755)
	if err != nil {
		return err
	}
	for _, jobId := range build.JobIds {
		err = this.DownloadLogFromJob(jobId)
		if err != nil {
			return err
		}
	}
	return nil
}
func (this *InCommand) DownloadLogFromJob(jobId uint) error {
	file, err := os.Create(filepath.Join(this.DestinationFolder, LOGS_FOLDER, fmt.Sprintf(LOGS_FILENAME_PATTERN, jobId)))
	if err != nil {
		return err
	}
	defer file.Close()
	logs, _, err := this.TravisClient.Jobs.RawLog(jobId)
	if err != nil {
		return err
	}
	file.Write(logs)
	return nil
}