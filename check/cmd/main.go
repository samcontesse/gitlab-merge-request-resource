package main

import (
	"encoding/json"
	"os"
	"github.com/xanzy/go-gitlab"
	"github.com/samcontesse/gitlab-merge-request-resource"
	"github.com/samcontesse/gitlab-merge-request-resource/check"
	"github.com/samcontesse/gitlab-merge-request-resource/common"
	"strings"
)

func main() {

	var request check.Request

	if err := json.NewDecoder(os.Stdin).Decode(&request); err != nil {
		common.Fatal("reading request from stdin", err)
	}

	api := gitlab.NewClient(nil, request.Source.PrivateToken)
	api.SetBaseURL(request.Source.GetBaseURL())

	options := &gitlab.ListMergeRequestsOptions{State: gitlab.String("opened"), OrderBy: gitlab.String("updated_at")}
	requests, _, err := api.MergeRequests.ListMergeRequests(request.Source.GetProjectPath(), options)

	if err != nil {
		common.Fatal("retrieving opened merge requests", err)
	}

	versions := []resource.Version{}

	for _, mr := range requests {

		commit, _, err := api.Commits.GetCommit(mr.ProjectID, mr.SHA)

		if err != nil {
			continue
		}

		if strings.Contains(commit.Title, "[skip ci]") || strings.Contains(commit.Message, "[skip ci]") {
			continue
		}

		mr.UpdatedAt = commit.CommittedDate

		if !request.Source.SkipTriggerComment {
			notes, _, _ := api.Notes.ListMergeRequestNotes(mr.ProjectID, mr.ID)
			refreshUpdatedAt(notes, mr)
		}

		if request.Source.SkipNotMergeable && mr.MergeStatus != "can_be_merged" {
			continue
		}

		if request.Source.SkipWorkInProgress && mr.WorkInProgress {
			continue
		}

		if !mr.UpdatedAt.After(request.Version.UpdatedAt) {
			continue
		}

		versions = append(versions, resource.Version{ID: mr.ID, UpdatedAt: *mr.UpdatedAt})

	}

	json.NewEncoder(os.Stdout).Encode(versions)

}

func refreshUpdatedAt(notes []*gitlab.Note, mr *gitlab.MergeRequest) {
	for _, note := range notes {
		if strings.Contains(note.Body, "[trigger ci]") && mr.UpdatedAt.Before(*note.UpdatedAt) {
			mr.UpdatedAt = note.UpdatedAt
		}
	}
}
