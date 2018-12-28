package gitlab_merge_request_test

import (
	"bytes"
	"encoding/json"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"

	"github.com/samcontesse/gitlab-merge-request-resource"
	"github.com/samcontesse/gitlab-merge-request-resource/out"
	"github.com/xanzy/go-gitlab"
	"log"
	"net/url"
	"path"
)

var (
	GitLabClient *gitlab.Client
	mux          *http.ServeMux
	server       *httptest.Server
)

var _ = Describe("Out", func() {
	var (
		srcDir         string
		req            out.Request
		res            out.Response
		localGitlabURL string
	)

	BeforeEach(func() {
		var err error
		srcDir, err = ioutil.TempDir("", "gitlab-merge-request-resource-dir")
		Expect(err).ToNot(HaveOccurred())
		_ = os.Mkdir(path.Join(srcDir, "repo"), 0744)
		_ = os.Mkdir(path.Join(srcDir, "repo", ".git"), 0744)
		_ = ioutil.WriteFile(
			path.Join(srcDir, "repo", ".git", "merge-request.json"),
			[]byte(Fixture("merge-request.json")),
			0744)

		localGitlabURL = setupLocalGitlab("/api/v4")
	})

	AfterEach(func() {
		Expect(os.RemoveAll(srcDir)).To(Succeed())
		teardown()
	})

	JustBeforeEach(func() {
		cmd := exec.Command(bins.Out, srcDir)
		payload, err := json.Marshal(req)
		Expect(err).ToNot(HaveOccurred())

		outBuf := new(bytes.Buffer)
		cmd.Stdin = bytes.NewBuffer(payload)
		cmd.Stdout = outBuf
		cmd.Stderr = GinkgoWriter

		err = cmd.Run()
		Expect(err).ToNot(HaveOccurred())

		err = json.Unmarshal(outBuf.Bytes(), &res)
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("Only update Status", func() {

		BeforeEach(func() {
			req = out.Request{
				Source: resource.Source{
					URI:          localGitlabURL,
					PrivateToken: "$$random",
					Insecure:     false,
				},
				Params: out.Params{
					Repository: "repo",
					Status:     "running",
				},
			}

			// mock gitlab request for commit status base on fixtures/merge-request.json
			mux.HandleFunc("/api/v4/projects/1/statuses/abc", func(w http.ResponseWriter, r *http.Request) {
				commitStatus := gitlab.CommitStatus{ID: 1, SHA: "abc"}
				output, _ := json.Marshal(commitStatus)
				w.Header().Set("content-type", "application/json")
				w.WriteHeader(http.StatusCreated)
				w.Write(output)
			})
		})

		It("check Version.ID", func() {
			// res.Version.ID should be equal with IID in merge-request.json
			Expect(res.Version.ID).To(Equal(12))
		})

	})

	Describe("Only update Labels", func() {

		BeforeEach(func() {
			req = out.Request{
				Source: resource.Source{
					URI:          localGitlabURL,
					PrivateToken: "$$random",
					Insecure:     false,
				},
				Params: out.Params{
					Repository: "repo",
					Labels:     []string{"in-stage", "ut-pass"},
				},
			}

			// mock gitlab request for update merge request base on fixtures/merge-request.json
			mux.HandleFunc("/api/v4/projects/1/merge_requests/12", func(w http.ResponseWriter, r *http.Request) {
				updatedMR := gitlab.MergeRequest{
					ID:           1,
					IID:          2,
					TargetBranch: "master",
					SourceBranch: "dev",
				}
				output, _ := json.Marshal(updatedMR)
				w.Header().Set("content-type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write(output)
			})
		})

		It("check Version.ID", func() {
			// res.Version.ID should be equal with IID in merge-request.json
			Expect(res.Version.ID).To(Equal(12))
		})

	})

	Describe("Only both Status and Labels", func() {

		BeforeEach(func() {
			req = out.Request{
				Source: resource.Source{
					URI:          localGitlabURL,
					PrivateToken: "$$random",
					Insecure:     false,
				},
				Params: out.Params{
					Repository: "repo",
					Status:     "running",
					Labels:     []string{"in-stage", "ut-pass"},
				},
			}

			// mock gitlab request for commit status base on fixtures/merge-request.json
			mux.HandleFunc("/api/v4/projects/1/statuses/abc", func(w http.ResponseWriter, r *http.Request) {
				commitStatus := gitlab.CommitStatus{ID: 1, SHA: "abc"}
				output, _ := json.Marshal(commitStatus)
				w.Header().Set("content-type", "application/json")
				w.WriteHeader(http.StatusCreated)
				w.Write(output)
			})

			// mock gitlab request for update merge request base on fixtures/merge-request.json
			mux.HandleFunc("/api/v4/projects/1/merge_requests/12", func(w http.ResponseWriter, r *http.Request) {
				updatedMR := gitlab.MergeRequest{
					ID:           1,
					IID:          2,
					TargetBranch: "master",
					SourceBranch: "dev",
				}
				output, _ := json.Marshal(updatedMR)
				w.Header().Set("content-type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write(output)
			})
		})

		It("check Version.ID", func() {
			// res.Version.ID should be equal with IID in merge-request.json
			Expect(res.Version.ID).To(Equal(12))
		})

	})

})

func setupLocalGitlab(versionPrefix string) string {
	os.Setenv("GITLAB_TOKEN", "$$$randome")

	// test server
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)
	base, _ := url.Parse(server.URL)

	u, err := url.Parse(versionPrefix)
	if err != nil {
		log.Fatal(err)
	}
	url := base.ResolveReference(u)

	// github client configured to use test server
	GitLabClient = gitlab.NewClient(nil, "")
	GitLabClient.SetBaseURL(url.String())
	return url.String()
}

func teardown() {
	os.Unsetenv("GITLAB_TOKEN")
	server.Close()
}
