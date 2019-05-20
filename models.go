package resource

import (
	"net/url"
	"os"
	"regexp"
	"time"
)

type Source struct {
	URI                string   `json:"uri"`
	PrivateToken       string   `json:"private_token"`
	Insecure           bool     `json:"insecure"`
	SkipWorkInProgress bool     `json:"skip_work_in_progress,omitempty"`
	SkipNotMergeable   bool     `json:"skip_not_mergeable,omitempty"`
	SkipTriggerComment bool     `json:"skip_trigger_comment,omitempty"`
	OnlyTriggerComment string   `json:"only_trigger_comment,omitempty"`
	ConcourseUrl       string   `json:"concourse_url,omitempty"`
	Labels             []string `json:"labels,omitempty"`
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
	return target.String()
}

func (source *Source) GetCoucourseUrl() string {
	if source.ConcourseUrl != "" {
		return source.ConcourseUrl
	} else {
		return os.Getenv("ATC_EXTERNAL_URL")
	}
}

func GetPipelineName() string {
	return os.Getenv("BUILD_PIPELINE_NAME")
}
