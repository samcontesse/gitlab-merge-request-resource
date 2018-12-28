package main

import (
	"encoding/json"
	"github.com/samcontesse/gitlab-merge-request-resource"
	"github.com/samcontesse/gitlab-merge-request-resource/common"
	"github.com/samcontesse/gitlab-merge-request-resource/out"
	"github.com/xanzy/go-gitlab"
	"io/ioutil"
	"log"
	"os"
	"path"
)

func main() {

	var request out.Request

	if len(os.Args) < 2 {
		println("usage: " + os.Args[0] + " <destination>")
		os.Exit(1)
	}

	if err := json.NewDecoder(os.Stdin).Decode(&request); err != nil {
		common.Fatal("reading request from stdin", err)
	}

	path := path.Join(os.Args[1], request.Params.Repository)
	if err := os.Chdir(path); err != nil {
		common.Fatal("changing directory to "+path, err)
	}

	raw, err := ioutil.ReadFile(".git/merge-request.json")
	if err != nil {
		common.Fatal("unmarshalling merge request information", err)
	}

	var mr gitlab.MergeRequest
	json.Unmarshal(raw, &mr)

	api := gitlab.NewClient(common.GetDefaultClient(request.Source.Insecure), request.Source.PrivateToken)
	api.SetBaseURL(request.Source.GetBaseURL())

	if request.Params.Status != "" {
		state := gitlab.BuildState(gitlab.BuildStateValue(request.Params.Status))
		target := request.Source.GetTargetURL()
		name := resource.GetPipelineName()
		options := gitlab.SetCommitStatusOptions{
			Name:      &name,
			TargetURL: &target,
			State:     *state,
		}

		_, res, err := api.Commits.SetCommitStatus(mr.ProjectID, mr.SHA, &options)
		if res.StatusCode != 201 {
			body, _ := ioutil.ReadAll(res.Body)
			log.Fatalf("status code unexpected: %d, response %s", res.StatusCode, string(body))
		}
		if err != nil {
			common.Fatal("Set commit status failed", err)
		}
	}

	if request.Params.Labels != nil {
		options := gitlab.UpdateMergeRequestOptions{
			Labels: request.Params.Labels,
		}
		_, res, err := api.MergeRequests.UpdateMergeRequest(mr.ProjectID, mr.IID, &options)
		if res.StatusCode != 200 {
			body, _ := ioutil.ReadAll(res.Body)
			log.Fatalf("status code unexpected: %d, response %s", res.StatusCode, string(body))
		}
		if err != nil {
			common.Fatal("Update merge request failed", err)
		}
	}

	response := out.Response{Version: resource.Version{
		ID:        mr.IID,
		UpdatedAt: mr.UpdatedAt,
	}}

	json.NewEncoder(os.Stdout).Encode(response)

}
