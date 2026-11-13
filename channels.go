package registry

import (
	"context"
	"fmt"
	"time"
)

// PluginChannel describes a promoted plugin release channel.
type PluginChannel struct {
	Channel   string    `json:"channel"`
	Branch    string    `json:"branch"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// ListChannels returns the active release channels for a plugin.
func (c *Client) ListChannels(ctx context.Context, pluginID string) ([]PluginChannel, error) {
	var channels []PluginChannel
	if err := c.get(ctx, fmt.Sprintf("/v1/plugins/%s/channels", pluginID), &channels); err != nil {
		return nil, err
	}
	return channels, nil
}
