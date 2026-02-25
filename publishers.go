package registry

import (
	"context"
	"fmt"
)

// GetPublisher returns a publisher by slug.
func (c *Client) GetPublisher(ctx context.Context, slug string) (*Publisher, error) {
	var p Publisher
	if err := c.get(ctx, fmt.Sprintf("/v1/publishers/%s", slug), &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// ListPublisherPlugins returns plugins for a publisher by slug.
func (c *Client) ListPublisherPlugins(ctx context.Context, slug string, opts *ListOptions) (ListResult[Plugin], error) {
	path := fmt.Sprintf("/v1/publishers/%s/plugins", slug) + opts.buildQuery()
	var items []Plugin
	pag, err := c.getList(ctx, path, &items)
	if err != nil {
		return ListResult[Plugin]{}, err
	}
	if items == nil {
		items = []Plugin{}
	}
	return ListResult[Plugin]{Items: items, Pagination: pag}, nil
}
