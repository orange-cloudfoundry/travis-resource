package command

import (
	"context"

	"github.com/Orange-OpenSource/travis-resource/messager"
	"github.com/Orange-OpenSource/travis-resource/model"
	"github.com/shuheiktgw/go-travis"
)

type CheckCommand struct {
	TravisClient *travis.Client
	Request      model.CheckRequest
	Messager     *messager.ResourceMessager
}

func (c *CheckCommand) SendResponse(buildNumber string) {
	var response model.CheckResponse
	if buildNumber != "" {
		response = model.CheckResponse{model.Version{buildNumber}}
	} else {
		response = model.CheckResponse{}
	}

	c.Messager.SendJsonResponse(response)
}
func (c *CheckCommand) GetBuildNumber() (string, error) {
	state := travis.BuildStatePassed
	if c.Request.Source.CheckOnState != "" {
		state = c.Request.Source.CheckOnState
	}
	if c.Request.Source.CheckAllBuilds {
		state = ""
	}

	options := travis.BuildsByRepoOption{
		BranchName: []string{c.Request.Source.Branch},
		State:      []string{state},
	}

	builds, _, err := c.TravisClient.Builds.ListByRepoSlug(
		context.Background(),
		c.Request.Source.Repository,
		&options,
	)

	if err != nil {
		return "", err
	}
	if len(builds) == 0 {
		return "", nil
	}

	return *builds[0].Number, nil
}
