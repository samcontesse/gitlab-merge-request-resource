package main

import (
	"encoding/json"
	"github.com/samcontesse/gitlab-merge-request-resource"
	"github.com/samcontesse/gitlab-merge-request-resource/common"
	"github.com/samcontesse/gitlab-merge-request-resource/in"
	"github.com/xanzy/go-gitlab"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func main() {

	if len(os.Args) < 2 {
		println("usage: " + os.Args[0] + " <destination>")
		os.Exit(1)
	}

	destination := os.Args[1]

	var request in.Request

	if err := json.NewDecoder(os.Stdin).Decode(&request); err != nil {
		common.Fatal("reading request from stdin", err)
	}

	api := gitlab.NewClient(common.GetDefaultClient(request.Source.Insecure), request.Source.PrivateToken)
	api.SetBaseURL(request.Source.GetBaseURL())

	mr, _, err := api.MergeRequests.GetMergeRequest(request.Source.GetProjectPath(), request.Version.ID, &gitlab.GetMergeRequestsOptions{})

	if err != nil {
		common.Fatal("getting merge request", err)
	}

	mr.UpdatedAt = request.Version.UpdatedAt

	target := createRepositoryUrl(api, mr.TargetProjectID, request.Source.PrivateToken)
	source := createRepositoryUrl(api, mr.SourceProjectID, request.Source.PrivateToken)

	commit, _, err := api.Commits.GetCommit(mr.SourceProjectID, mr.SHA)
	if err != nil {
		common.Fatal("listing merge request commits", err)
	}

	execGitCommand([]string{"clone", "-c", "http.sslVerify=" + strconv.FormatBool(!request.Source.Insecure), "-o", "target", "-b", mr.TargetBranch, target.String(), destination})
	os.Chdir(destination)
	execGitCommand([]string{"remote", "add", "source", source.String()})
	execGitCommand([]string{"remote", "update"})
	execGitCommand([]string{"merge", "--no-ff", "--no-commit", mr.SHA})

	notes, _ := json.Marshal(mr)
	err = ioutil.WriteFile(".git/merge-request.json", notes, 0644)

	response := in.Response{Version: request.Version, Metadata: buildMetadata(mr, commit)}

	json.NewEncoder(os.Stdout).Encode(response)
}

func execGitCommand(args []string) {
	cmd := "git"
	command := exec.Command(cmd, args...)
	command.Stdin = os.Stdin
	command.Stderr = os.Stderr
	err := command.Run()
	if err != nil {
		common.Fatal("executing git "+strings.Join(args, " "), err)
	}
}

func createRepositoryUrl(api *gitlab.Client, pid int, token string) *url.URL {
	project, _, err := api.Projects.GetProject(pid, &gitlab.GetProjectOptions{})
	if err != nil {
		common.Fatal("reading project from api", err)
	}

	u, err := url.Parse(project.HTTPURLToRepo)
	if err != nil {
		common.Fatal("parsing repository url", err)
	}
	u.User = url.UserPassword("gitlab-ci-token", token)
	return u
}

func buildMetadata(mr *gitlab.MergeRequest, commit *gitlab.Commit) resource.Metadata {
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
