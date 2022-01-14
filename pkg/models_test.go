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

func TestSource_AcceptPath(t *testing.T) {
	type fields struct {
		Paths       []string
		IgnorePaths []string
	}
	type args struct {
		path string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "match with no include and no exclude",
			fields: fields{
				Paths:       nil,
				IgnorePaths: nil,
			},
			args: args{
				path: "Makefile",
			},
			want: true,
		},
		{
			name: "match with include and no exclude",
			fields: fields{
				Paths:       []string{"Makefile"},
				IgnorePaths: nil,
			},
			args: args{
				path: "Makefile",
			},
			want: true,
		},
		{
			name: "match with glob include and no exclude",
			fields: fields{
				Paths:       []string{"**/README.md"},
				IgnorePaths: nil,
			},
			args: args{
				path: "README.md",
			},
			want: false,
		},
		{
			name: "match with include pattern and no exclude",
			fields: fields{
				Paths:       []string{"Make*"},
				IgnorePaths: nil,
			},
			args: args{
				path: "Makefile",
			},
			want: true,
		},
		{
			name: "no match with include pattern and no exclude",
			fields: fields{
				Paths:       []string{"Other*"},
				IgnorePaths: nil,
			},
			args: args{
				path: "Makefile",
			},
			want: false,
		},
		{
			name: "no match with include pattern and exclude pattern",
			fields: fields{
				Paths:       []string{"Make*"},
				IgnorePaths: []string{"Makefi??"},
			},
			args: args{
				path: "Makefile",
			},
			want: false,
		},
		{
			name: "no match with include and no exclude",
			fields: fields{
				Paths:       []string{"Makefile"},
				IgnorePaths: nil,
			},
			args: args{
				path: "Other",
			},
			want: false,
		},
		{
			name: "no match with include and exclude",
			fields: fields{
				Paths:       []string{"Makefile"},
				IgnorePaths: []string{"Makefile"},
			},
			args: args{
				path: "Makefile",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := &Source{
				Paths:       tt.fields.Paths,
				IgnorePaths: tt.fields.IgnorePaths,
			}
			if got := source.AcceptPath(tt.args.path); got != tt.want {
				t.Errorf("AcceptPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
