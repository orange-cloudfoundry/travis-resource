package command

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/alphagov/travis-resource/common"
	"github.com/alphagov/travis-resource/messager"
	"github.com/alphagov/travis-resource/model"
	"github.com/shuheiktgw/go-travis"
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
func (c *OutCommand) SendResponse(build *travis.Build) {
	response := model.InResponse{
		Metadata: common.GetMetadatasFromBuild(*build),
		Version:  model.Version{fmt.Sprint(*build.Id)},
	}
	c.Messager.SendJsonResponse(response)
}
func (c *OutCommand) Restart(ctx context.Context, build *travis.Build) (*travis.Build, error) {
	c.TravisClient.Builds.Restart(ctx, *build.Id)
	c.Messager.LogItLn("Build '%s' on repository '[blue]%s[reset]' restarted, see details here: [blue]%s[reset] .\n", *build.Id, c.Repository, c.GetBuildUrl(*build.Id))
	if !c.Request.OutParams.SkipWait {
		c.waitBuild(ctx, *build.Id)
	}

	options := travis.BuildOption{
		Include: []string{"build.commit"},
	}

	build, _, err := c.TravisClient.Builds.Find(ctx, *build.Id, &options)
	if err != nil {
		return build, err
	}
	return build, nil
}
func (c *OutCommand) waitBuild(ctx context.Context, buildId uint) {
	var build *travis.Build
	var err error
	c.Messager.LogIt("Wait build to finish on travis")
	for {
		build, _, err = c.TravisClient.Builds.Find(ctx, buildId, nil)
		c.Messager.FatalIf("can't get build after restart", err)
		if !common.StringInSlice(*build.State, []string{travis.BuildStateCreated, travis.BuildStateReceived, travis.BuildStateStarted}) {
			break
		}
		c.Messager.LogIt(".")
		time.Sleep(5 * time.Second)
	}
	c.Messager.LogIt(*build.State)
	if *build.State != travis.BuildStatePassed {
		c.Messager.FatalIf("Build '"+*build.Number+"' failed",
			errors.New("\n\tstate: "+*build.State+"\n\tsee: "+c.GetBuildUrl(buildId)))
	}
}
func (c *OutCommand) GetBuildUrl(buildId uint) string {
	var travisUrl string
	if c.Request.Source.Url != "" {
		travisUrl = c.Request.Source.Url
	} else {
		travisUrl = common.GetTravisUrl(c.Request.Source.Pro)
	}
	travisUrl = common.GetTravisDashboardUrl(travisUrl)
	return travisUrl + c.Repository + "/builds/" + strconv.Itoa(int(buildId))
}
func (c *OutCommand) GetBuildUrlLink(build travis.Build) string {
	buildUrl := c.GetBuildUrl(*build.Id)
	return fmt.Sprintf("<a href=\"%s\">%s</a>", buildUrl, buildUrl)
}
func (c *OutCommand) GetBuild(ctx context.Context, buildParam string) (*travis.Build, error) {
	var build *travis.Build
	var err error
	if buildParam == "latest" || (c.Request.OutParams.Repository != "" && c.Request.OutParams.Build == "" && c.Request.OutParams.Branch == "") {
		options := travis.BuildsByRepoOption{
			State:   []string{travis.BuildStatePassed, travis.BuildStateFailed, travis.BuildStateErrored, travis.BuildStateCanceled},
			Include: []string{"build.commit"},
			Limit:   1,
		}

		builds, _, err := c.TravisClient.Builds.ListByRepoSlug(
			ctx,
			c.Request.Source.Repository,
			&options,
		)

		if err != nil {
			return nil, errors.New("can't get build for repository " + c.Repository + " with latest build " + err.Error())
		}
		return builds[0], nil
	} else if buildParam != "" {
		options := travis.BuildOption{
			Include: []string{"build.commit"},
		}

		buildId, _ := strconv.ParseUint(buildParam, 10, 32)
		build, _, err = c.TravisClient.Builds.Find(ctx, uint(buildId), &options)
		if err != nil {
			return build, errors.New("can't get build for repository " + c.Repository + " with build " + buildParam + " " + err.Error())
		}
		return build, nil
	} else if c.Request.OutParams.Branch != "" {
		options := travis.BuildsByRepoOption{
			State:      []string{travis.BuildStatePassed, travis.BuildStateFailed, travis.BuildStateErrored, travis.BuildStateCanceled},
			Include:    []string{"build.commit"},
			Limit:      1,
			BranchName: []string{c.Request.OutParams.Branch},
		}

		builds, _, err := c.TravisClient.Builds.ListByRepoSlug(
			ctx,
			c.Request.Source.Repository,
			&options,
		)

		if err != nil {
			return nil, errors.New("can't get build for repository " + c.Repository + " with latest build " + err.Error())
		}
		return builds[0], nil
	}

	options := travis.BuildOption{
		Include: []string{"build.commit"},
	}

	buildId, _ := strconv.ParseUint(buildParam, 10, 32)
	build, _, err = c.TravisClient.Builds.Find(ctx, uint(buildId), &options)
	if err != nil {
		return build, errors.New("can't get build for repository " + c.Repository + " with build " + c.Request.Version.BuildId + " " + err.Error())
	}
	return build, nil
}
