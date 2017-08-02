package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"strconv"
	"github.com/samcontesse/gitlab-merge-request-resource"
	"github.com/samcontesse/gitlab-merge-request-resource/common"
	"github.com/samcontesse/gitlab-merge-request-resource/in"
	"github.com/xanzy/go-gitlab"
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

	api := gitlab.NewClient(nil, request.Source.PrivateToken)
	api.SetBaseURL(request.Source.GetBaseURL())

	mr, _, err := api.MergeRequests.GetMergeRequest(request.Source.GetProjectPath(), request.Version.ID)
	commit, _, err := api.Commits.GetCommit(mr.ProjectID, mr.SHA)

	if err != nil {
		common.Fatal("getting merge request", err)
	}

	cmd := "git"
	args := []string{"clone", "-b", mr.SourceBranch, "--single-branch", request.Source.GetCloneURL(), destination}
	command := exec.Command(cmd, args...)
	command.Stdin = os.Stdin
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		common.Fatal("cloning repository", err)
	}

	os.Chdir(destination)

	args = []string{"reset", "--hard", mr.SHA}
	if err := exec.Command(cmd, args...).Run(); err != nil {
		common.Fatal("resetting HEAD to "+mr.SHA, err)
	}

	addCommitNotes(mr, "mr");
	response := in.Response{Version: request.Version, Metadata: buildMetadata(mr, commit)}

	json.NewEncoder(os.Stdout).Encode(response)
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

func addCommitNotes(object interface{}, ref string) {
	notes, err := json.Marshal(object)
	if err != nil {
		common.Fatal("marshalling "+ref+" notes", err)
	}

	cmd := "git"
	args := []string{"notes", "--ref=" + ref, "add", "-f", "-m", string(notes) }
	if err := exec.Command(cmd, args...).Run(); err != nil {
		common.Fatal("adding notes `"+ref+"` to commit", err)
	}
}
