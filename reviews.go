package registry

import (
	"context"
	"fmt"
)

// ListReviews returns a paginated list of reviews for a plugin.
func (c *Client) ListReviews(ctx context.Context, pluginID string, opts *ListOptions) (ListResult[Review], error) {
	path := fmt.Sprintf("/v1/plugins/%s/reviews", pluginID) + opts.buildQuery()
	var items []Review
	pag, err := c.getList(ctx, path, &items)
	if err != nil {
		return ListResult[Review]{}, err
	}
	if items == nil {
		items = []Review{}
	}
	return ListResult[Review]{Items: items, Pagination: pag}, nil
}

// CreateReview creates a review for a plugin. Requires authentication (WithToken).
func (c *Client) CreateReview(ctx context.Context, pluginID string, input *CreateReviewInput) (*Review, error) {
	var r Review
	if err := c.post(ctx, fmt.Sprintf("/v1/plugins/%s/reviews", pluginID), input, &r); err != nil {
		return nil, err
	}
	return &r, nil
}
