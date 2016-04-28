package common

import (
	"github.com/ArthurHlt/travis-resource/travis"
	"os"
	"net/http"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/ArthurHlt/travis-resource/model"
	"time"
	"net"
)

var FILENAME_BUILD_INFO string = "build-info.json"

func MakeTravisClient(request model.DefaultSource) (*travis.Client, error) {
	if request.GithubToken == "" && request.TravisToken == "" {
		return nil, errors.New("Not github token or travis token set")
	}
	baseTransport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).Dial,
		DisableKeepAlives: true,
	}
	httpClient := &http.Client{
		Transport: baseTransport,
	}
	var travisClient *travis.Client
	if request.TravisToken != "" {
		travisClient = travis.NewClient(travis.TRAVIS_API_DEFAULT_URL, request.TravisToken, httpClient)
	} else {
		travisClient = travis.NewClient(travis.TRAVIS_API_DEFAULT_URL, "", httpClient)
		// authWithGithubClient.IsAuthenticated() will return false
		_, _, err := travisClient.Authentication.UsingGithubToken(request.GithubToken)
		if err != nil {
			return nil, err
		}
	}
	return travisClient, nil
}
func GetMetadatasFromBuild(build travis.Build) ([]model.Metadata) {
	metadatas := make([]model.Metadata, 0)
	metadatas = append(metadatas, model.Metadata{"build_id", build.Id})
	metadatas = append(metadatas, model.Metadata{"build_repository_id", build.RepositoryId})
	metadatas = append(metadatas, model.Metadata{"build_slug", build.Slug})
	metadatas = append(metadatas, model.Metadata{"build_commit_id", build.CommitId})
	metadatas = append(metadatas, model.Metadata{"build_number", build.Number})
	metadatas = append(metadatas, model.Metadata{"build_pull_request", build.PullRequest})
	metadatas = append(metadatas, model.Metadata{"build_pull_request_title", build.PullRequestTitle})
	metadatas = append(metadatas, model.Metadata{"build_pull_request_number", build.PullRequestNumber})
	metadatas = append(metadatas, model.Metadata{"build_state", build.State})
	metadatas = append(metadatas, model.Metadata{"build_started_at", build.StartedAt})
	metadatas = append(metadatas, model.Metadata{"build_finished_at", build.FinishedAt})
	metadatas = append(metadatas, model.Metadata{"build_duration", build.Duration})
	metadatas = append(metadatas, model.Metadata{"build_after_number", build.AfterNumber})
	metadatas = append(metadatas, model.Metadata{"build_event_type", build.AfterNumber})
	metadatas = append(metadatas, model.Metadata{"build_succeeded", build.State == travis.SUCCEEDED_STATE})
	return metadatas
}
func FatalIf(doing string, err error) {
	if err != nil {
		Fatal(doing + ": " + err.Error())
	}
}

func Fatal(message string) {
	println(message)
	os.Exit(1)
}