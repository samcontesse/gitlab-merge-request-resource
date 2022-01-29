package check

import (
	"github.com/samcontesse/gitlab-merge-request-resource/pkg"
	"github.com/xanzy/go-gitlab"
	"strings"
	"time"
)

type Command struct {
	client *gitlab.Client
}

func NewCommand(client *gitlab.Client) *Command {
	return &Command{
		client: client,
	}
}

func (command *Command) Run(request Request) (Response, error) {
	labels := gitlab.Labels(request.Source.Labels)

	// https://docs.gitlab.com/ee/api/pipelines.html#list-project-pipelines
	sort, err := request.Source.GetSort()
	if err != nil {
		return Response{}, err
	}

	options := &gitlab.ListProjectMergeRequestsOptions{
		State:        gitlab.String("opened"),
		OrderBy:      gitlab.String("updated_at"),
		Sort:         gitlab.String(sort),
		Labels:       &labels,
		TargetBranch: gitlab.String(request.Source.TargetBranch),
	}
	requests, _, err := command.client.MergeRequests.ListProjectMergeRequests(request.Source.GetProjectPath(), options)
	if err != nil {
		return Response{}, err
	}

	versions := make([]pkg.Version, 0)

	for _, mr := range requests {

		commit, _, err := command.client.Commits.GetCommit(mr.ProjectID, mr.SHA)
		if err != nil {
			return Response{}, err
		}
		updatedAt := commit.CommittedDate

		if err != nil {
			continue
		}

		if strings.Contains(commit.Title, "[skip ci]") || strings.Contains(commit.Message, "[skip ci]") {
			continue
		}

		if !request.Source.SkipTriggerComment {
			notes, _, _ := command.client.Notes.ListMergeRequestNotes(mr.ProjectID, mr.IID, &gitlab.ListMergeRequestNotesOptions{})
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

		target := request.Source.GetTargetURL()
		name := request.Source.GetPipelineName()

		options := gitlab.SetCommitStatusOptions{
			Name:      &name,
			TargetURL: &target,
			State:     gitlab.Pending,
		}

		_, _, _ = command.client.Commits.SetCommitStatus(mr.SourceProjectID, mr.SHA, &options)

		versions = append(versions, pkg.Version{ID: mr.IID, UpdatedAt: updatedAt})

	}

	return versions, nil
}

func getMostRecentUpdateTime(notes []*gitlab.Note, updatedAt *time.Time) *time.Time {
	for _, note := range notes {
		if strings.Contains(note.Body, "[trigger ci]") && updatedAt.Before(*note.UpdatedAt) {
			return note.UpdatedAt
		}
	}
	return updatedAt
}
