This repo is responsible for the go code which is uploaded for lambda function
to act as a REST API backend, all it does is, wait for a get request from AWS API gateway
then send request to ElasticSearch with a search query
And then send the same search results as response to whoever called this endpoint

For now the search query is pretty simple

```
{
		"size": 200,
		"sort": { "Month": "asc", "last_update": "asc"},
		"query": {
		   "match_all": {}
		}
}
```

can easily be modified if requirements are changed

Also since this API is running in AWS_PROXY mode, to enable CORS
I just send these as response headers

```
{
    "Access-Control-Allow-Headers": "Content-Type",
	"Access-Control-Allow-Origin":  "*",
	"Access-Control-Allow-Methods": "*"
}
```

![react-app](https://github.com/ShubhamDebnath/aws_assessment_lambda_rest_api/blob/main/react-app.JPG)

Commands used

for creating the lambda function, which is hosting the rest api,

```
aws lambda create-function --function-name cupcakeAPI
--runtime go1.x --zip-file fileb://main.zip --handler main --role arn:aws:iam::234825976347:role/service-role/892905-lambda-role
```

1. Create rest api

```
aws apigateway create-rest-api --name cupcakeRestAPI
```

```
{
    "id": "v9s1jkzlu3",
    "createdDate": "2021-04-04T21:02:03+05:30",
    "apiKeySource": "HEADER",
    "endpointConfiguration": {
        "types": [
            "EDGE"
        ]
    },
    "disableExecuteApiEndpoint": false
}
```

2. Get root API resource Id

```
aws apigateway get-resources --rest-api-id v9s1jkzlu3
```

```
{
    "items": [
        {
            "id": "3o2xh8rp04",
            "path": "/"
        }
    ]
}
```

3. Create new resource under the root path

```
aws apigateway create-resource --rest-api-id v9s1jkzlu3 --parent-id 3o2xh8rp04 --path-part records
```

```
{
    "id": "y2c689",
    "parentId": "3o2xh8rp04",
    "pathPart": "records",
    "path": "/records"
}
```

4. Allow ANY https methods on previously created resource

```
aws apigateway put-method --rest-api-id v9s1jkzlu3 --resource-id y2c689 --http-method ANY --authorization-type NONE
```

```
{
    "httpMethod": "ANY",
    "authorizationType": "NONE",
    "apiKeyRequired": false
}
```

5. Integrate with already created lambda function

```
aws apigateway put-integration --rest-api-id v9s1jkzlu3 --resource-id y2c689 --http-method ANY --type AWS_PROXY --integration-http-method POST --uri arn:aws:apigateway:ap-south-1:lambda:path/2015-03-31/functions/arn:aws:lambda:ap-south-1:234825976347:function:cupcakeAPI/invocations
```

```
{
    "type": "AWS_PROXY",
    "httpMethod": "POST",
    "uri": "arn:aws:apigateway:ap-south-1:lambda:path/2015-03-31/functions/arn:aws:lambda:ap-south-1:234825976347:function:cupcakeAPI/invocations",
    "passthroughBehavior": "WHEN_NO_MATCH",
    "timeoutInMillis": 29000,
    "cacheNamespace": "y2c689",
    "cacheKeyParameters": []
}
```

6. Give the API to access the lambda function beneath

```
aws lambda add-permission --function-name cupcakeAPI --statement-id cupcake-api-stmt-001 --action lambda:InvokeFunction --principal apigateway.amazonaws.com --source-arn arn:aws:execute-api:ap-south-1:234825976347:v9s1jkzlu3/*/*/*
```

```
{
    "Statement": "{\"Sid\":\"cupcake-api-stmt-001\",\"Effect\":\"Allow\",\"Principal\":{\"Service\":\"apigateway.amazonaws.com\"},\"Action\":\"lambda:InvokeFunction\",\"Resource\":\"arn:aws:lambda:ap-south-1:234825976347:function:cupcakeAPI\",\"Condition\":{\"ArnLike\":{\"AWS:SourceArn\":\"arn:aws:execute-api:ap-south-1:234825976347:v9s1jkzlu3/*/*/*\"}}}"
}
```

Testing the created API gateway (will fail the first time, for that need to add persions to lambda function)

```
aws apigateway test-invoke-method --rest-api-id v9s1jkzlu3 --resource-id y2c689 --http-method "GET"
```

7. If the test for api gateway is successful, deploy the api

```
aws apigateway create-deployment --rest-api-id v9s1jkzlu3 --stage-name staging
```

```
{
    "id": "01f7m4",
    "createdDate": "2021-04-04T21:28:54+05:30"
}
```

Finally access the following endpoint to see fruits of your labour

```
https://v9s1jkzlu3.execute-api.ap-south-1.amazonaws.com/staging/records
```
