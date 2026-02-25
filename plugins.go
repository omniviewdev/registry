package registry

import (
	"context"
	"fmt"
)

// ListPlugins returns a paginated list of plugins.
func (c *Client) ListPlugins(ctx context.Context, opts *ListOptions) (ListResult[Plugin], error) {
	path := "/v1/plugins" + opts.buildQuery()
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

// GetPlugin returns a single plugin by ID.
func (c *Client) GetPlugin(ctx context.Context, pluginID string) (*Plugin, error) {
	var p Plugin
	if err := c.get(ctx, fmt.Sprintf("/v1/plugins/%s", pluginID), &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// ListCategories returns all categories with plugin counts.
func (c *Client) ListCategories(ctx context.Context) ([]CategoryCount, error) {
	var cats []CategoryCount
	if err := c.get(ctx, "/v1/categories", &cats); err != nil {
		return nil, err
	}
	return cats, nil
}
