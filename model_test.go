package resource

import "testing"

func TestSource_GetProjectPath_WithNamespace(t *testing.T) {
	source := Source{
		URI: "https://git.example.com/namespace/project.git",
	}

	actual := source.GetProjectPath()
	expected := "namespace/project"

	if actual != expected {
		t.Errorf("Project path %s expected to be %s", actual, expected)
	}
}

func TestSource_GetProjectPath_WithSubgroup(t *testing.T) {
	source := Source{
		URI: "https://git.example.com/group/subgroup1/subgroup2/project.git",
	}

	actual := source.GetProjectPath()
	expected := "group/subgroup1/subgroup2/project"

	if actual != expected {
		t.Errorf("Project path %s expected to be %s", actual, expected)
	}
}

func TestSource_GetProjectPath_WithoutNamespace(t *testing.T) {
	source := Source{
		URI: "https://git.example.com/project.git",
	}

	actual := source.GetProjectPath()
	expected := "project"

	if actual != expected {
		t.Errorf("Project path %s expected to be %s", actual, expected)
	}
}

func TestSource_GetSort(t *testing.T) {
	tests := []struct {
		sort string
		want string
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
				URI: "https://git.example.com/project.git",
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
