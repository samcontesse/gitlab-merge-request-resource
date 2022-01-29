package in

import (
	. "github.com/samcontesse/gitlab-merge-request-resource/pkg"
)

type Request struct {
	Source  Source  `json:"source"`
	Version Version `json:"version"`
}

type Response struct {
	Version  Version  `json:"version"`
	Metadata Metadata `json:"metadata"`
}
