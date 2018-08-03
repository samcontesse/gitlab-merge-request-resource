package common

import (
	"os"
	"fmt"
	"net/http"
	"crypto/tls"
	"github.com/xanzy/go-gitlab"
	"github.com/samcontesse/gitlab-merge-request-resource"
)

func Fatal(doing string, err error) {
	println("error " + doing + ": " + err.Error())
	os.Exit(1)
}

func Sayf(message string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, message, args...)
}

func GetDefaultClient(insecure bool) *http.Client {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: insecure}
	return http.DefaultClient
}

func UpdateCommitStatus(mr *gitlab.MergeRequest, source resource.Source, state gitlab.BuildStateValue) {
	api := gitlab.NewClient(GetDefaultClient(source.Insecure), source.PrivateToken)
	api.SetBaseURL(source.GetBaseURL())

	target := source.GetTargetURL()
	name := resource.GetPipelineName()

	options := gitlab.SetCommitStatusOptions{
		Name:      &name,
		TargetURL: &target,
		State:     state,
	}

	api.Commits.SetCommitStatus(mr.ProjectID, mr.SHA, &options)
}

