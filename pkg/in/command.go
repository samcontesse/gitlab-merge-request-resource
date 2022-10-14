package in

import (
	"encoding/json"
	"github.com/samcontesse/gitlab-merge-request-resource/pkg"
	"github.com/xanzy/go-gitlab"
	"io/ioutil"
	"net/url"
	"os"
	"strconv"
)

type Command struct {
	client *gitlab.Client
	runner GitRunner
}

func NewCommand(client *gitlab.Client) *Command {
	return &Command{
		client,
		NewRunner(),
	}
}

func (command *Command) WithRunner(runner GitRunner) *Command {
	command.runner = runner
	return command
}

func (command *Command) Run(destination string, request Request) (Response, error) {
	err := os.MkdirAll(destination, 0755)
	if err != nil {
		return Response{}, err
	}

	user, _, err := command.client.Users.CurrentUser()
	err = command.runner.Run("config", "--global", "user.email", user.Email)
	if err != nil {
		return Response{}, err
	}

	err = command.runner.Run("config", "--global", "user.name", user.Name)
	if err != nil {
		return Response{}, err
	}

	mr, _, err := command.client.MergeRequests.GetMergeRequest(request.Source.GetProjectPath(), request.Version.ID, &gitlab.GetMergeRequestsOptions{})
	if err != nil {
		return Response{}, err
	}

	mr.UpdatedAt = request.Version.UpdatedAt

	target, err := command.createRepositoryUrl(mr.TargetProjectID, request.Source.PrivateToken)
	if err != nil {
		return Response{}, err
	}
	source, err := command.createRepositoryUrl(mr.SourceProjectID, request.Source.PrivateToken)
	if err != nil {
		return Response{}, err
	}

	commit, _, err := command.client.Commits.GetCommit(mr.SourceProjectID, mr.SHA)
	if err != nil {
		return Response{}, err
	}

	err = command.runner.Run("clone", "-c", "http.sslVerify="+strconv.FormatBool(!request.Source.Insecure), "-o", "target", "-b", mr.TargetBranch, target.String(), destination)
	if err != nil {
		return Response{}, err
	}

	os.Chdir(destination)

	err = command.runner.Run("remote", "add", "source", source.String())
	if err != nil {
		return Response{}, err
	}

	err = command.runner.Run("remote", "update")
	if err != nil {
		return Response{}, err
	}

	err = command.runner.Run("merge", "--no-ff", "--no-commit", mr.SHA)
	if err != nil {
		return Response{}, err
	}

	if request.Source.Recursive {
		err = command.runner.Run("submodule", "update", "--init", "--recursive")
		if err != nil {
			return Response{}, err
		}
	}

	notes, _ := json.Marshal(mr)
	err = ioutil.WriteFile(".git/merge-request.json", notes, 0644)
	if err != nil {
		return Response{}, err
	}

	err = ioutil.WriteFile(".git/merge-request-source-branch", []byte(mr.SourceBranch), 0644)
	if err != nil {
		return Response{}, err
	}

	response := Response{Version: request.Version, Metadata: buildMetadata(mr, commit)}

	return response, nil
}

func (command *Command) createRepositoryUrl(pid int, token string) (*url.URL, error) {
	project, _, err := command.client.Projects.GetProject(pid, &gitlab.GetProjectOptions{})
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(project.HTTPURLToRepo)
	if err != nil {
		return nil, err
	}

	u.User = url.UserPassword("gitlab-ci-token", token)

	return u, nil
}

func buildMetadata(mr *gitlab.MergeRequest, commit *gitlab.Commit) pkg.Metadata {
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
			Name:  "message",
			Value: commit.Title,
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
