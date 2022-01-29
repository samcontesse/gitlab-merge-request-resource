package out_test

import (
	"encoding/json"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samcontesse/gitlab-merge-request-resource/pkg"
	"github.com/samcontesse/gitlab-merge-request-resource/pkg/out"
	"github.com/xanzy/go-gitlab"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
)

var _ = Describe("Out", func() {
	var (
		mux         *http.ServeMux
		server      *httptest.Server
		root        *url.URL
		command     *out.Command
		destination string
	)

	BeforeEach(func() {

		mux = http.NewServeMux()
		server = httptest.NewServer(mux)
		root, _ = url.Parse(server.URL)
		context, _ := url.Parse("/api/v4")
		base := root.ResolveReference(context)
		client, _ := gitlab.NewClient("$", gitlab.WithBaseURL(base.String()))
		destination, _ = ioutil.TempDir("", "gitlab-merge-request-resource-out")

		_ = os.Setenv("ATC_EXTERNAL_URL", "https://concourse-ci.company.ltd")
		_ = os.Setenv("BUILD_TEAM_NAME", "winner")
		_ = os.Setenv("BUILD_PIPELINE_NAME", "baltic")
		_ = os.Setenv("BUILD_JOB_NAME", "release")
		_ = os.Setenv("BUILD_NAME", "1")

		command = out.NewCommand(client)
	})

	AfterEach(func() {
		os.Remove(destination)
		server.Close()
	})

	Describe("Only update status", func() {

		BeforeEach(func() {
			mr := gitlab.MergeRequest{ID: 1, IID: 42, SHA: "abc", SourceProjectID: 1, Author: &gitlab.BasicUser{Name: "john"}}
			content, _ := json.Marshal(mr)

			_ = os.Mkdir(path.Join(destination, "repo"), 0755)
			_ = os.Mkdir(path.Join(destination, "repo", ".git"), 0755)
			_ = ioutil.WriteFile(path.Join(destination, "repo", ".git", "merge-request.json"), content, 0644)
		})

		It("Sets the commit status", func() {
			project, _ := url.Parse("namespace/project.git")
			uri := root.ResolveReference(project)

			request := out.Request{
				Source: pkg.Source{URI: uri.String()},
				Params: out.Params{
					Repository: "repo",
					Status:     "running",
				},
			}

			mux.HandleFunc("/api/v4/projects/1/statuses/abc", func(w http.ResponseWriter, r *http.Request) {
				body, _ := ioutil.ReadAll(r.Body)
				Expect(string(body)).To(ContainSubstring(`"state":"running"`))
				status := gitlab.CommitStatus{ID: 1, SHA: "abc"}
				output, _ := json.Marshal(status)
				w.Header().Set("content-type", "application/json")
				w.WriteHeader(http.StatusCreated)
				w.Write(output)
			})

			response, err := command.Run(destination, request)
			Expect(err).Should(BeNil())
			Expect(response.Version.ID).To(Equal(42))
		})

	})

	Describe("Only update labels", func() {

		BeforeEach(func() {
			mux.HandleFunc("/api/v4/projects/1/merge_requests/42", func(w http.ResponseWriter, r *http.Request) {
				mr := gitlab.MergeRequest{ID: 1, IID: 42, SHA: "abc", SourceProjectID: 1, Author: &gitlab.BasicUser{Name: "john"}}
				output, _ := json.Marshal(mr)
				w.Header().Set("content-type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write(output)
			})

			mr := gitlab.MergeRequest{ID: 1, IID: 42, SourceProjectID: 1, Labels: []string{"existing-label"}, Author: &gitlab.BasicUser{Name: "john"}}
			content, _ := json.Marshal(mr)

			_ = os.Mkdir(path.Join(destination, "repo"), 0755)
			_ = os.Mkdir(path.Join(destination, "repo", ".git"), 0755)
			_ = ioutil.WriteFile(path.Join(destination, "repo", ".git", "merge-request.json"), content, 0644)
		})

		It("check Version.ID", func() {
			project, _ := url.Parse("namespace/project.git")
			uri := root.ResolveReference(project)

			request := out.Request{
				Source: pkg.Source{URI: uri.String()},
				Params: out.Params{
					Repository: "repo",
					Labels:     []string{"in-stage", "ut-pass"},
				},
			}

			response, err := command.Run(destination, request)
			Expect(err).Should(BeNil())
			Expect(response.Version.ID).To(Equal(42))
		})

	})

	Describe("Both status and labels", func() {

		BeforeEach(func() {
			mux.HandleFunc("/api/v4/projects/1/merge_requests/42", func(w http.ResponseWriter, r *http.Request) {
				mr := gitlab.MergeRequest{ID: 1, IID: 42, SHA: "abc", SourceProjectID: 1, Author: &gitlab.BasicUser{Name: "john"}}
				output, _ := json.Marshal(mr)
				w.Header().Set("content-type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write(output)
			})

			mr := gitlab.MergeRequest{ID: 1, IID: 42, SourceProjectID: 1, SHA: "abc", Labels: []string{"existing-label"}, Author: &gitlab.BasicUser{Name: "john"}}
			content, _ := json.Marshal(mr)

			_ = os.Mkdir(path.Join(destination, "repo"), 0755)
			_ = os.Mkdir(path.Join(destination, "repo", ".git"), 0755)
			_ = ioutil.WriteFile(path.Join(destination, "repo", ".git", "merge-request.json"), content, 0644)
		})

		It("check Version.ID", func() {
			project, _ := url.Parse("namespace/project.git")
			uri := root.ResolveReference(project)

			request := out.Request{
				Source: pkg.Source{URI: uri.String()},
				Params: out.Params{
					Repository: "repo",
					Status:     "running",
					Labels:     []string{"in-stage", "ut-pass"},
				},
			}

			mux.HandleFunc("/api/v4/projects/1/statuses/abc", func(w http.ResponseWriter, r *http.Request) {
				body, _ := ioutil.ReadAll(r.Body)
				Expect(string(body)).To(ContainSubstring(`"state":"running"`))
				status := gitlab.CommitStatus{ID: 1, SHA: "abc"}
				output, _ := json.Marshal(status)
				w.Header().Set("content-type", "application/json")
				w.WriteHeader(http.StatusCreated)
				w.Write(output)
			})

			response, err := command.Run(destination, request)
			Expect(err).Should(BeNil())
			Expect(response.Version.ID).To(Equal(42))
		})

	})

	Describe("Only add comment", func() {

		BeforeEach(func() {
			mux.HandleFunc("/api/v4/projects/1/merge_requests/42", func(w http.ResponseWriter, r *http.Request) {
				mr := gitlab.MergeRequest{ID: 1, IID: 42, SHA: "abc", SourceProjectID: 1, Author: &gitlab.BasicUser{Name: "john"}}
				output, _ := json.Marshal(mr)
				w.Header().Set("content-type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write(output)
			})

			mr := gitlab.MergeRequest{ID: 1, IID: 42, SourceProjectID: 1, SHA: "abc", Labels: []string{"existing-label"}, Author: &gitlab.BasicUser{Name: "john"}}
			content, _ := json.Marshal(mr)

			_ = os.Mkdir(path.Join(destination, "repo"), 0755)
			_ = os.Mkdir(path.Join(destination, "repo", ".git"), 0755)
			_ = ioutil.WriteFile(path.Join(destination, "repo", ".git", "merge-request.json"), content, 0644)
			_ = ioutil.WriteFile(path.Join(destination, "comment.txt"), []byte("lorem ipsum"), 0644)
		})

		It("check Version.ID", func() {

			project, _ := url.Parse("namespace/project.git")
			uri := root.ResolveReference(project)

			request := out.Request{
				Source: pkg.Source{URI: uri.String()},
				Params: out.Params{
					Repository: "repo",
					Comment:    out.Comment{FilePath: "comment.txt", Text: "new comment, $FILE_CONTENT"},
				},
			}

			mux.HandleFunc("/api/v4/projects/1/merge_requests/42/notes", func(w http.ResponseWriter, r *http.Request) {
				n := gitlab.Note{ID: 1}
				output, _ := json.Marshal(n)
				w.Header().Set("content-type", "application/json")
				w.WriteHeader(http.StatusCreated)
				w.Write(output)
			})

			response, err := command.Run(destination, request)
			Expect(err).Should(BeNil())
			Expect(response.Version.ID).To(Equal(42))
		})

	})

})
