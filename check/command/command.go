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
	if c.Request.Source.Branch != "" {
		if c.Request.Source.CheckAllBuilds {
			builds, _, _, _, err = c.TravisClient.Builds.ListFromRepositoryWithBranch(c.Request.Source.Repository, c.Request.Source.Branch, nil)
		} else {
			builds, _, _, _, err = c.TravisClient.Builds.ListSucceededFromRepositoryWithBranch(c.Request.Source.Repository, c.Request.Source.Branch, nil)
		}
	} else {
		if c.Request.Source.CheckAllBuilds {
			builds, _, _, _, err = c.TravisClient.Builds.ListFromRepository(c.Request.Source.Repository, nil)
		} else {
			builds, _, _, _, err = c.TravisClient.Builds.ListSucceededFromRepository(c.Request.Source.Repository, nil)
		}
	}
	if err != nil {
		return "", err
	}
	if len(builds) == 0 {
		return "", errors.New("there is no builds in travis")
	}
	return builds[0].Number, nil
}
