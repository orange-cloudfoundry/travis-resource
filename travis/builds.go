// Copyright (c) 2015 Ableton AG, Berlin. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// Fragments of this file have been copied from the go-github (https://github.com/google/go-github)
// project, and is therefore licensed under the following copyright:
// Copyright 2013 The go-github AUTHORS. All rights reserved.

package travis

import (
	"fmt"
	"net/http"
	"errors"
)

var SUCCEEDED_STATE string = "passed"
var RUNNING_STATE []string = []string{"created", "started"}
// BuildsService handles communication with the builds
// related methods of the Travis CI API.
type BuildsService struct {
	client *Client
}

// Build represents a Travis CI build
type Build struct {
	Id                uint   `json:"id,omitempty"`
	RepositoryId      uint   `json:"repository_id,omitempty"`
	Slug              string `json:"slug,omitempty"`
	CommitId          uint   `json:"commit_id,omitempty"`
	Number            string `json:"number,omitempty"`
	// Config            Config `json:"config,omitempty"`
	PullRequest       bool   `json:"pull_request,omitempty"`
	PullRequestTitle  string `json:"pull_request_title,omitempty"`
	PullRequestNumber uint   `json:"pull_request_number,omitempty"`
	State             string `json:"state,omitempty"`
	StartedAt         string `json:"started_at,omitempty"`
	FinishedAt        string `json:"finished_at,omitempty"`
	Duration          uint   `json:"duration,omitempty"`
	JobIds            []uint `json:"job_ids,omitempty"`
	AfterNumber       uint   `json:"after_number,omitempty"`
	EventType         string `json:"event_type,omitempty"`
}

// listBuildsResponse represents the response of a call
// to the Travis CI list builds endpoint.
type ListBuildsResponse struct {
	Builds  []Build  `json:"builds,omitempty"`
	Commits []Commit `json:"commits,omitempty"`
	Jobs    []Job    `json:"jobs,omitempty"`
}

// getBuildResponse represents the response of a call
// to the Travis CI get build endpoint.
type getBuildResponse struct {
	Build  Build  `json:"build"`
	Commit Commit `json:"commit"`
	Jobs   []Job  `json:"jobs"`
}

// BuildListOptions specifies the optional parameters to the
// BuildsService.List method.
type BuildListOptions struct {
	ListOptions

	Ids          []uint `url:"ids,omitempty"`
	RepositoryId uint   `url:"repository_id,omitempty"`
	Slug         string `url:"slug,omitempty"`
	Number       string `url:"number,omitempty"`
	EventType    string `url:"event_type,omitempty"`
}

// List the builds for the authenticated user.
//
// Travis CI API docs: http://docs.travis-ci.com/api/#builds
func (bs *BuildsService) List(opt *BuildListOptions) ([]Build, []Job, []Commit, *http.Response, error) {
	u, err := urlWithOptions("/builds", opt)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	req, err := bs.client.NewRequest("GET", u, nil, nil)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	var buildsResp ListBuildsResponse
	resp, err := bs.client.Do(req, &buildsResp)
	if err != nil {
		return nil, nil, nil, resp, err
	}

	return buildsResp.Builds, buildsResp.Jobs, buildsResp.Commits, resp, err
}

// List a repository builds based on it's provided slug.
//
// Travis CI API docs: http://docs.travis-ci.com/api/#builds
func (bs *BuildsService) ListFromRepository(slug string, opt *BuildListOptions) ([]Build, []Job, []Commit, *http.Response, error) {
	u, err := urlWithOptions(fmt.Sprintf("/repos/%v/builds", slug), opt)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	req, err := bs.client.NewRequest("GET", u, nil, nil)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	var buildsResp ListBuildsResponse
	resp, err := bs.client.Do(req, &buildsResp)
	if err != nil {
		return nil, nil, nil, resp, err
	}

	return buildsResp.Builds, buildsResp.Jobs, buildsResp.Commits, resp, err
}
func (bs *BuildsService) GetFirstBuildFromBuildNumber(repository string, buildNumber string) (Build, error) {
	builds, _, _, _, err := bs.ListFromRepository(repository, &BuildListOptions{
		Number: buildNumber,
	})
	if err != nil {
		return Build{}, err
	}
	if len(builds) == 0 {
		return Build{}, errors.New("this build doesn't exists")
	}
	return builds[0], nil
}
func (bs *BuildsService) GetFirstFinishedBuild(repository string) (Build, error) {
	builds, _, _, _, err := bs.ListFromRepository(repository, nil)
	if err != nil {
		return Build{}, err
	}
	if len(builds) == 0 {
		return Build{}, errors.New("No build found")
	}
	for _, build := range builds {
		if stringInSlice(build.State, RUNNING_STATE) {
			continue
		}
		return build, nil
	}
	return Build{}, errors.New("No build found")
}
func (bs *BuildsService) GetFirstFinishedBuildWithBranch(repository string, branch string) (Build, error) {
	builds, _, commits, _, err := bs.ListFromRepository(repository, nil)
	if err != nil {
		return Build{}, err
	}
	if len(builds) == 0 {
		return Build{}, errors.New("No build found")
	}
	for index, build := range builds {
		if stringInSlice(build.State, RUNNING_STATE) || len(commits) <= index || commits[index].Branch != branch {
			continue
		}
		return build, nil
	}
	return Build{}, errors.New("No build found")
}
func stringInSlice(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}

