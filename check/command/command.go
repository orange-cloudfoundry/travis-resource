package command

import (
	"github.com/Orange-OpenSource/travis-resource/travis"
	"github.com/Orange-OpenSource/travis-resource/model"
	"io"
	"encoding/json"
	"errors"
)

type CheckCommand struct {
	TravisClient *travis.Client
	Request      model.CheckRequest
}

func (this *CheckCommand) SendResponse(buildNumber string, w io.Writer) {
	response := model.CheckResponse{model.Version{buildNumber}}
	json.NewEncoder(w).Encode(response)
}
func (this *CheckCommand) GetBuildNumber() (string, error) {
	var builds []travis.Build
	var err error
	if this.Request.Source.Branch != "" {
		if this.Request.Source.CheckAllBuilds {
			builds, _, _, _, err = this.TravisClient.Builds.ListFromRepositoryWithBranch(this.Request.Source.Repository, this.Request.Source.Branch, nil)
		} else {
			builds, _, _, _, err = this.TravisClient.Builds.ListSucceededFromRepositoryWithBranch(this.Request.Source.Repository, this.Request.Source.Branch, nil)
		}
	} else {
		if this.Request.Source.CheckAllBuilds {
			builds, _, _, _, err = this.TravisClient.Builds.ListFromRepository(this.Request.Source.Repository, nil)
		} else {
			builds, _, _, _, err = this.TravisClient.Builds.ListSucceededFromRepository(this.Request.Source.Repository, nil)
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
