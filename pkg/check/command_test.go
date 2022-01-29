package check_test

import (
	"encoding/json"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samcontesse/gitlab-merge-request-resource/pkg"
	"github.com/samcontesse/gitlab-merge-request-resource/pkg/check"
	"github.com/xanzy/go-gitlab"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"time"
)

var _ = Describe("Check", func() {

	var (
		t       time.Time
		mux     *http.ServeMux
		command *check.Command
		root    *url.URL
	)

	BeforeEach(func() {
		t, _ = time.Parse(time.RFC3339, "2022-01-01T08:00:00Z")
		mux = http.NewServeMux()
		server := httptest.NewServer(mux)
		root, _ = url.Parse(server.URL)
		context, _ := url.Parse("/api/v4")
		base := root.ResolveReference(context)
		client, _ := gitlab.NewClient("$", gitlab.WithBaseURL(base.String()))

		_ = os.Setenv("ATC_EXTERNAL_URL", "https://concourse-ci.company.ltd")
		_ = os.Setenv("BUILD_TEAM_NAME", "winner")
		_ = os.Setenv("BUILD_PIPELINE_NAME", "baltic")
		_ = os.Setenv("BUILD_JOB_NAME", "release")
		_ = os.Setenv("BUILD_NAME", "1")

		command = check.NewCommand(client)
	})

	Describe("Run", func() {

		Context("When it has a minimal valid configuration", func() {

			BeforeEach(func() {
				mux.HandleFunc("/api/v4/projects/namespace/project/merge_requests", func(w http.ResponseWriter, r *http.Request) {
					mr := gitlab.MergeRequest{IID: 88, ID: 99, SHA: "abc", ProjectID: 42}
					output, _ := json.Marshal([]gitlab.MergeRequest{mr})
					w.Header().Set("content-type", "application/json")
					w.WriteHeader(http.StatusOK)
					w.Write(output)
				})

				mux.HandleFunc("/api/v4/projects/42/repository/commits/abc", func(w http.ResponseWriter, r *http.Request) {
					commit := gitlab.Commit{CommittedDate: &t}
					output, _ := json.Marshal(commit)
					w.Header().Set("content-type", "application/json")
					w.WriteHeader(http.StatusOK)
					w.Write(output)
				})
			})

			It("Should return a single version", func() {

				project, _ := url.Parse("namespace/project.git")
				uri := root.ResolveReference(project)

				request := check.Request{
					Source: pkg.Source{
						URI:          uri.String(),
						PrivateToken: "$",
					},
				}

				response, err := command.Run(request)
				Expect(err).Should(BeNil())
				Expect(len(response)).To(Equal(1))
				Expect(response[0].ID).To(Equal(88))
				Expect(response[0].UpdatedAt).To(Equal(&t))
			})

		})

		Context("When it contains an invalid project uri", func() {

			BeforeEach(func() {
				mux.HandleFunc("/api/v4/projects/namespace/project/merge_requests", http.NotFound)
			})

			It("Should error when project uri is invalid", func() {

				project, _ := url.Parse("invalid/project.git")
				uri := root.ResolveReference(project)

				request := check.Request{
					Source: pkg.Source{
						URI:          uri.String(),
						PrivateToken: "$",
					},
				}

				response, err := command.Run(request)
				Expect(response).To(Equal(check.Response{}))
				Expect(err).NotTo(BeNil())
			})

		})

	})

})
