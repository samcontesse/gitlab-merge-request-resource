package pkg

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

type Source struct {
	URI                string   `json:"uri"`
	PrivateToken       string   `json:"private_token"`
	Insecure           bool     `json:"insecure"`
	Recursive          bool     `json:"recursive,omitempty"`
	SkipWorkInProgress bool     `json:"skip_work_in_progress,omitempty"`
	SkipNotMergeable   bool     `json:"skip_not_mergeable,omitempty"`
	SkipTriggerComment bool     `json:"skip_trigger_comment,omitempty"`
	ConcourseUrl       string   `json:"concourse_url,omitempty"`
	PipelineName       string   `json:"pipeline_name,omitempty"`
	Labels             []string `json:"labels,omitempty"`
	TargetBranch       string   `json:"target_branch,omitempty"`
	SourceBranch       string   `json:"source_branch,omitempty"`
	Sort               string   `json:"sort,omitempty"`
	Paths              []string `json:"paths,omitempty"`
	IgnorePaths        []string `json:"ignore_paths,omitempty"`
	SshKeys            []string `json:"ssh_keys,omitempty"`
}

type Version struct {
	ID        int        `json:"id,string"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type Metadata []MetadataField

type MetadataField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// GetBaseURL extracts host from URI (repository URL) and appends the v3 API suffix.
func (source *Source) GetBaseURL() string {
	r, _ := regexp.Compile("https?://[^/]+")
	host := r.FindString(source.URI)
	return host + "/api/v4"
}

// GetProjectPath extracts project path from URI (repository URL).
func (source *Source) GetProjectPath() string {
	r, _ := regexp.Compile("(https?|ssh)://([^/]*)/(.*)\\.git$")
	return r.FindStringSubmatch(source.URI)[3]
}

func (source *Source) GetTargetURL() string {
	target, _ := url.Parse(source.GetCoucourseUrl())
	target.Path += "/teams/" + url.QueryEscape(os.Getenv("BUILD_TEAM_NAME"))
	target.Path += "/pipelines/" + url.QueryEscape(os.Getenv("BUILD_PIPELINE_NAME"))
	target.Path += "/jobs/" + url.QueryEscape(os.Getenv("BUILD_JOB_NAME"))
	target.Path += "/builds/" + url.QueryEscape(os.Getenv("BUILD_NAME"))

	query := os.Getenv("BUILD_PIPELINE_INSTANCE_VARS")
	if query != "" {
		values := url.Values{}
		var vars map[string]string
		err := json.Unmarshal([]byte(query), &vars)
		if err == nil {
			for k, v := range vars {
				values.Set(fmt.Sprintf("vars.%s", k), fmt.Sprintf(`"%s"`, v))
			}
			target.RawQuery = values.Encode()
		}
	}

	return target.String()
}

func (source *Source) GetCoucourseUrl() string {
	if source.ConcourseUrl != "" {
		return source.ConcourseUrl
	} else {
		return os.Getenv("ATC_EXTERNAL_URL")
	}
}

func (source *Source) GetPipelineName() string {
	if source.PipelineName != "" {
		return source.PipelineName
	} else {
		return os.Getenv("BUILD_PIPELINE_NAME")
	}

}

func (source *Source) GetSort() (string, error) {
	order := strings.ToLower(source.Sort)
	switch order {
	case "":
		return "asc", nil
	case "asc", "desc":
		return order, nil
	}
	return "", fmt.Errorf("invalid value for sort: %v", source.Sort)
}

func (source *Source) AcceptPath(path string) bool {

	excluded := len(source.IgnorePaths) > 0
	included := len(source.Paths) == 0

	if excluded {
		excluded = matchPath(source.IgnorePaths, path)
	}

	if len(source.Paths) > 0 {
		included = matchPath(source.Paths, path)
	}

	return !excluded && included
}
