package registry

import (
	"context"
	"fmt"
	"time"
)

// Submission represents a plugin submission in the review pipeline.
type Submission struct {
	ID                  string     `json:"id"`
	PublisherID         string     `json:"publisher_id"`
	PluginID            string     `json:"plugin_id"`
	Version             string     `json:"version"`
	Status              string     `json:"status"`
	ArtifactS3Prefix    string     `json:"artifact_s3_prefix"`
	Metadata            string     `json:"metadata_json"`
	ValidationResult    string     `json:"validation_result"`
	SubmittedByID       uint       `json:"submitted_by_id"`
	ReviewerID          *uint      `json:"reviewer_id"`
	ReviewerNotes       string     `json:"reviewer_notes"`
	Changelog           string     `json:"changelog"`
	SubmittedAt         *time.Time `json:"submitted_at"`
	ValidationStartedAt *time.Time `json:"validation_started_at"`
	ReviewedAt          *time.Time `json:"reviewed_at"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// CreateSubmissionRequest is the request body for creating a submission.
type CreateSubmissionRequest struct {
	PluginID  string `json:"plugin_id"`
	Version   string `json:"version"`
	Changelog string `json:"changelog"`
}

// UploadURLRequest is the request body for generating presigned upload URLs.
type UploadURLRequest struct {
	Architectures []string `json:"architectures"`
}

// UploadURLResponse contains presigned URLs keyed by architecture.
type UploadURLResponse struct {
	URLs map[string]string `json:"urls"`
}

// MultipartUpload contains a presigned URL for each part, in part-number order.
type MultipartUpload struct {
	UploadID string   `json:"upload_id"`
	PartURLs []string `json:"part_urls"`
}

// MultipartPart identifies an uploaded part by its one-based number and S3 ETag.
type MultipartPart struct {
	PartNumber int    `json:"part_number"`
	ETag       string `json:"etag"`
}

// CreateSubmission creates a new plugin submission for a publisher.
func (c *Client) CreateSubmission(ctx context.Context, publisherSlug string, req *CreateSubmissionRequest) (*Submission, error) {
	var sub Submission
	path := fmt.Sprintf("/v1/publishers/%s/submissions", publisherSlug)
	if err := c.post(ctx, path, req, &sub); err != nil {
		return nil, err
	}
	return &sub, nil
}

// GetSubmission returns a submission by ID.
func (c *Client) GetSubmission(ctx context.Context, id string) (*Submission, error) {
	var sub Submission
	path := fmt.Sprintf("/v1/submissions/%s", id)
	if err := c.get(ctx, path, &sub); err != nil {
		return nil, err
	}
	return &sub, nil
}

// ListSubmissions returns submissions for a publisher.
func (c *Client) ListSubmissions(ctx context.Context, publisherSlug string, opts *ListOptions) (ListResult[Submission], error) {
	path := fmt.Sprintf("/v1/publishers/%s/submissions", publisherSlug)
	if opts != nil {
		path += opts.buildQuery()
	}
	var items []Submission
	pg, err := c.getList(ctx, path, &items)
	if err != nil {
		return ListResult[Submission]{}, err
	}
	return ListResult[Submission]{Items: items, Pagination: pg}, nil
}

// SubmitForReview transitions a submission to pending review.
func (c *Client) SubmitForReview(ctx context.Context, id string) (*Submission, error) {
	var sub Submission
	path := fmt.Sprintf("/v1/submissions/%s/submit", id)
	if err := c.post(ctx, path, nil, &sub); err != nil {
		return nil, err
	}
	return &sub, nil
}

// WithdrawSubmission withdraws a submission from review.
func (c *Client) WithdrawSubmission(ctx context.Context, id string) (*Submission, error) {
	var sub Submission
	path := fmt.Sprintf("/v1/submissions/%s/withdraw", id)
	if err := c.post(ctx, path, nil, &sub); err != nil {
		return nil, err
	}
	return &sub, nil
}

// GenerateUploadURLs generates presigned upload URLs for submission artifacts.
func (c *Client) GenerateUploadURLs(ctx context.Context, id string, architectures []string) (*UploadURLResponse, error) {
	var resp UploadURLResponse
	path := fmt.Sprintf("/v1/submissions/%s/upload-urls", id)
	if err := c.post(ctx, path, &UploadURLRequest{Architectures: architectures}, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// InitMultipartUpload starts an upload. Use 64 MiB parts except for the last.
func (c *Client) InitMultipartUpload(ctx context.Context, id, architecture string, partCount int) (*MultipartUpload, error) {
	var upload MultipartUpload
	body := map[string]interface{}{"architecture": architecture, "part_count": partCount}
	if err := c.post(ctx, fmt.Sprintf("/v1/submissions/%s/multipart/init", id), body, &upload); err != nil {
		return nil, err
	}
	return &upload, nil
}

// CompleteMultipartUpload completes an upload with the ETag returned for each part.
func (c *Client) CompleteMultipartUpload(ctx context.Context, id, architecture, uploadID string, parts []MultipartPart) error {
	body := map[string]interface{}{"architecture": architecture, "upload_id": uploadID, "parts": parts}
	return c.post(ctx, fmt.Sprintf("/v1/submissions/%s/multipart/complete", id), body, nil)
}

// AbortMultipartUpload releases an incomplete upload.
func (c *Client) AbortMultipartUpload(ctx context.Context, id, architecture, uploadID string) error {
	body := map[string]string{"architecture": architecture, "upload_id": uploadID}
	return c.post(ctx, fmt.Sprintf("/v1/submissions/%s/multipart/abort", id), body, nil)
}
