package user

import "context"

type relationshipReader interface {
	GetPartnerByUserID(ctx context.Context, userID string) (string, error)
	IsFollowing(ctx context.Context, followerID, followeeID string) (bool, error)
}

type Client struct {
	relationshipReader relationshipReader
}

func NewClient(relationshipReader relationshipReader) *Client {
	return &Client{relationshipReader: relationshipReader}
}

func (c *Client) GetPartnerByUserID(ctx context.Context, userID string) (string, error) {
	return c.relationshipReader.GetPartnerByUserID(ctx, userID)
}

func (c *Client) IsFollowing(ctx context.Context, followerID, followeeID string) (bool, error) {
	return c.relationshipReader.IsFollowing(ctx, followerID, followeeID)
}
