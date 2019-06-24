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
	"regexp"
	"fmt"
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

	paths := fmt.Sprintf(`^(%v)\/.*`, strings.Join(request.Source.Paths, "|"))
	ignorePaths := fmt.Sprintf(`^(%v)\/.*`, strings.Join(request.Source.IgnorePaths, "|"))
	
	for _, mr := range requests {

		commit, _, err := api.Commits.GetCommit(mr.ProjectID, mr.SHA)
		updatedAt := commit.CommittedDate

		if err != nil {
			continue
		}

		if strings.Contains(commit.Title, "[skip ci]") || strings.Contains(commit.Message, "[skip ci]") {
			continue
		}

		if !request.Source.SkipTriggerComment {
			notes, _, _ := api.Notes.ListMergeRequestNotes(mr.ProjectID, mr.IID, &gitlab.ListMergeRequestNotesOptions{})
			updatedAt = getMostRecentUpdateTime(notes, updatedAt)
		}

		if request.Source.SkipNotMergeable && mr.MergeStatus != "can_be_merged" {
			continue
		}

		if request.Source.SkipWorkInProgress && mr.WorkInProgress {
			continue
		}

		if request.Version.UpdatedAt != nil && !updatedAt.After(*request.Version.UpdatedAt) {
			continue
		}

		if len(request.Source.Paths) > 0 && !checkIncludePaths(api, paths, mr.ProjectID, mr.IID) {			
			continue
		}

		if len(request.Source.IgnorePaths) > 0 && checkIncludePaths(api, ignorePaths, mr.ProjectID, mr.IID) {
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


func checkIncludePaths(api *gitlab.Client, paths string, projectid int, mriid int) bool {

	versions, _, err := api.MergeRequests.GetMergeRequestDiffVersions(projectid, mriid, nil)
	if err != nil {
		common.Fatal("retrieving merge request diff versions", err)
	}
	
	diff, _, err := api.MergeRequests.GetSingleMergeRequestDiffVersion(projectid, mriid, versions[0].ID)	
	if err != nil {
		common.Fatal("retrieving merge request diff", err)
	}

	for _, d := range diff.Diffs {
		if match, _ := regexp.MatchString(paths, d.OldPath); match {
			return true
		}
		if match, _ := regexp.MatchString(paths, d.NewPath); match {
			return true
		}
	}
	return false
}


func getMostRecentUpdateTime(notes []*gitlab.Note, updatedAt *time.Time) *time.Time {
	for _, note := range notes {
		if strings.Contains(note.Body, "[trigger ci]") && updatedAt.Before(*note.UpdatedAt) {
			return note.UpdatedAt
		}
	}
	return updatedAt
}
