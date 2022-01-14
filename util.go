package resource

import "path/filepath"

func matchPath(patterns []string, path string) bool {
	for _, pattern := range patterns {
		ok, _ := filepath.Match(pattern, path)
		if ok {
			return true
		}
	}
	return false
}
