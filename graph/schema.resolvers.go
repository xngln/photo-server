package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v71"
	stripeSession "github.com/stripe/stripe-go/v71/checkout/session"
	"github.com/xngln/photo-server/graph/generated"
	"github.com/xngln/photo-server/graph/model"
)

var sess, err = session.NewSession(&aws.Config{
	Region: aws.String("us-east-2"),
})

/***
TODO:
update thumbnail and fullsize img upload/delete
***/

const thumbnailBucket = "david-photo-store-images-thumbnails"
const fullsizeBucket = "david-photo-store-images-full"
const presetsBucket = "david-photo-store-lr-presets"

var s3client = s3.New(sess)
var dbclient = dynamodb.New(sess)

func (r *mutationResolver) UploadImage(ctx context.Context, input model.NewImage) (*model.Image, error) {
	uploader := s3manager.NewUploader(sess)

	// upload to thumbnail bucket
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket:    aws.String(thumbnailBucket),
		Key:       aws.String(input.Name),
		Body:      input.File.File,
		GrantRead: aws.String(`uri="http://acs.amazonaws.com/groups/global/AllUsers"`),
	})
	if err != nil {
		// add error handling
		return nil, err
	}

	// upload to fullsize bucket
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(fullsizeBucket),
		Key:    aws.String(input.Name),
		Body:   input.File.File,
	})
	if err != nil {
		// add error handling
		return nil, err
	}

	// add to dynamodb
	item := model.ImageDB{
		ID:    uuid.New().String(),
		Name:  input.Name,
		Price: input.Price,
	}
	av, err := dynamodbattribute.MarshalMap(item)
	if err != nil {
		fmt.Println("error marshalling new image item")
		return nil, err
	}
	dbinput := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String("Images"),
	}
	_, err = dbclient.PutItem(dbinput)
	if err != nil {
		fmt.Println("Got error calling PutItem")
		return nil, err
	}

	// get image urls
	req, _ := s3client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(thumbnailBucket),
		Key:    aws.String(input.Name),
	})
	thumbnailURL, err := req.Presign(5 * time.Minute)
	if err != nil {
		log.Println("Failed to sign request", err)
	}

	req, _ = s3client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(fullsizeBucket),
		Key:    aws.String(input.Name),
	})
	fullsizeURL, err := req.Presign(5 * time.Minute)
	if err != nil {
		log.Println("Failed to sign request", err)
	}

	ret := model.Image{
		ID:           item.ID,
		Name:         input.Name,
		Price:        input.Price,
		ThumbnailURL: thumbnailURL,
		FullsizeURL:  fullsizeURL,
	}

	return &ret, nil
}

func (r *mutationResolver) DeleteImage(ctx context.Context, id string) (*model.Image, error) {
	// get image name from db
	result, err := dbclient.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("Images"),
		Key: map[string]*dynamodb.AttributeValue{
			"_id": {S: aws.String(id)},
		},
	})
	if err != nil || result.Item == nil {
		fmt.Println("error getting item")
		return nil, err
	}
	image := model.ImageDB{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &image)

	// delete thumbnail from s3
	_, err = s3client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(thumbnailBucket),
		Key:    aws.String(image.Name),
	})
	if err != nil {
		fmt.Println("error deleting object from bucket")
		return nil, err
	}
	err = s3client.WaitUntilObjectNotExists(&s3.HeadObjectInput{
		Bucket: aws.String(thumbnailBucket),
		Key:    aws.String(image.Name),
	})
	if err != nil {
		fmt.Println("error deleteing object from bucket")
		return nil, err
	}

	// delete fullsize from s3
	_, err = s3client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(fullsizeBucket),
		Key:    aws.String(image.Name),
	})
	if err != nil {
		fmt.Println("error deleting object from bucket")
		return nil, err
	}
	err = s3client.WaitUntilObjectNotExists(&s3.HeadObjectInput{
		Bucket: aws.String(fullsizeBucket),
		Key:    aws.String(image.Name),
	})
	if err != nil {
		fmt.Println("error deleteing object from bucket")
		return nil, err
	}

	// delete from db
	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"_id": {
				S: aws.String(id),
			},
		},
		TableName: aws.String("Images"),
	}

	_, err = dbclient.DeleteItem(input)
	if err != nil {
		fmt.Println("Got error calling DeleteItem")
		return nil, err
	}
	returnImage := model.Image{
		ID:           image.ID,
		Name:         image.Name,
		Price:        image.Price,
		ThumbnailURL: "",
		FullsizeURL:  "",
	}
	return &returnImage, nil
}

