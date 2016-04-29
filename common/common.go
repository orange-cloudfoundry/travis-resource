package common

import (
	"github.com/Orange-OpenSource/travis-resource/travis"
	"os"
	"net/http"
	"github.com/Orange-OpenSource/travis-resource/model"
	"time"
	"net"
	"strconv"
	"strings"
)

const (
	FILENAME_BUILD_INFO string = "build-info.json"
)

func MakeTravisClient(request model.Source) (*travis.Client, error) {

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
	var travisUrl string
	if request.Url != "" {
		travisUrl = request.Url
	} else {
		travisUrl = GetTravisUrl(request.Pro)
	}
	if request.GithubToken != "" {
		travisClient = travis.NewClient(travisUrl, "", httpClient)
		_, _, err := travisClient.Authentication.UsingGithubToken(request.GithubToken)
		if err != nil {
			return nil, err
		}
	} else {
		travisClient = travis.NewClient(travisUrl, request.TravisToken, httpClient)
	}
	return travisClient, nil
}
func GetMetadatasFromBuild(build travis.Build) ([]model.Metadata) {
	metadatas := make([]model.Metadata, 0)
	metadatas = append(metadatas, model.Metadata{"travis_succeeded", strconv.FormatBool(build.State == travis.SUCCEEDED_STATE)})
	metadatas = append(metadatas, model.Metadata{"travis_build_state", build.State})
	return metadatas
}
func GetTravisUrl(pro bool) (string) {
	if pro {
		return travis.TRAVIS_API_PRO_URL
	}
	return travis.TRAVIS_API_DEFAULT_URL
}
func GetTravisDashboardUrl(url string) string {
	return strings.Replace(url, "api.", "", 1)
}
func FatalIf(doing string, err error) {
	if err != nil {
		Fatal(doing + ": " + err.Error())
	}
}
func StringInSlice(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}
func Fatal(message string) {
	println(message)
	os.Exit(1)
}