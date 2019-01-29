# Asset/file cloud storage api

Technology used:
Golang
AWS S3
Dynamodb
deployed on an ECS Fargate cluster

I decided to implement token-based access as my additional requirement.

I also attempted to handle orchestration via terraform, got a long way in, and had an issue with my ELB (I think) not routing properly
and decided to focus on getting it deployed manually on a new cluster, and focusing on documentation instead.

### Building and running local

Note: You'll need go 11.1.x or newer to use the following instructions. You also will need to be logged into the
 aws account via aws cli, and set up a dynamodb table with string fields PK ObjID and SK ObjSort, and an s3 bucket.  

I've provided a makefile useful for building.

```make all```  
And then:  
```DYNAMODB_TABLE={dynamodb_table_name} S3_BUCKET={s3_bucket_name} PORT={port} ./main```  
or export those env vars if you'd prefer and then just ```./main```

Testing:  
```make test```

Accessing the hosted version: 

Project is currently hosted at ```http://lienmeat-lb-1721212180.us-west-2.elb.amazonaws.com```

### Endpoints:

To add an asset/file:

Fields:  
token = 1 designates you wish to generate a token to access the file  
expiry = # of minutes the token is valid from the request time

POST /asset/:assetname?token=1&expiry=20
Where the request body is the file data.

```
curl -i -X POST \
   -H "Content-Type:text/csv" \
   -T "./something.txt" \
 'http://lienmeat-lb-1721212180.us-west-2.elb.amazonaws.com/asset/something.txt?token=1&expiry=200'
```
 
or, multipart/form-encoded:
POST /asset

```
curl -i -X POST \
   -H "Content-Type:multipart/form-data" \
   -F "token=1" \
   -F "expiry=20" \
   -F "file=@\"./something.txt\";type=text/csv;filename=\"something.txt\"" \
 'http://lienmeat-lb-1721212180.us-west-2.elb.amazonaws.com/asset'
```

#### Example response, with token generated:

```
{
    "asset":{
        "id": "6b84149d-332c-4152-bb73-0ca9da463eaf", //access via id, never expires
        "name": "something.txt",
        "size": 84,
        "version": 0 //I wanted to implement asset versioning via DynamoDB SortKey column usage, but decided I didn't have time
    },
    "token":{
        "token": "405ae415-3c44-487c-8024-4294f2d4c680",    //token to access the asset
        "expiry": 1548663712,                               //unix timestamp of expiry, UTC
        "asset_id": "6b84149d-332c-4152-bb73-0ca9da463eaf"  //id of asset the token corresponds to
    },
    "error": null                                           //any err that occured durring the request
}
```

To get a asset/file:

GET /asset/{asset_id}

```
curl -i -X GET \
 'http://lienmeat-lb-1721212180.us-west-2.elb.amazonaws.com/asset/6b84149d-332c-4152-bb73-0ca9da463eaf'
```

GET /asset-token/{token}

```
curl -i -X GET \
 'http://lienmeat-lb-1721212180.us-west-2.elb.amazonaws.com/asset-token/405ae415-3c44-487c-8024-4294f2d4c680'
```

You will get a HTTP 204 if either the resource doesn't exist or the token is expired, or a 200 & file download otherwise.


## Technical Decisions:

 I used aws S3 for the asset/file data storage, and dynamodb to keep track of metadata and tokens.
 
 S3 for large file sizes, DynamoDB for fast lookups that scale well in number if you structure your table's Partition
 and Sort keys well for your use case.  DynamoDB alone was not possible because of the limited size of dynamodb records.
 
 However, S3 was not good enough alone, because it would be too slow to scan via ListObjects() or
 similar for metadata if you were to implement user-based-access, tokens, or really anything interesting as the number
 of files grew.  I strove to structure the table so that I did not need a Global Secondary Index, while duplicating as
 little data as possible.  The end result was:  
 ![schema](https://www.dropbox.com/s/dl/jdf61bgks49x8lf/dynamodbtable.png "schema")  
 Where ObjID was a unique id for each object type, and ObjSort some value to sort by, or use as an overloaded Global Secondary
 Index if the application/featureset needed that.  
 
 This design would have allowed me to add many more features on top of these without a lot more effort.  I could have added
 user-owned files, listed files owned by a user somewhat easily without degrading performance of lookups. 
 
One of the best aspects of my implementation is that because it passes the io.ReadCloser that is the request.Body directly
through to the s3.PutObject() body, which is also an io.ReadCloser, there is VERY little memory usage, as it's streaming
the upload straight to s3.  Similarly, it streams it to the client when they GET it.  
 
 I used uuid V4 for both asset ids and tokens.  I don't think this is definitely the best choice as a shorter id or token
 would be better for sharing or handling without typos, but I think it works well as a proof of concept.
 
 I might have went a bit overboard on the interfaces in assetstore.go I think, but small interfaces in go can often 
 make unexpected design changes down the road much easier.  I've found going overboard is generally better than having
 too few interfaces, or too big of interfaces.
 
 It's hosted on AWS Fargate because I like docker and Fargate is pretty easy to scale and set up.  I considered lambda, but the
 downsides using api gateway and timeouts for long-running processes like a large file upload would be a disastrous limitation.
 
 As I stated previously, I was going to use terraform to orchestrate the service, but I ran out of time when I hit an issue
 getting the ELB or perhaps the gateway/nat to route properly.  Too bad, I was getting close.  I left the "tf" stuff there
 because I figure it shows I was on the right track.
 
 
