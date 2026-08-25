package user

import "context"

type clientRepo interface {
	GetPartnerByUserID(ctx context.Context, userID string) (string, error)
	IsFollowing(ctx context.Context, followerID, followeeID string) (bool, error)
}

type Client struct {
	userRepo clientRepo
}

func NewClient(userRepo clientRepo) *Client {
	return &Client{userRepo: userRepo}
}

func (c *Client) GetPartnerByUserID(ctx context.Context, userID string) (string, error) {
	return c.userRepo.GetPartnerByUserID(ctx, userID)
}

func (c *Client) IsFollowing(ctx context.Context, followerID, followeeID string) (bool, error) {
	return c.userRepo.IsFollowing(ctx, followerID, followeeID)
}
