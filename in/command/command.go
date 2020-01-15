package command

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/Orange-OpenSource/travis-resource/common"
	"github.com/Orange-OpenSource/travis-resource/messager"
	"github.com/Orange-OpenSource/travis-resource/model"
	"github.com/cheggaaa/pb"
	"github.com/shuheiktgw/go-travis"
)

const (
	LOGS_FOLDER           string = "travis-logs"
	LOGS_FILENAME_PATTERN        = "job-%d.log"
)

type InCommand struct {
	TravisClient      *travis.Client
	Request           model.InRequest
	DestinationFolder string
	Messager          *messager.ResourceMessager
}

func (c *InCommand) SendResponse(build *travis.Build) {

	response := model.InResponse{
		Metadata: common.GetMetadatasFromBuild(*build),
		Version:  model.Version{fmt.Sprint(*build.Id)},
	}
	c.Messager.SendJsonResponse(response)
}
func (c *InCommand) GetBuildInfo(ctx context.Context) (*travis.Build, error) {
	var build *travis.Build
	var err error

	options := travis.BuildOption{
		Include: []string{"build.commit"},
	}

	buildId, _ := strconv.ParseUint(c.Request.Version.BuildId, 10, 32)
	build, _, err = c.TravisClient.Builds.Find(ctx, uint(buildId), &options)
	if err != nil {
		return build, err
	}

	return build, nil
}
func (c *InCommand) WriteInBuildInfoFile(build *travis.Build) error {
	fileLocation := filepath.Join(c.DestinationFolder, common.FILENAME_BUILD_INFO)
	c.Messager.LogItLn("Writing build informations in file '[blue]%s[reset]' ...", common.FILENAME_BUILD_INFO)
	file, err := os.Create(fileLocation)
	if err != nil {
		return err
	}
	defer file.Close()
	buildJson, err := json.MarshalIndent(*build, "", "\t")
	if err != nil {
		return err
	}
	c.Messager.LogItLn("Build informations wrote: '[blue]%s[reset]'", string(buildJson))
	file.Write(buildJson)
	c.Messager.LogItLn("Finished to write in file '[blue]%s[reset]' .\n", fileLocation)
	return nil
}
func (c *InCommand) DownloadLogs(ctx context.Context, build *travis.Build) error {
	if !c.Request.InParams.DownloadLogs {
		return nil
	}
	logsLocation := filepath.Join(c.DestinationFolder, LOGS_FOLDER)
	c.Messager.LogItLn("Downloading logs in folder '[blue]%s[reset]' ...", LOGS_FOLDER)
	err := os.MkdirAll(logsLocation, 0755)
	if err != nil {
		return err
	}
	for _, job := range build.Jobs {
		err = c.downloadLogFromJob(ctx, *job.Id)
		if err != nil {
			return err
		}
	}
	c.Messager.LogItLn("Finished to download logs in folder '[blue]%s[reset]' .\n", LOGS_FOLDER)
	return nil
}
func (c *InCommand) downloadLogFromJob(ctx context.Context, jobId uint) error {
	logLocation := filepath.Join(LOGS_FOLDER, fmt.Sprintf(LOGS_FILENAME_PATTERN, jobId))
	file, err := os.Create(filepath.Join(c.DestinationFolder, logLocation))
	if err != nil {
		return err
	}
	defer file.Close()
	c.Messager.LogItLn("-------\nDownloading log for job '[blue]%d[reset]' as file '[blue]%s[reset]' ...", jobId, logLocation)
	_, resp, err := c.TravisClient.Logs.FindByJobId(ctx, jobId)
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
