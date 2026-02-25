package registry

// PluginMeta is the canonical plugin metadata type.
// This replaces copies in plugin-sdk and registry-cli.
type PluginMeta struct {
	ID            string       `json:"id"               yaml:"id"`
	Version       string       `json:"version"          yaml:"version"`
	Name          string       `json:"name"             yaml:"name"`
	Icon          string       `json:"icon"             yaml:"icon"`
	IconURL       string       `json:"icon_url"         yaml:"icon_url,omitempty"`
	Description   string       `json:"description"      yaml:"description"`
	Repository    string       `json:"repository"       yaml:"repository"`
	Website       string       `json:"website"          yaml:"website"`
	MinIDEVersion string       `json:"min_ide_version"  yaml:"min_ide_version"`
	MaxIDEVersion string       `json:"max_ide_version"  yaml:"max_ide_version"`
	Category      string       `json:"category"         yaml:"category"`
	License       string       `json:"license"          yaml:"license,omitempty"`
	Author        *Author      `json:"author,omitempty" yaml:"author,omitempty"`
	Maintainers   []Maintainer `json:"maintainers"      yaml:"maintainers"`
	Tags          []string     `json:"tags"             yaml:"tags"`
	Dependencies  []string     `json:"dependencies"     yaml:"dependencies"`
	Capabilities  []string     `json:"capabilities"     yaml:"capabilities"`
	Theme         PluginTheme  `json:"theme"            yaml:"theme"`
}

// Author identifies the primary author of a plugin.
type Author struct {
	Name  string `json:"name"  yaml:"name"`
	Email string `json:"email" yaml:"email,omitempty"`
	URL   string `json:"url"   yaml:"url,omitempty"`
}

// Maintainer identifies a plugin maintainer.
type Maintainer struct {
	Name  string `json:"name"  yaml:"name"`
	Email string `json:"email" yaml:"email"`
}

// PluginTheme controls how the plugin is displayed in the IDE.
type PluginTheme struct {
	PrimaryColor string `json:"primary_color" yaml:"primary_color"`
	DarkMode     bool   `json:"dark_mode"     yaml:"dark_mode"`
}
