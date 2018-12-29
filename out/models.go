package out

import (
	"github.com/samcontesse/gitlab-merge-request-resource"
	"github.com/samcontesse/gitlab-merge-request-resource/common"
	"io/ioutil"
	"path"
	"strings"
)

type Request struct {
	Source resource.Source `json:"source"`
	Params Params          `json:"params"`
}

type Response struct {
	Version  resource.Version  `json:"version"`
	Metadata resource.Metadata `json:"metadata"`
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

// Generate comment content
func (comment Comment) GetContent(basePath string) string {
	var (
		commentContent string
		fileContent    string
	)
	if comment.FilePath != "" {
		filePath := path.Join(basePath, comment.FilePath)
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			common.Fatal("Can't read from "+filePath, err)
		} else {
			commentContent = string(content)
			fileContent = string(content)
		}
	}

	if comment.Text != "" {
		commentRaw := comment.Text
		commentContent = strings.Replace(commentRaw, "$FILE_CONTENT", fileContent, -1)
	}

	return commentContent
}