func (bs *BuildsService) ListSucceededFromRepository(slug string, opt *BuildListOptions) ([]Build, []Job, []Commit, *http.Response, error) {
	u, err := urlWithOptions(fmt.Sprintf("/repos/%v/builds", slug), opt)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	req, err := bs.client.NewRequest("GET", u, nil, nil)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	var buildsResp ListBuildsResponse
	resp, err := bs.client.Do(req, &buildsResp)
	if err != nil {
		return nil, nil, nil, resp, err
	}
	succeededBuilds := make([]Build, 0)
	succeededCommits := make([]Commit, 0)
	succeededJobs := make([]Job, 0)
	for index, build := range buildsResp.Builds {
		if build.State != SUCCEEDED_STATE {
			continue
		}
		succeededBuilds = append(succeededBuilds, build)
		if len(buildsResp.Jobs) > index {
			succeededJobs = append(succeededJobs, buildsResp.Jobs[index])
		}
		if len(buildsResp.Commits) > index {
			succeededCommits = append(succeededCommits, buildsResp.Commits[index])
		}
	}
	return succeededBuilds, succeededJobs, succeededCommits, resp, err
}
func (bs *BuildsService) ListSucceededFromRepositoryWithBranch(slug string, branch string, opt *BuildListOptions) ([]Build, []Job, []Commit, *http.Response, error) {
	u, err := urlWithOptions(fmt.Sprintf("/repos/%v/builds", slug), opt)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	req, err := bs.client.NewRequest("GET", u, nil, nil)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	var buildsResp ListBuildsResponse
	resp, err := bs.client.Do(req, &buildsResp)
	if err != nil {
		return nil, nil, nil, resp, err
	}
	succeededBuilds := make([]Build, 0)
	succeededCommits := make([]Commit, 0)
	succeededJobs := make([]Job, 0)
	for index, build := range buildsResp.Builds {
		if build.State != SUCCEEDED_STATE || len(buildsResp.Commits) <= index || buildsResp.Commits[index].Branch != branch {
			continue
		}
		succeededBuilds = append(succeededBuilds, build)
		succeededCommits = append(succeededCommits, buildsResp.Commits[index])
		if len(buildsResp.Jobs) > index {
			succeededJobs = append(succeededJobs, buildsResp.Jobs[index])
		}
	}
	return succeededBuilds, succeededJobs, succeededCommits, resp, err
}

func (bs *BuildsService) ListFromRepositoryWithBranch(slug string, branch string, opt *BuildListOptions) ([]Build, []Job, []Commit, *http.Response, error) {
	u, err := urlWithOptions(fmt.Sprintf("/repos/%v/builds", slug), opt)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	req, err := bs.client.NewRequest("GET", u, nil, nil)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	var buildsResp ListBuildsResponse
	resp, err := bs.client.Do(req, &buildsResp)
	if err != nil {
		return nil, nil, nil, resp, err
	}
	succeededBuilds := make([]Build, 0)
	succeededCommits := make([]Commit, 0)
	succeededJobs := make([]Job, 0)
	for index, build := range buildsResp.Builds {
		if len(buildsResp.Commits) <= index || buildsResp.Commits[index].Branch != branch {
			continue
		}
		succeededBuilds = append(succeededBuilds, build)
		succeededCommits = append(succeededCommits, buildsResp.Commits[index])
		if len(buildsResp.Jobs) > index {
			succeededJobs = append(succeededJobs, buildsResp.Jobs[index])
		}
	}
	return succeededBuilds, succeededJobs, succeededCommits, resp, err
}
// Get fetches a build based on the provided id.
//
// Travis CI API docs: http://docs.travis-ci.com/api/#builds
func (bs *BuildsService) Get(id uint) (*Build, []Job, *Commit, *http.Response, error) {
	u, err := urlWithOptions(fmt.Sprintf("/builds/%d", id), nil)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	req, err := bs.client.NewRequest("GET", u, nil, nil)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	var buildResp getBuildResponse
	resp, err := bs.client.Do(req, &buildResp)
	if err != nil {
		return nil, nil, nil, resp, err
	}

	return &buildResp.Build, buildResp.Jobs, &buildResp.Commit, resp, err
}

// Cancel build with the provided id.
//
// Travis CI API docs: http://docs.travis-ci.com/api/#builds
func (bs *BuildsService) Cancel(id uint) (*http.Response, error) {
	u, err := urlWithOptions(fmt.Sprintf("/builds/%d/cancel", id), nil)
	if err != nil {
		return nil, err
	}

	req, err := bs.client.NewRequest("POST", u, nil, nil)
	if err != nil {
		return nil, err
	}

	resp, err := bs.client.Do(req, nil)
	if err != nil {
		return resp, err
	}

	return resp, err
}

// Restart build with the provided id.
//
// Travis CI API docs: http://docs.travis-ci.com/api/#builds
func (bs *BuildsService) Restart(id uint) (*http.Response, error) {
	u, err := urlWithOptions(fmt.Sprintf("/builds/%d/restart", id), nil)
	if err != nil {
		return nil, err
	}

	req, err := bs.client.NewRequest("POST", u, nil, nil)
	if err != nil {
		return nil, err
	}

	resp, err := bs.client.Do(req, nil)
	if err != nil {
		return resp, err
	}

	return resp, err
}
