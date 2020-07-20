package main

import (
	"encoding/json"
	"fmt"
	"github.com/samcontesse/gitlab-merge-request-resource"
	"github.com/samcontesse/gitlab-merge-request-resource/common"
	"github.com/samcontesse/gitlab-merge-request-resource/out"
	"github.com/xanzy/go-gitlab"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
)

func main() {

	var request out.Request
	var message string

	if len(os.Args) < 2 {
		println("usage: " + os.Args[0] + " <destination>")
		os.Exit(1)
	}

	if err := json.NewDecoder(os.Stdin).Decode(&request); err != nil {
		common.Fatal("reading request from stdin", err)
	}

	workDirPath := path.Join(os.Args[1], request.Params.Repository)
	if err := os.Chdir(workDirPath); err != nil {
		common.Fatal("changing directory to "+workDirPath, err)
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
		message = message + fmt.Sprintf("Set Status: %s \n", request.Params.Status)
		state := gitlab.BuildState(gitlab.BuildStateValue(request.Params.Status))
		target := request.Source.GetTargetURL()
		name := request.Source.GetPipelineName()
		options := gitlab.SetCommitStatusOptions{
			Name:      &name,
			TargetURL: &target,
			State:     *state,
		}

		_, res, err := api.Commits.SetCommitStatus(mr.SourceProjectID, mr.SHA, &options)
		if res.StatusCode != 201 {
			body, _ := ioutil.ReadAll(res.Body)
			log.Fatalf("Set commit status failed: %d, response %s", res.StatusCode, string(body))
		}
		if err != nil {
			common.Fatal("Set commit status failed", err)
		}
	}

	if request.Params.Labels != nil {
		message = message + fmt.Sprintf("Set Labels: %s \n", request.Params.Labels)
		// Make sure `Labels` is present in that merge request
		currentLabels := mr.Labels
		for _, newLabel := range request.Params.Labels {
			if !Contains(currentLabels, newLabel) {
				currentLabels = append(currentLabels, newLabel)
			}
		}
		options := gitlab.UpdateMergeRequestOptions{
			Labels: &currentLabels,
		}
		_, res, err := api.MergeRequests.UpdateMergeRequest(mr.ProjectID, mr.IID, &options)
		if res.StatusCode != 200 {
			body, _ := ioutil.ReadAll(res.Body)
			log.Fatalf("Update merge request failed: %d, response %s", res.StatusCode, string(body))
		}
		if err != nil {
			common.Fatal("Update merge request failed", err)
		}
	}

	commentBody := request.Params.Comment.GetContent(os.Args[1])
	if commentBody != "" {
		message = message + fmt.Sprintf("New comment: %s \n", commentBody)
		options := gitlab.CreateMergeRequestNoteOptions{
			Body: &commentBody,
		}
		_, res, err := api.Notes.CreateMergeRequestNote(mr.ProjectID, mr.IID, &options)
		if res.StatusCode != 201 {
			body, _ := ioutil.ReadAll(res.Body)
			log.Fatalf("Add merge request comment failed: %d, response %s", res.StatusCode, string(body))
		}
		if err != nil {
			common.Fatal("Add merge request comment failed", err)
		}
	}

	response := out.Response{
		Version: resource.Version{
			ID:        mr.IID,
			UpdatedAt: mr.UpdatedAt,
		},
		Metadata: buildMetadata(&mr, message),
	}

	json.NewEncoder(os.Stdout).Encode(response)
}

func buildMetadata(mr *gitlab.MergeRequest, message string) resource.Metadata {

	return []resource.MetadataField{
		{
			Name:  "id",
			Value: strconv.Itoa(mr.ID),
		},
		{
			Name:  "iid",
			Value: strconv.Itoa(mr.IID),
		},
		{
			Name:  "sha",
			Value: mr.SHA,
		},
		{
			Name:  "message",
			Value: message,
		},
		{
			Name:  "title",
			Value: mr.Title,
		},
		{
			Name:  "author",
			Value: mr.Author.Name,
		},
		{
			Name:  "source",
			Value: mr.SourceBranch,
		},
		{
			Name:  "target",
			Value: mr.TargetBranch,
		},
		{
			Name:  "url",
			Value: mr.WebURL,
		},
	}
}

func Contains(sl []string, v string) bool {
	for _, vv := range sl {
		if vv == v {
			return true
		}
	}
	return false
}
