# Photography Ecommerce Shop

This is my submission for the 2021 Shopify Backend Challenge.
The current repo contains the backend for the challenge, 
and I've made a [UI as well which I'm also using as my photography portfolio](https://github.com/xngln/photo-store-client).

The GraphQL server, written in Golang, supports:
1. Uploading new images. The metadata (name, id, price) is stored in AWS DynamoDB and the files themselves are stored in AWS S3 buckets.
2. Deleting images by ID.

These two features are mainly for the store owner to manage which images they want to make available for sale on the store.

Next, the server also supports:

3. Retrieving a single image by ID.
4. Retrieving all images.

These endpoints are consumed by the UI to display the images in the shop.

And lastly, there is also an endpoint for:

5. Creating a Stripe checkout session. This is also only used by the UI to create a new Stripe checkout session to start the checkout process once the customer has decided to pay.

The schemas for the endpoints, queries, and mutations can be found in `graph/schema.graphqls`.

The resolvers for all queries and mutations are implemented in `graph/schema.resolvers.go`.

The server, hosted on Heroku, exposes one GraphQL endpoint at `https://protected-bastion-36826.herokuapp.com/query`. 
You can also visit `https://protected-bastion-36826.herokuapp.com/` to use the GraphQL playground. Since the project is on the free tier of Heroku, the first request may take up to 30 seconds to complete.

## Example Requests 
Adding new image to shop:
```bash
curl https://protected-bastion-36826.herokuapp.com/query \
  -F operations='{ "query": "mutation ($input: NewImage!) { uploadImage(input: $input) { _id, name, price }}", "variables": { "input": {"name": "test.jpg", "price": 1, "file": null } } }' \
  -F map='{ "0": ["variables.input.file"] }' \
  -F 0=@./testimages/test.jpg
```

Deleting image by ID (using the GraphQL playground):
```graphql
mutation deleteImage($id: String!) {
  deleteAuthor(id: $id) {
     _id
     name
  }
}
```

Creating a new Stripe checkout session using product(image) ID:
```graphql
mutation createCheckoutSession {
    createCheckoutSession(photoID: "7018c40f-ad75-4d23-b1a8-33eb051f8e1f")
}
```

Query for single image by ID:
```graphql
query {
  image(_id: "7018c40f-ad75-4d23-b1a8-33eb051f8e1f") {
    _id
    price	
    name
    thumbnail_url
    fullsize_url
  }
}
```

## Shop
To checkout out the shop, go to [photo.davidxliu.com](https://www.photo.davidxliu.com). The code can be viewed [here](https://github.com/xngln/photo-store-client). This is the homepage of the portfolio. In the header, select *shop*.
This will display all the images which are currently available to purchase for download. 
Clicking the prompt to buy an image will ask the server to start a new Stripe checkout session and redirect you to the checkout page. 
Sadly, a "cart" feature hasn't been implemented yet, so one can only buy a single image at a time. The shop is still in test mode, so you can try 
the following credit card to get a successful purchase:

> Card #: 4242 4242 4242 4242. Name, address, cvv can be anything. Expiry date must be any future date. 

After the payment is processed, you will be redirected to the success page, from which you can download your purchased high res image file :) 
This download url is actually a pre-signed URL to access the object from its S3 bucket, so it is only available for a few minutes after the payment is completed. 
