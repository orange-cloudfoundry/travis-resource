package command

import (
	"github.com/Orange-OpenSource/travis-resource/common"
	"github.com/Orange-OpenSource/travis-resource/model"
	"errors"
	"github.com/Orange-OpenSource/travis-resource/travis"
	"strconv"
	"time"
	"github.com/Orange-OpenSource/travis-resource/messager"
	"fmt"
)

type OutCommand struct {
	TravisClient *travis.Client
	Request      model.OutRequest
	Repository   string
	Messager     *messager.ResourceMessager
}

func (c *OutCommand) LoadRepository() {
	c.Repository = c.Request.Source.Repository
	if c.Request.OutParams.Repository != "" {
		c.Repository = c.Request.OutParams.Repository
	}
}
func (c *OutCommand) GetBuildParam() string {
	buildParam := ""
	if buildParamNumber, ok := c.Request.OutParams.Build.(float64); ok {
		buildParam = strconv.FormatFloat(buildParamNumber, 'f', 0, 64)
	}
	if buildParamString, ok := c.Request.OutParams.Build.(string); ok {
		buildParam = buildParamString
	}
	return buildParam
}
func (c *OutCommand) SendResponse(build travis.Build) {
	response := model.InResponse{
		Metadata: common.GetMetadatasFromBuild(build),
		Version: model.Version{build.Number},
	}
	c.Messager.SendJsonResponse(response)
}
func (c *OutCommand) Restart(build travis.Build) (travis.Build, error) {
	c.TravisClient.Builds.Restart(build.Id)
	c.Messager.LogItLn("Build '%s' on repository '[blue]%s[reset]' restarted, see details here: [blue]%s[reset] .\n", build.Number, c.Repository, c.GetBuildUrl(build))
	if !c.Request.OutParams.SkipWait {
		c.WaitBuild(build.Number)
	}
	return c.TravisClient.Builds.GetFirstBuildFromBuildNumber(c.Repository, build.Number)
}
func (c *OutCommand) WaitBuild(buildNumber string) {
	var build travis.Build
	var err error
	c.Messager.LogIt("Wait build to finish on travis")
	for {
		build, err = c.TravisClient.Builds.GetFirstBuildFromBuildNumber(c.Repository, buildNumber)
		common.FatalIf("can't get build after restart", err)
		if !common.StringInSlice(build.State, travis.RUNNING_STATE) {
			break
		}
		c.Messager.LogIt(".")
		time.Sleep(5 * time.Second)
	}
	c.Messager.LogIt(build.State)
	if build.State != travis.SUCCEEDED_STATE {
		common.FatalIf("Build '" + build.Number + "' failed",
			errors.New("\n\tstate: " + build.State + "\n\tsee: " + c.GetBuildUrl(build)))
	}
}
func (c *OutCommand) GetBuildUrl(build travis.Build) string {
	var travisUrl string
	if c.Request.Source.Url != "" {
		travisUrl = c.Request.Source.Url
	} else {
		travisUrl = common.GetTravisUrl(c.Request.Source.Pro)
	}
	travisUrl = common.GetTravisDashboardUrl(travisUrl)
	return travisUrl + c.Repository + "/builds/" + strconv.Itoa(int(build.Id))
}
func (c *OutCommand) GetBuildUrlLink(build travis.Build) string {
	buildUrl := c.GetBuildUrl(build)
	return fmt.Sprintf("<a href=\"%s\">%s</a>", buildUrl, buildUrl)
}
func (c *OutCommand) GetBuild(buildParam string) (travis.Build, error) {
	var build travis.Build
	var err error
	if buildParam == "latest" || (c.Request.OutParams.Repository != "" && c.Request.OutParams.Build == "" && c.Request.OutParams.Branch == "") {
		build, err = c.TravisClient.Builds.GetFirstFinishedBuild(c.Repository)
		if err != nil {
			return build, errors.New("can't get build for repository " + c.Repository + " with latest build " + err.Error())
		}
		return build, nil
	} else if buildParam != "" {
		build, err = c.TravisClient.Builds.GetFirstBuildFromBuildNumber(c.Repository, buildParam)
		if err != nil {
			return build, errors.New("can't get build for repository " + c.Repository + " with build " + buildParam + " " + err.Error())
		}
		return build, nil
	} else if c.Request.OutParams.Branch != "" {
		build, err = c.TravisClient.Builds.GetFirstFinishedBuildWithBranch(c.Repository, c.Request.OutParams.Branch)
		if err != nil {
			return build, errors.New("can't get build for repository " + c.Repository + " with branch " + c.Request.OutParams.Branch + " " + err.Error())
		}
		return build, nil
	}
	build, err = c.TravisClient.Builds.GetFirstBuildFromBuildNumber(c.Repository, c.Request.Version.BuildNumber)
	if err != nil {
		return build, errors.New("can't get build for repository " + c.Repository + " with build " + c.Request.Version.BuildNumber + " " + err.Error())
	}
	return build, nil

}