func (r *mutationResolver) CreateCheckoutSession(ctx context.Context, photoID string) (string, error) {
	var domain string
	if os.Getenv("HEROKU_ENV") == "PROD" {
		domain = "https://protected-bastion-36826.herokuapp.com"
	} else {
		domain = "http://localhost:8080"
	}

	// get image name from db
	result, err := dbclient.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("Images"),
		Key: map[string]*dynamodb.AttributeValue{
			"_id": {S: aws.String(photoID)},
		},
	})
	if err != nil || result.Item == nil {
		fmt.Println("error getting item")
		return "", err
	}
	image := model.ImageDB{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &image)

	// get fullsize download url
	/***
	req_old, _ := s3client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(fullsizeBucket),
		Key:    aws.String(image.Name),
	})
	fullsizeURL, err := req_old.Presign(5 * time.Minute)
	if err != nil {
		log.Println("Failed to sign request", err)
	}
	***/

	// get LR Preset download url
	presetFileName := image.Name
	req, _ := s3client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(presetsBucket),
		Key:    aws.String(presetFileName),
	})
	presetURL, err := req.Presign(5 * time.Minute)
	if err != nil {
		log.Println("Failed to sign request", err)
	}

	// create checkout session
	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			&stripe.CheckoutSessionLineItemParams{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String(string(stripe.CurrencyCAD)),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String(image.Name),
					},
					UnitAmountDecimal: stripe.Float64(math.Round(image.Price * 100)),
				},
				Quantity: stripe.Int64(1),
			},
		},
		Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
		// SuccessURL: stripe.String(domain + "/success?downloadurl=" + hex.EncodeToString([]byte(fullsizeURL))),
		SuccessURL: stripe.String(domain + "/success?downloadurl=" + hex.EncodeToString([]byte(presetURL))),
		CancelURL:  stripe.String(domain + "/cancel"),
	}
	sess, err := stripeSession.New(params)
	if err != nil {
		log.Printf("stripeSession.New: %v", err)
	}
	return sess.ID, nil
}

func (r *queryResolver) Image(ctx context.Context, id string) (*model.Image, error) {
	// get image name from db
	result, err := dbclient.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("Images"),
		Key: map[string]*dynamodb.AttributeValue{
			"_id": {S: aws.String(id)},
		},
	})
	if err != nil || result.Item == nil {
		fmt.Println("error getting item")
		return nil, err
	}
	image := model.ImageDB{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &image)

	// get image urls
	req, _ := s3client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(thumbnailBucket),
		Key:    aws.String(image.Name),
	})
	thumbnailURL, err := req.Presign(5 * time.Minute)
	if err != nil {
		log.Println("Failed to sign request", err)
	}

	req, _ = s3client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(fullsizeBucket),
		Key:    aws.String(image.Name),
	})
	fullsizeURL, err := req.Presign(5 * time.Minute)
	if err != nil {
		log.Println("Failed to sign request", err)
	}

	ret := model.Image{
		ID:           id,
		Name:         image.Name,
		Price:        1.00,
		ThumbnailURL: thumbnailURL,
		FullsizeURL:  fullsizeURL,
	}
	return &ret, nil
}

func (r *queryResolver) Images(ctx context.Context) ([]*model.Image, error) {
	params := &dynamodb.ScanInput{
		TableName: aws.String("Images"),
	}
	result, err := dbclient.Scan(params)
	if err != nil {
		fmt.Println("error while scanning table")
		return nil, err
	}
	obj := []*model.ImageDB{}
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &obj)
	if err != nil {
		fmt.Println("error while unmarshaling list of maps")
	}
	ret := []*model.Image{}
	for _, v := range obj {
		image := model.Image{
			ID:    v.ID,
			Name:  v.Name,
			Price: v.Price,
		}

		// get image urls
		thumbnailName := image.Name[:2] + "-after.jpg"
		req, _ := s3client.GetObjectRequest(&s3.GetObjectInput{
			Bucket: aws.String(thumbnailBucket),
			Key:    aws.String(thumbnailName),
		})
		thumbnailURL, err := req.Presign(5 * time.Minute)
		if err != nil {
			log.Println("Failed to sign request", err)
		}

		req, _ = s3client.GetObjectRequest(&s3.GetObjectInput{
			Bucket: aws.String(fullsizeBucket),
			Key:    aws.String(image.Name),
		})
		fullsizeURL, err := req.Presign(5 * time.Minute)
		if err != nil {
			log.Println("Failed to sign request", err)
		}
		image.ThumbnailURL = thumbnailURL
		image.FullsizeURL = fullsizeURL

		ret = append(ret, &image)
	}

	return ret, nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
