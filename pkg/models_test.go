package pkg

import "testing"

func TestSource_GetProjectPath(t *testing.T) {
	tests := []struct {
		uri  string
		want string
	}{
		{uri: "https://git.example.com/project.git", want: "project"},
		{uri: "https://git.example.com/namespace/project.git", want: "namespace/project"},
		{uri: "https://git.example.com/group/subgroup1/subgroup2/project.git", want: "group/subgroup1/subgroup2/project"},
	}
	for _, tt := range tests {
		t.Run(tt.uri, func(t *testing.T) {
			source := Source{URI: tt.uri}
			got := source.GetProjectPath()
			if got != tt.want {
				t.Errorf("GetProjectPath() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSource_GetSort(t *testing.T) {
	tests := []struct {
		sort    string
		want    string
		wantErr bool
	}{
		{"asc", "asc", false},
		{"desc", "desc", false},
		{"", "asc", false},
		{"invalid", "", true},
		{"AsC", "asc", false},
		{"DESC", "desc", false},
	}
	for _, tt := range tests {
		t.Run(tt.sort, func(t *testing.T) {
			source := Source{
				URI:  "https://git.example.com/project.git",
				Sort: tt.sort,
			}
			got, err := source.GetSort()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSort() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("GetSort() got = %v, want %v", got, tt.want)
			}
		})
	}
}
