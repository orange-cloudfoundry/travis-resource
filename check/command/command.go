package command

import (
	"github.com/Orange-OpenSource/travis-resource/travis"
	"github.com/Orange-OpenSource/travis-resource/model"
	"github.com/Orange-OpenSource/travis-resource/messager"
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
	var builds []travis.Build
	var err error
	var state string
	state = travis.STATE_PASSED
	if c.Request.Source.CheckOnState != "" {
		state = c.Request.Source.CheckOnState
	}
	if c.Request.Source.CheckAllBuilds {
		state = ""
	}
	builds, _, _, _, err = c.TravisClient.Builds.ListFromRepositoryWithInfos(
		c.Request.Source.Repository,
		c.Request.Source.Branch,
		c.Request.Source.BranchRegex,
		state,
		nil,
	)
	if err != nil {
		return "", err
	}
	if len(builds) == 0 {
		return "", nil
	}
	return builds[0].Number, nil
}
