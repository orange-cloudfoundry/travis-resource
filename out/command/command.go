package command

import (
	"github.com/Orange-OpenSource/travis-resource/common"
	"github.com/Orange-OpenSource/travis-resource/model"
	"encoding/json"
	"errors"
	"github.com/Orange-OpenSource/travis-resource/travis"
	"strconv"
	"time"
	"io"
)

type OutCommand struct {
	TravisClient *travis.Client
	Request      model.OutRequest
	Repository   string
}

func (this *OutCommand) LoadRepository() {
	this.Repository = this.Request.Source.Repository
	if this.Request.OutParams.Repository != "" {
		this.Repository = this.Request.OutParams.Repository
	}
}
func (this *OutCommand) GetBuildParam() string {
	buildParam := ""
	if buildParamNumber, ok := this.Request.OutParams.Build.(float64); ok {
		buildParam = strconv.FormatFloat(buildParamNumber, 'f', 0, 64)
	}
	if buildParamString, ok := this.Request.OutParams.Build.(string); ok {
		buildParam = buildParamString
	}
	return buildParam
}
func (this *OutCommand) SendResponse(build travis.Build, w io.Writer) {
	response := model.InResponse{
		Metadata: common.GetMetadatasFromBuild(build),
		Version: model.Version{this.Request.Version.BuildNumber},
	}
	json.NewEncoder(w).Encode(response)
}
func (this *OutCommand) Restart(build travis.Build) (travis.Build, error) {
	this.TravisClient.Builds.Restart(build.Id)
	if !this.Request.OutParams.SkipWait {
		this.WaitBuild(build.Number)
	}
	return this.TravisClient.Builds.GetFirstBuildFromBuildNumber(this.Repository, build.Number)
}
func (this *OutCommand) WaitBuild(buildNumber string) {
	var build travis.Build
	var err error
	var travisUrl string
	if this.Request.Source.Url != "" {
		travisUrl = this.Request.Source.Url
	} else {
		travisUrl = common.GetTravisUrl(this.Request.Source.Pro)
	}
	travisUrl = common.GetTravisDashboardUrl(travisUrl)
	for {
		build, err = this.TravisClient.Builds.GetFirstBuildFromBuildNumber(this.Repository, buildNumber)
		common.FatalIf("can't get build after restart", err)
		if !common.StringInSlice(build.State, travis.RUNNING_STATE) {
			break
		}
		time.Sleep(5 * time.Second)
	}
	if build.State != travis.SUCCEEDED_STATE {
		common.FatalIf("Build '" + build.Number + "' failed",
			errors.New("\n\tstate: " + build.State + "\n\tsee: " + travisUrl + this.Repository + "/builds/" + strconv.Itoa(int(build.Id))))
	}
}
func (this *OutCommand) GetBuild(buildParam string) (travis.Build, error) {
	var build travis.Build
	var err error
	if buildParam == "latest" || (this.Request.OutParams.Repository != "" && this.Request.OutParams.Build == "" && this.Request.OutParams.Branch == "") {
		build, err = this.TravisClient.Builds.GetFirstFinishedBuild(this.Repository)
		if err != nil {
			return build, errors.New("can't get build for repository " + this.Repository + " with latest build " + err.Error())
		}
		return build, nil
	} else if buildParam != "" {
		build, err = this.TravisClient.Builds.GetFirstBuildFromBuildNumber(this.Repository, buildParam)
		if err != nil {
			return build, errors.New("can't get build for repository " + this.Repository + " with build " + buildParam + " " + err.Error())
		}
		return build, nil
	} else if this.Request.OutParams.Branch != "" {
		build, err = this.TravisClient.Builds.GetFirstFinishedBuildWithBranch(this.Repository, this.Request.OutParams.Branch)
		if err != nil {
			return build, errors.New("can't get build for repository " + this.Repository + " with branch " + this.Request.OutParams.Branch + " " + err.Error())
		}
		return build, nil
	}
	build, err = this.TravisClient.Builds.GetFirstBuildFromBuildNumber(this.Repository, this.Request.Version.BuildNumber)
	if err != nil {
		return build, errors.New("can't get build for repository " + this.Repository + " with build " + this.Request.Version.BuildNumber + " " + err.Error())
	}
	return build, nil

}

