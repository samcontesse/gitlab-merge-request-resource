package out_test

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestOut(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Out Suite")
}

func fixture(filename string) string {
	path := filepath.Join("fixtures", filename)
	contents, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	return string(contents)
}
