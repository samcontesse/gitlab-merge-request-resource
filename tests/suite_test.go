package gitlab_merge_request_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var bins struct {
	In    string `json:"in"`
	Out   string `json:"out"`
	Check string `json:"check"`
}

var _ = SynchronizedBeforeSuite(
	func() []byte {
		var err error

		b := bins

		if _, err := os.Stat("/opt/resource/in"); err == nil {
			b.In = "/opt/resource/in"
		} else {
			b.In, err = gexec.Build("github.com/samcontesse/gitlab-merge-request-resource/in/cmd")
			Expect(err).ToNot(HaveOccurred())
		}

		if _, err := os.Stat("/opt/resource/out"); err == nil {
			b.Out = "/opt/resource/out"
		} else {
			b.Out, err = gexec.Build("github.com/samcontesse/gitlab-merge-request-resource/out/cmd")
			Expect(err).ToNot(HaveOccurred())
		}

		if _, err := os.Stat("/opt/resource/check"); err == nil {
			b.Check = "/opt/resource/check"
		} else {
			b.Check, err = gexec.Build("github.com/samcontesse/gitlab-merge-request-resource/check/cmd")
			Expect(err).ToNot(HaveOccurred())
		}

		j, err := json.Marshal(b)
		Expect(err).ToNot(HaveOccurred())

		return j
	},
	func(bp []byte) {
		err := json.Unmarshal(bp, &bins)
		Expect(err).ToNot(HaveOccurred())
	})

var _ = SynchronizedAfterSuite(
	func() {
		// TODO
	}, func() {
		gexec.CleanupBuildArtifacts()
	})

func TestGitLabMergeRequestResource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GitlabResource Suite")
}

func Fixture(filename string) string {
	path := filepath.Join("fixtures", filename)
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	return string(contents)
}
