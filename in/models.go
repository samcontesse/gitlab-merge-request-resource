package in

import (
	"github.com/samcontesse/gitlab-merge-request-resource"
)

type Request struct {
	Source  resource.Source  `json:"source"`
	Version resource.Version `json:"version"`
	Params Params            `json:"params"`
}

type Response struct {
	Version  resource.Version  `json:"version"`
	Metadata resource.Metadata `json:"metadata"`
}

type Params struct {
	Submodules string   `json:"submodules"`
}
