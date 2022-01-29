package check

import (
	. "github.com/samcontesse/gitlab-merge-request-resource/pkg"
)

type Request struct {
	Source  Source  `json:"source"`
	Version Version `json:"version,omitempty"`
}

type Response []Version
