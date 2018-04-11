package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"github.com/samcontesse/gitlab-merge-request-resource"
	"github.com/samcontesse/gitlab-merge-request-resource/common"
	"github.com/samcontesse/gitlab-merge-request-resource/out"
	"github.com/xanzy/go-gitlab"
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

	revision := readRevision()

	var mr gitlab.MergeRequest
	unmarshallNotes("mr", revision, &mr)

	api := gitlab.NewClient(nil, request.Source.PrivateToken)
	api.SetBaseURL(request.Source.GetBaseURL())

	state := gitlab.BuildState(gitlab.BuildStateValue(request.Params.Status))
	target := resource.GetTargetURL()
	name := resource.GetPipelineName()

	options := gitlab.SetCommitStatusOptions{
		Name:      &name,
		TargetURL: &target,
		State:     *state,
	}

	api.Commits.SetCommitStatus(mr.ProjectID, revision, &options)

	response := out.Response{Version: resource.Version{
		ID:        mr.IID,
		UpdatedAt: mr.UpdatedAt,
	}}

	json.NewEncoder(os.Stdout).Encode(response)

}

func readRevision() string {
	var outbuf bytes.Buffer
	command := exec.Command("git", "rev-parse", "HEAD")
	command.Stdout = &outbuf
	if err := command.Run(); err != nil {
		common.Fatal("reading HEAD revision", err)
	}
	return strings.TrimSpace(outbuf.String())
}

func unmarshallNotes(ref string, revision string, v interface{}) {
	var outbuf bytes.Buffer
	command := exec.Command("git", "notes", "--ref="+ref, "show", revision)
	command.Stdout = &outbuf
	if err := command.Run(); err != nil {
		common.Fatal("reading build notes", err)
	}
	err := json.Unmarshal(outbuf.Bytes(), &v)
	if err != nil {
		common.Fatal("unmarshalling "+ref+" from "+revision+" notes", err)
	}

}
