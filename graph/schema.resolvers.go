package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/xngln/photo-server/graph/generated"
	"github.com/xngln/photo-server/graph/model"
)

func (r *mutationResolver) UploadImage(ctx context.Context, input model.NewImage) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) DeleteImage(ctx context.Context, id string) (*model.Image, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CreateCheckoutSession(ctx context.Context, photoID string) (string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Image(ctx context.Context, id string) (*model.Image, error) {
	dummyImage := model.Image{
		ID:           "abc123z",
		Name:         "tokyo",
		Price:        1.00,
		ThumbnailURL: "google.com",
		FullsizeURL:  "amazon.com",
	}
	return &dummyImage, nil
}

func (r *queryResolver) Images(ctx context.Context) ([]*model.Image, error) {
	panic(fmt.Errorf("not implemented"))
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
