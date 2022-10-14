package out

import (
	"encoding/json"
	"github.com/samcontesse/gitlab-merge-request-resource/pkg"
	"github.com/xanzy/go-gitlab"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Command struct {
	client *gitlab.Client
}

func NewCommand(client *gitlab.Client) *Command {
	return &Command{client}
}

func (command *Command) Run(destination string, request Request) (Response, error) {
	repo := filepath.Join(destination, request.Params.Repository)
	err := os.MkdirAll(repo, 0755)
	if err != nil {
		return Response{}, err
	}

	err = os.Chdir(repo)
	if err != nil {
		return Response{}, err
	}

	file, err := os.ReadFile(".git/merge-request.json")
	if err != nil {
		return Response{}, err
	}

	var mr gitlab.MergeRequest
	err = json.Unmarshal(file, &mr)
	if err != nil {
		return Response{}, err
	}

	err = command.updateCommitStatus(request, mr)
	if err != nil {
		return Response{}, err
	}

	err = command.updateLabels(request, mr)
	if err != nil {
		return Response{}, err
	}

	err = command.createNote(destination, request, mr)
	if err != nil {
		return Response{}, err
	}

	response := Response{
		Version: pkg.Version{
			ID:        mr.IID,
			UpdatedAt: mr.UpdatedAt,
		},
		Metadata: buildMetadata(&mr),
	}

	return response, nil
}

func (command *Command) createNote(destination string, request Request, mr gitlab.MergeRequest) error {
	body, err := request.Params.Comment.ReadContent(destination)
	if err != nil {
		return err
	}

	if body != "" {
		options := gitlab.CreateMergeRequestNoteOptions{Body: &body}
		_, _, err := command.client.Notes.CreateMergeRequestNote(mr.SourceProjectID, mr.IID, &options)
		if err != nil {
			return err
		}
	}
	return nil
}

func (command *Command) updateLabels(request Request, mr gitlab.MergeRequest) error {
	if request.Params.Labels != nil {

		// exclude empty string when there is no tags
		// this should be fixed in go-gitlab
		if len(mr.Labels) == 1 && mr.Labels[0] == "" {
			mr.Labels = mr.Labels[1:]
		}

		labels := append(mr.Labels, request.Params.Labels...)
		options := gitlab.UpdateMergeRequestOptions{Labels: &labels}

		result, _, err := command.client.MergeRequests.UpdateMergeRequest(mr.SourceProjectID, mr.IID, &options)
		if err != nil {
			return err
		}

		mr = *result
	}
	return nil
}

func (command *Command) updateCommitStatus(request Request, mr gitlab.MergeRequest) error {
	if request.Params.Status != "" {
		state := gitlab.BuildState(gitlab.BuildStateValue(request.Params.Status))
		target := request.Source.GetTargetURL()
		name := request.Source.GetPipelineName()
		options := gitlab.SetCommitStatusOptions{
			Name:      &name,
			TargetURL: &target,
			State:     *state,
		}

		_, _, err := command.client.Commits.SetCommitStatus(mr.SourceProjectID, mr.SHA, &options)
		if err != nil {
			return err
		}
	}
	return nil
}

func buildMetadata(mr *gitlab.MergeRequest) pkg.Metadata {

	return []pkg.MetadataField{
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
		{
			Name:  "labels",
			Value: strings.Join(mr.Labels, ","),
		},
	}
}
