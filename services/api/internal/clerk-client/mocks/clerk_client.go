package mocks

import (
	"context"

	clerk "github.com/clerk/clerk-sdk-go/v2"
	clerkuser "github.com/clerk/clerk-sdk-go/v2/user"
	"github.com/stretchr/testify/mock"
)

type ClerkClient struct {
	mock.Mock
}

func (m *ClerkClient) UpdateMetadata(ctx context.Context, id string, params *clerkuser.UpdateMetadataParams) (*clerk.User, error) {
	args := m.Called(ctx, id, params)
	var user *clerk.User
	if args.Get(0) != nil {
		user = args.Get(0).(*clerk.User)
	}
	return user, args.Error(1)
}
