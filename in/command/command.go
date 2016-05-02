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
	"github.com/Orange-OpenSource/travis-resource/messager"
	"github.com/cheggaaa/pb"
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
	Messager          *messager.ResourceMessager
}

func (c *InCommand) SendResponse(build travis.Build, commit travis.Commit) {

	response := model.InResponse{
		Metadata: common.GetMetadatasFromBuild(build, commit),
		Version: model.Version{build.Number},
	}
	c.Messager.SendJsonResponse(response)
}
func (c *InCommand) GetBuildInfo() (travis.Build, travis.ListBuildsResponse, error) {
	var build travis.Build
	var err error
	var listBuild travis.ListBuildsResponse

	builds, jobs, commits, _, err := c.TravisClient.Builds.ListFromRepository(c.Request.Source.Repository, &travis.BuildListOptions{
		Number: c.Request.Version.BuildNumber,
	})
	if err != nil {
		return build, listBuild, err
	}
	if len(builds) == 0 {
		return build, listBuild, errors.New("this build doesn't exists in travis")
	}
	build = builds[0]
	listBuild = travis.ListBuildsResponse{
		Builds: builds,
		Jobs: jobs,
		Commits: commits,
	}
	return build, listBuild, nil
}
func (c *InCommand) WriteInBuildInfoFile(listBuild travis.ListBuildsResponse) error {
	fileLocation := filepath.Join(c.DestinationFolder, common.FILENAME_BUILD_INFO)
	c.Messager.LogItLn("Writing build informations in file '[blue]%s[reset]' ...", common.FILENAME_BUILD_INFO)
	file, err := os.Create(fileLocation)
	if err != nil {
		return err
	}
	defer file.Close()
	listBuildJson, err := json.MarshalIndent(listBuild, "", "\t")
	if err != nil {
		return err
	}
	c.Messager.LogItLn("Build informations wrote: '[blue]%s[reset]'", string(listBuildJson))
	file.Write(listBuildJson)
	c.Messager.LogItLn("Finished to write in file '[blue]%s[reset]' .\n", fileLocation)
	return nil
}
func (c *InCommand) DownloadLogs(build travis.Build) error {
	if !c.Request.InParams.DownloadLogs {
		return nil
	}
	logsLocation := filepath.Join(c.DestinationFolder, LOGS_FOLDER)
	c.Messager.LogItLn("Downloading logs in folder '[blue]%s[reset]' ...", LOGS_FOLDER)
	err := os.MkdirAll(logsLocation, 0755)
	if err != nil {
		return err
	}
	for _, jobId := range build.JobIds {
		err = c.downloadLogFromJob(jobId)
		if err != nil {
			return err
		}
	}
	c.Messager.LogItLn("Finished to download logs in folder '[blue]%s[reset]' .\n", LOGS_FOLDER)
	return nil
}
func (c *InCommand) downloadLogFromJob(jobId uint) error {
	logLocation := filepath.Join(LOGS_FOLDER, fmt.Sprintf(LOGS_FILENAME_PATTERN, jobId))
	file, err := os.Create(filepath.Join(c.DestinationFolder, logLocation))
	if err != nil {
		return err
	}
	defer file.Close()
	c.Messager.LogItLn("-------\nDownloading log for job '[blue]%d[reset]' as file '[blue]%s[reset]' ...", jobId, logLocation)
	resp, err := c.TravisClient.Jobs.RawLogOnlyResponse(jobId)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bar := pb.New64(resp.ContentLength).SetUnits(pb.U_BYTES)
	bar.Output = c.Messager.GetLogWriter()
	bar.Start()
	reader := bar.NewProxyReader(resp.Body)
	_, err = io.Copy(file, reader)
	if err != nil {
		return err
	}
	c.Messager.LogItLn("\nFinished to download log for job '[blue]%d[reset]' as file '[blue]%s[reset]' .\n-------", jobId, logLocation)
	return nil
}