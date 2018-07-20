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
