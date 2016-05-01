package command

import (
	"github.com/Orange-OpenSource/travis-resource/travis"
	"github.com/Orange-OpenSource/travis-resource/model"
	"errors"
	"github.com/Orange-OpenSource/travis-resource/messager"
)

type CheckCommand struct {
	TravisClient *travis.Client
	Request      model.CheckRequest
	Messager     *messager.ResourceMessager
}

func (c *CheckCommand) SendResponse(buildNumber string) {
	response := model.CheckResponse{model.Version{buildNumber}}
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
	builds, _, _, _, err = c.TravisClient.Builds.ListFromRepositoryWithInfos(c.Request.Source.Repository, c.Request.Source.Branch, state, nil)
	if err != nil {
		return "", err
	}
	if len(builds) == 0 {
		return "", errors.New("there is no builds in travis")
	}
	return builds[0].Number, nil
}
