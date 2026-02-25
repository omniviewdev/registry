package registry

import (
	"context"
	"fmt"
)

// ListVersions returns a paginated list of versions for a plugin.
func (c *Client) ListVersions(ctx context.Context, pluginID string, opts *ListOptions) (ListResult[PluginVersion], error) {
	path := fmt.Sprintf("/v1/plugins/%s/versions", pluginID) + opts.buildQuery()
	var items []PluginVersion
	pag, err := c.getList(ctx, path, &items)
	if err != nil {
		return ListResult[PluginVersion]{}, err
	}
	if items == nil {
		items = []PluginVersion{}
	}
	return ListResult[PluginVersion]{Items: items, Pagination: pag}, nil
}

// GetVersion returns a specific version of a plugin.
func (c *Client) GetVersion(ctx context.Context, pluginID, version string) (*PluginVersion, error) {
	var v PluginVersion
	if err := c.get(ctx, fmt.Sprintf("/v1/plugins/%s/versions/%s", pluginID, version), &v); err != nil {
		return nil, err
	}
	return &v, nil
}
