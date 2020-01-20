package common

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Orange-OpenSource/travis-resource/model"
	"github.com/shuheiktgw/go-travis"
)

const (
	FILENAME_BUILD_INFO string = "build-info.json"
)

func MakeTravisClient(ctx context.Context, request model.Source) (*travis.Client, error) {
	var travisClient *travis.Client
	var travisUrl string
	if request.Url != "" {
		travisUrl = request.Url
	} else {
		travisUrl = GetTravisUrl(request.Pro)
	}
	if request.GithubToken != "" {
		travisClient = travis.NewClient(travisUrl, "")
		_, _, err := travisClient.Authentication.UsingGithubToken(ctx, request.GithubToken)
		if err != nil {
			return nil, err
		}
	} else {
		travisClient = travis.NewClient(travisUrl, request.TravisToken)
	}

	return travisClient, nil
}
func GetMetadatasFromBuild(build travis.Build) []model.Metadata {
	metadatas := make([]model.Metadata, 0)
	metadatas = append(metadatas, model.Metadata{"travis_succeeded", strconv.FormatBool(*build.State == travis.BuildStatePassed)})
	metadatas = append(metadatas, model.Metadata{"travis_build_state", *build.State})
	if *build.State != travis.BuildStateCreated {
		metadatas = append(metadatas, model.Metadata{"travis_started_at", *build.StartedAt})
	}
	if *build.State == travis.BuildStatePassed {
		duration, _ := time.ParseDuration(fmt.Sprint(*build.Duration) + "s")
		metadatas = append(metadatas, model.Metadata{
			Name:  "travis_build_duration",
			Value: duration.String(),
		})
	}
	metadatas = append(metadatas, model.Metadata{"commit_author", build.Commit.Author.Name})
	metadatas = append(metadatas, model.Metadata{"commit_author_date", *build.Commit.CommittedAt})
	metadatas = append(metadatas, model.Metadata{"commit_ref", *build.Commit.Sha})
	metadatas = append(metadatas, model.Metadata{"commit_message", *build.Commit.Message})

	return metadatas
}
func GetTravisUrl(pro bool) string {
	if pro {
		return travis.ApiComUrl
	}
	return travis.ApiOrgUrl
}
func GetTravisDashboardUrl(url string) string {
	return strings.Replace(url, "api.", "", 1)
}

func StringInSlice(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}
