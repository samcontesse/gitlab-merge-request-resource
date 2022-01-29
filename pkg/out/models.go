package out

import (
	"github.com/samcontesse/gitlab-merge-request-resource/pkg"
	"io/ioutil"
	"path"
	"strings"
)

type Request struct {
	Source pkg.Source `json:"source"`
	Params Params     `json:"params"`
}

type Response struct {
	Version  pkg.Version  `json:"version"`
	Metadata pkg.Metadata `json:"metadata"`
}

type Params struct {
	Repository string   `json:"repository"`
	Status     string   `json:"status"`
	Labels     []string `json:"labels"`
	Comment    Comment  `json:"comment"`
}

type Comment struct {
	FilePath string `json:"file"`
	Text     string `json:"text"`
}

func (comment Comment) ReadContent(folder string) (string, error) {
	var (
		commentContent string
		fileContent    string
	)
	if comment.FilePath != "" {
		content, err := ioutil.ReadFile(path.Join(folder, comment.FilePath))
		if err != nil {
			return "", err
		} else {
			commentContent = string(content)
			fileContent = string(content)
		}
	}

	if comment.Text != "" {
		commentRaw := comment.Text
		commentContent = strings.Replace(commentRaw, "$FILE_CONTENT", fileContent, -1)
	}

	return commentContent, nil
}
