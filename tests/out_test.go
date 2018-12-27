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
		srcDir    string
		req       out.Request
		res       out.Response
		gitlabURL string
	)

	BeforeEach(func() {
		var err error
		srcDir, err = ioutil.TempDir("", "gitlab-merge-request-resource-dir")
		Expect(err).ToNot(HaveOccurred())
		gitlabURL = setup()
		mux.HandleFunc("/api/v4/projects/1/statuses/abc", func(w http.ResponseWriter, r *http.Request) {
			var commitStatus gitlab.CommitStatus = gitlab.CommitStatus{
				ID:          1,
				SHA:         "12311",
				Ref:         "",
				Status:      "",
				Name:        "",
				TargetURL:   "",
				Description: "",
			}

			output, err := json.Marshal(commitStatus)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			w.Header().Set("content-type", "application/json")
			w.Write(output)
		})
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

	Describe("image metadata", func() {

		BeforeEach(func() {
			_ = os.Mkdir(
				path.Join(srcDir, "repo"),
				0744)
			_ = os.Mkdir(
				path.Join(srcDir, "repo", ".git"),
				0744)
			_ = ioutil.WriteFile(
				path.Join(srcDir, "repo", ".git", "merge-request.json"),
				[]byte(Fixture("merge-request.json")),
				0744)

			req = out.Request{
				Source: resource.Source{
					URI:          gitlabURL,
					PrivateToken: "$$random",
					Insecure:     false,
				},
				Params: out.Params{
					Repository: "repo",
					Status:     "running",
				},
			}
		})

		It("works", func() {
			Expect(res.Version.ID).To(Equal(1))
		})

	})

})

func setup() string {
	os.Setenv("GITLAB_TOKEN", "$$$randome")

	// test server
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)
	base, _ := url.Parse(server.URL)

	u, err := url.Parse("/api/v4")
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
