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
	SkipWorkInProgress bool     `json:"skip_work_in_progress,omitempty"`
	SkipNotMergeable   bool     `json:"skip_not_mergeable,omitempty"`
	SkipTriggerComment bool     `json:"skip_trigger_comment,omitempty"`
	Labels             []string `json:"labels,omitempty"`
}

type Version struct {
	ID        int       `json:"id,string"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type Metadata []MetadataField

type MetadataField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type BuildMetadata struct {
	BuildId           string `json:"build_id"`
	BuildName         string `json:"build_name"`
	BuildTeamName     string `json:"build_team_name"`
	BuildJobName      string `json:"build_job_name"`
	BuildPipelineName string `json:"build_pipeline_name"`
	AtcExternalURL    string `json:"atc_external_url"`
}

// GetBaseURL extracts host from URI (repository URL) and appends the v3 API suffix.
func (source *Source) GetBaseURL() string {
	r, _ := regexp.Compile("https?://[^/]+")
	host := r.FindString(source.URI)
	return host + "/api/v4"
}

// GetProjectPath extracts project path from URI (repository URL).
func (source *Source) GetProjectPath() string {
	r, _ := regexp.Compile("([^/]+/[^/]+)\\.git/?$")
	return r.FindStringSubmatch(source.URI)[1]
}

// GetCloneURL add the private token in the URI
func (source *Source) GetCloneURL() string {
	r, _ := regexp.Compile("(https?://)(.*)")
	prefix := r.FindStringSubmatch(source.URI)[1]
	suffix := r.FindStringSubmatch(source.URI)[2]
	return prefix + "gitlab-ci-token:" + source.PrivateToken + "@" + suffix;
}

func GetTargetURL() string {
	target, _ := url.Parse(os.Getenv("ATC_EXTERNAL_URL"))
	target.Path += "/teams/" + url.QueryEscape(os.Getenv("BUILD_TEAM_NAME"))
	target.Path += "/pipelines/" + url.QueryEscape(os.Getenv("BUILD_PIPELINE_NAME"))
	target.Path += "/jobs/" + url.QueryEscape(os.Getenv("BUILD_JOB_NAME"))
	target.Path += "/builds/" + url.QueryEscape(os.Getenv("BUILD_NAME"))
	return target.String()
}

func GetPipelineName() string {
	return os.Getenv("BUILD_PIPELINE_NAME")
}
