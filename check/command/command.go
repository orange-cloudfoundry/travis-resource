package command

import (
	"context"
	"fmt"

	"github.com/Orange-OpenSource/travis-resource/messager"
	"github.com/Orange-OpenSource/travis-resource/model"
	"github.com/shuheiktgw/go-travis"
)

type CheckCommand struct {
	TravisClient *travis.Client
	Request      model.CheckRequest
	Messager     *messager.ResourceMessager
}

func (c *CheckCommand) SendResponse(buildId string) {
	var response model.CheckResponse
	if buildId != "" {
		response = model.CheckResponse{model.Version{buildId}}
	} else {
		response = model.CheckResponse{}
	}

	c.Messager.SendJsonResponse(response)
}
func (c *CheckCommand) GetBuildId() (string, error) {
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

	return fmt.Sprint(*builds[0].Id), nil
}
