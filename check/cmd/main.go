package main

import (
	"encoding/json"
	"github.com/samcontesse/gitlab-merge-request-resource"
	"github.com/samcontesse/gitlab-merge-request-resource/check"
	"github.com/samcontesse/gitlab-merge-request-resource/common"
	"github.com/xanzy/go-gitlab"
	"os"
	"strings"
	"time"
)

func main() {

	var request check.Request

	if err := json.NewDecoder(os.Stdin).Decode(&request); err != nil {
		common.Fatal("reading request from stdin", err)
	}

	api := gitlab.NewClient(common.GetDefaultClient(request.Source.Insecure), request.Source.PrivateToken)
	api.SetBaseURL(request.Source.GetBaseURL())

	options := &gitlab.ListProjectMergeRequestsOptions{
		State:   gitlab.String("opened"),
		OrderBy: gitlab.String("updated_at"),
		Labels:  request.Source.Labels}
	requests, _, err := api.MergeRequests.ListProjectMergeRequests(request.Source.GetProjectPath(), options)

	if err != nil {
		common.Fatal("retrieving opened merge requests", err)
	}

	var versions []resource.Version
	versions = make([]resource.Version, 0)

	for _, mr := range requests {

		commit, _, err := api.Commits.GetCommit(mr.ProjectID, mr.SHA)
		updatedAt := commit.CommittedDate

		if err != nil {
			continue
		}

		if strings.Contains(commit.Title, "[skip ci]") || strings.Contains(commit.Message, "[skip ci]") {
			continue
		}

		if request.Source.SkipNotMergeable && mr.MergeStatus != "can_be_merged" {
			continue
		}

		if !request.Source.SkipTriggerComment {
			notes, _, _ := api.Notes.ListMergeRequestNotes(mr.ProjectID, mr.IID, &gitlab.ListMergeRequestNotesOptions{})
			updatedAt = getMostRecentUpdateTime(request.Source.OnlyTriggerComment, notes, updatedAt)
		}

		if len(strings.TrimSpace(request.Source.OnlyTriggerComment)) == 0 &&
				request.Source.SkipWorkInProgress && mr.WorkInProgress {
			continue
		}

		if request.Version.UpdatedAt != nil && !updatedAt.After(*request.Version.UpdatedAt) {
			continue
		}

		target := request.Source.GetTargetURL()
		name := resource.GetPipelineName()

		options := gitlab.SetCommitStatusOptions{
			Name:      &name,
			TargetURL: &target,
			State:     gitlab.Pending,
		}

		api.Commits.SetCommitStatus(mr.SourceProjectID, mr.SHA, &options)

		versions = append(versions, resource.Version{ID: mr.IID, UpdatedAt: updatedAt})

	}

	json.NewEncoder(os.Stdout).Encode(versions)

}

func getMostRecentUpdateTime(triggerComment string, notes []*gitlab.Note, updatedAt *time.Time) *time.Time {
	for _, note := range notes {
		if strings.Contains(note.Body, triggerComment) && updatedAt.Before(*note.UpdatedAt) {
			return note.UpdatedAt
		}
	}
	return updatedAt
}
