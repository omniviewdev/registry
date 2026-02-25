package registry

import "time"

// Plugin represents a plugin in the registry.
type Plugin struct {
	ID            string       `json:"id"`
	Name          string       `json:"name"`
	Description   string       `json:"description"`
	IconURL       string       `json:"icon_url"`
	Category      string       `json:"category"`
	Tags          []string     `json:"tags"`
	License       string       `json:"license"`
	Repository    string       `json:"repository"`
	URL           string       `json:"url"`
	Readme        string       `json:"readme"`
	Official      bool         `json:"official"`
	Featured      bool         `json:"featured"`
	DownloadCount int64        `json:"download_count"`
	AverageRating float64      `json:"average_rating"`
	ReviewCount   int64        `json:"review_count"`
	PublisherID   string       `json:"publisher_id"`
	Author        PluginAuthor `json:"author"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
}

// PluginAuthor identifies the author of a plugin.
type PluginAuthor struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	URL   string `json:"url"`
}

// PluginVersion represents a specific version of a plugin.
type PluginVersion struct {
	ID            string              `json:"id"`
	PluginID      string              `json:"plugin_id"`
	Version       string              `json:"version"`
	Description   string              `json:"description"`
	Changelog     string              `json:"changelog"`
	MinIDEVersion string              `json:"min_ide_version"`
	MaxIDEVersion string              `json:"max_ide_version"`
	Capabilities  []string            `json:"capabilities"`
	Visible       bool                `json:"visible"`
	Artifacts     map[string]Artifact `json:"artifacts,omitempty"`
	CreatedAt     time.Time           `json:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at"`
}

// Artifact represents a platform-specific build artifact.
type Artifact struct {
	Checksum    string `json:"checksum"`     // SHA-256 hex
	Signature   string `json:"signature"`    // base64 Ed25519
	DownloadURL string `json:"download_url"` // relative CDN path
	Size        int64  `json:"size"`
}

// Review represents a user review of a plugin.
type Review struct {
	ID         string     `json:"id"`
	PluginID   string     `json:"plugin_id"`
	UserID     uint       `json:"user_id"`
	Rating     int        `json:"rating"`
	Title      string     `json:"title"`
	Body       string     `json:"body"`
	Response   string     `json:"response,omitempty"`
	ResponseAt *time.Time `json:"response_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// CreateReviewInput is the request body for creating a review.
type CreateReviewInput struct {
	Rating int    `json:"rating"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}

// Publisher represents a plugin publisher organization.
type Publisher struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	Website     string    `json:"website"`
	Logo        string    `json:"logo"`
	Verified    bool      `json:"verified"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// DownloadStats holds aggregate download statistics for a plugin.
type DownloadStats struct {
	PluginID   string             `json:"plugin_id"`
	TotalCount int64              `json:"total_count"`
	ByVersion  []VersionDownloads `json:"by_version,omitempty"`
}

// VersionDownloads holds per-version download counts.
type VersionDownloads struct {
	Version string `json:"version"`
	Count   int64  `json:"count"`
}

// DailyDownloads holds a single day's download count.
type DailyDownloads struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// CategoryCount holds a category and its plugin count.
type CategoryCount struct {
	Category string `json:"category"`
	Count    int64  `json:"count"`
}
