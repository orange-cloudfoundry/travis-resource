package common

import (
	"github.com/ArthurHlt/travis-resource/travis"
	"os"
	"net/http"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/ArthurHlt/travis-resource/model"
	"time"
	"net"
	"strconv"
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
	metadatas = append(metadatas, model.Metadata{"travis_id", strconv.Itoa(int(build.Id))})
	metadatas = append(metadatas, model.Metadata{"travis_repository_id", strconv.Itoa(int(build.RepositoryId))})
	metadatas = append(metadatas, model.Metadata{"travis_slug", build.Slug})
	metadatas = append(metadatas, model.Metadata{"travis_commit_id", strconv.Itoa(int(build.CommitId))})
	metadatas = append(metadatas, model.Metadata{"travis_number", build.Number})
	metadatas = append(metadatas, model.Metadata{"travis_pull_request", strconv.FormatBool(build.PullRequest)})
	metadatas = append(metadatas, model.Metadata{"travis_pull_request_title", build.PullRequestTitle})
	metadatas = append(metadatas, model.Metadata{"travis_pull_request_number", strconv.Itoa(int(build.PullRequestNumber))})
	metadatas = append(metadatas, model.Metadata{"travis_state", build.State})
	metadatas = append(metadatas, model.Metadata{"travis_started_at", build.StartedAt})
	metadatas = append(metadatas, model.Metadata{"travis_finished_at", build.FinishedAt})
	metadatas = append(metadatas, model.Metadata{"travis_duration", strconv.Itoa(int(build.Duration))})
	metadatas = append(metadatas, model.Metadata{"travis_after_number", strconv.Itoa(int(build.AfterNumber))})
	metadatas = append(metadatas, model.Metadata{"travis_event_type", build.EventType})
	metadatas = append(metadatas, model.Metadata{"travis_succeeded", strconv.FormatBool(build.State == travis.SUCCEEDED_STATE)})
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