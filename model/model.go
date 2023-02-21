package model

import (
	"fmt"
	"reflect"
)

type Audited struct {
	CreatedBy      string
	Created        int
	LastModifiedBy string
	LastModified   int
}

type PublicAssetVersion struct {
	Audited      `json:"audited,omitempty"`
	Version      string           `json:"version,omitempty"`
	Checksum     string           `json:"checksum,omitempty"`
	Artifact     Artifact         `json:"artifact,omitempty"`
	Repository   Repository       `json:"repository,omitempty"`
	Content      AssetContent     `json:"content,omitempty"`
	Dependencies []AssetReference `json:"dependencies,omitempty"`
	Readme       TypedText        `json:"readme,omitempty"`
	Current      bool             `json:"current,omitempty"`
}

type AssetReference struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type TypedText struct {
	Type    string `json:"type,omitempty"`
	Content string `json:"content,omitempty"`
}
type Repository struct {
	Type    string      `json:"type,omitempty"`
	Main    bool        `json:"main,omitempty"`
	Branch  string      `json:"branch,omitempty"`
	Commit  string      `json:"commit,omitempty"`
	Details interface{} `json:"details,omitempty"`
}

type GitRepository struct {
	Type    string      `json:"type,omitempty"`
	Main    bool        `json:"main,omitempty"`
	Branch  string      `json:"branch,omitempty"`
	Commit  string      `json:"commit,omitempty"`
	Details *GitDetails `json:"details,omitempty"`
}

type NoRepository struct {
	Type    string      `json:"type,omitempty"`
	Main    bool        `json:"main,omitempty"`
	Branch  string      `json:"branch,omitempty"`
	Commit  string      `json:"commit,omitempty"`
	Details interface{} `json:"details,omitempty"`
}

type GitDetails struct {
	URL    string `json:"url,omitempty"`
	Remote string `json:"remote,omitempty"`
	Path   string `json:"path,omitempty"`
}

type AssetContent struct {
	Kind     string                 `json:"kind"`
	Metadata Metadata               `json:"metadata"`
	Spec     map[string]interface{} `json:"spec"`
}

type Metadata struct {
	// Name is the unique identifier for the resource. Must have a format such as "handle/some-type"
	Name string `json:"name"`

	// Title is the human-readable name for the resource.
	Title string `json:"title"`

	// Description is a free text description of the resource.
	Description string `json:"description"`

	// visibility is the visibility of the resource, this can be public or private.
	Visibility string `json:"visibility" default:"public"`
}

type Artifact struct {
	Type    string      `json:"type,omitempty"`
	Details interface{} `json:"details,omitempty"`
}
type YAMLArtifact struct {
	Artifact *Artifact   `json:"artifact,omitempty"`
	Details  YAMLDetails `json:"details,omitempty"`
}
type DockerArtifact struct {
	Artifact *Artifact     `json:"artifact,omitempty"`
	Details  DockerDetails `json:"details,omitempty"`
}

type NPMArtifact struct {
	Artifact *Artifact  `json:"artifact,omitempty"`
	Details  NPMDetails `json:"details,omitempty"`
}

type MavenArtifact struct {
	Artifact *Artifact    `json:"artifact,omitempty"`
	Details  MavenDetails `json:"details,omitempty"`
}

type DockerDetails struct {
	Name    string   `json:"name,omitempty"`
	Primary string   `json:"primary,omitempty"`
	Tags    []string `json:"tags,omitempty"`
}

type NPMDetails struct {
	Name     string `json:"name,omitempty"`
	Version  string `json:"version,omitempty"`
	Registry string `json:"registry,omitempty"`
}

type MavenDetails struct {
	GroupId  string `json:"group_id,omitempty"`
	Artifact string `json:"artifact,omitempty"`
	Version  string `json:"version,omitempty"`
	Registry string `json:"registry,omitempty"`
}

type YAMLDetails struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
}

func (r *Repository) AsGit() *GitRepository {
	if _, ok := r.Details.(*GitDetails); !ok {
		panic(fmt.Errorf("invalid type: %T", r.Details))
	}
	return &GitRepository{
		Type:    r.Type,
		Main:    r.Main,
		Branch:  r.Branch,
		Commit:  r.Commit,
		Details: r.Details.(*GitDetails),
	}
}

func (a *Artifact) AsDocker() *DockerArtifact {
	if reflect.TypeOf(a.Details) != reflect.TypeOf(&DockerDetails{}) {
		panic("Invalid type")
	}
	return &DockerArtifact{
		Artifact: a,
	}
}

func (a *Artifact) AsNPM() *NPMArtifact {
	if reflect.TypeOf(a.Details) != reflect.TypeOf(&NPMDetails{}) {
		panic("Invalid type")
	}
	return &NPMArtifact{
		Artifact: a,
	}
}

func (a *Artifact) AsMaven() *MavenArtifact {
	if reflect.TypeOf(a.Details) != reflect.TypeOf(&MavenDetails{}) {
		panic("Invalid type")
	}
	return &MavenArtifact{
		Artifact: a,
	}
}

func (a *Artifact) AsYAML() *YAMLArtifact {
	if reflect.TypeOf(a.Details) != reflect.TypeOf(&YAMLDetails{}) {
		panic("Invalid type")
	}
	return &YAMLArtifact{
		Artifact: a,
	}
}
