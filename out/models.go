package out

import (
	"github.com/samcontesse/gitlab-merge-request-resource"
)

type Request struct {
	Source resource.Source `json:"source"`
	Params Params          `json:"params"`
}

type Response struct {
	Version resource.Version `json:"version"`
}

type Params struct {
	Repository string `json:"repository"`
	Status     string `json:"status"`
}
