# AWS Image Library Example #

# Configure #

1. In the AWS Console, in your preferred region, create:  
  * An S3 bucket.
  * A DynamoDB table:
    * For the table name, use `Images`.
    * For a primary partition key, use `S3ObjectKey` (String).
    * Add a secondary index and, as a primary key, use the partition key `Category` (String). (The console will auto-generate the index name "Category-index"; use this, as the code expects it.)
2. Go to config.js and configure the names of your bucket and AWS region.

# Run #

```
npm start
```

# Test #

We recommend using a REST client app to test your service. We like [Postman](https://www.getpostman.com/).

Alternatively, you can use [curl](https://curl.haxx.se/) to test endpoints.

To post to the /images endpoint in your app, you'd do something like this:
```
curl -F photo=@/Full/path/to/local/image.jpg http://localhost:3000/images?category=my_category
```
To issue a GET request, you could do something like this:
```
curl http://localhost:3000/images?category=my_category
```
