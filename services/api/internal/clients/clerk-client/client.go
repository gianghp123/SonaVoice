package clerkclient

import (
	"context"

	clerk "github.com/clerk/clerk-sdk-go/v2"
	clerkuser "github.com/clerk/clerk-sdk-go/v2/user"
)

type IClerkClient interface {
	UpdateMetadata(ctx context.Context, id string, params *clerkuser.UpdateMetadataParams) (*clerk.User, error)
}

func NewClient() IClerkClient {
	return &clerkuser.Client{
		Backend: clerk.GetBackend(),
	}
}
