package main

import (
	"encoding/json"
	"github.com/samcontesse/gitlab-merge-request-resource/pkg"
	"github.com/samcontesse/gitlab-merge-request-resource/pkg/check"
	"github.com/xanzy/go-gitlab"
	"os"
)

func main() {

	var request check.Request
	inputRequest(&request)

	client, err := gitlab.NewClient(request.Source.PrivateToken, gitlab.WithHTTPClient(pkg.GetDefaultClient(request.Source.Insecure)), gitlab.WithBaseURL(request.Source.GetBaseURL()))
	if err != nil {
		pkg.Fatal("initializing gitlab client", err)
	}

	command := check.NewCommand(client)
	response, err := command.Run(request)
	if err != nil {
		pkg.Fatal("running command", err)
	}

	outputResponse(response)
}

func inputRequest(request *check.Request) {
	if err := json.NewDecoder(os.Stdin).Decode(request); err != nil {
		pkg.Fatal("reading request from stdin", err)
	}
}

func outputResponse(response check.Response) {
	if err := json.NewEncoder(os.Stdout).Encode(response); err != nil {
		pkg.Fatal("writing response to stdout", err)
	}
}
