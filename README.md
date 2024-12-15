
# Portfolio WebApp Deployment Using CloudFormation SAM

This project deploys a Lambda function that stores visitor logs in DynamoDB, an API Gateway to handle incoming HTTP requests, and an S3 bucket for static website hosting. The deployment is done using AWS SAM (Serverless Application Model) and the resources are managed with AWS CloudFormation.

## Architecture:
- Deploy Static Portfolio website to S3
- Create API Gateway to connect Portfolio and Lambda
- Deploy Container Image of Go App on AWS Lambda
- Create DynamoDB table to store Visitor-Logs

## Prerequisites

1. **AWS CLI**: Make sure the AWS Command Line Interface (CLI) is installed and configured with your credentials.
   
   ```bash
   aws configure
   ```

2. **Docker**: Required for building and pushing the Lambda container image.

3. **AWS SAM CLI**: To deploy and manage the serverless application.
   - Install SAM CLI: https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/serverless-prereqs.html

4. **Go**: Ensure that Go is installed to compile and build the Go application.

## Deployment Steps

### Step 1: Clone the Repository

Clone the repository to your local machine.

```bash
git https://github.com/Fazal-Rehaman07/Portfolio_CF_Deploy.git
cd Portfolio_CF_Deploy
```

### Step 2: Build the Go Application

The Go application will be packaged into a Docker container, which will then be deployed to AWS Lambda.

1. Build and push the Docker image to Amazon ECR.

   - Create an ECR repository if not already created:

     ```bash
     aws ecr create-repository --repository-name visitor-logs --region us-east-1
     ```

   - Authenticate Docker to AWS ECR:

     ```bash
     aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin <Account_ID>.dkr.ecr.us-east-1.amazonaws.com
     ```

   - Build and tag the Docker image:

     ```bash
     docker build --platform linux/amd64 -t visitor-logs .
     ```

   - Push the Docker image to ECR:

     ```bash
     docker tag visitor-logs:latest <Account_ID>.dkr.ecr.us-east-1.amazonaws.com/visitor-logs:latest
     docker push <Account_ID>.dkr.ecr.us-east-1.amazonaws.com/visitor-logs:latest
     ```

### Step 3: Deploy the SAM Template

Now that the Lambda function image is in ECR, we can deploy the infrastructure using AWS SAM.

1. **Build the SAM application**:

   ```bash
   sam build
   ```

2. **Deploy the SAM application**:

   ```bash
   sam deploy --guided
   ```

   Follow the prompts:
   - Choose the stack name.
   - Select the AWS region (e.g., `us-east-1`).
   - Allow SAM to create IAM roles for Lambda execution.

   This will deploy the API Gateway, DynamoDB, Lambda function, and S3 bucket.

### Step 4: Update the Frontend Code

Update the frontend code (index.html) to use the new Lambda API endpoint:

In your frontend code, locate the API call to the Lambda endpoint, and update it to the new API endpoint provided by API Gateway.

Replace the existing API URL with the newly deployed endpoint (you can find this in the SAM deployment outputs).

For example:

```javascript
fetch('https://<your API Gateway Endpoint>', {  
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({IP: data.ip}),
                mode: 'no-cors'
            });
```

### Step 5: Test the Deployment

Once the stack is deployed successfully, you can test the static website hosted on the S3 bucket. To access the website, visit the following URL:

```
http://<your-s3-bucket-name>.s3-website-us-east-1.amazonaws.com
```

This will show the Portfolio website hosted in S3 and it will save the user IP and Timestamp in DynamoDB.

Additionally, you can test the Lambda function by sending a POST request to the API Gateway endpoint.

Use tools like `curl` or Postman to send a POST request to the API Gateway.

Example:

```bash
curl -X POST https://<api-gateway-id>.execute-api.us-east-1.amazonaws.com/Prod/visitor \
     -H "Content-Type: application/json" \
     -d '{"IP": "192.168.1.1"}'
```

This should store the visitor log in DynamoDB and return a success message.


## Go Lambda Function Explanation

The Go Lambda function receives an event (the visitor log) from API Gateway, processes it, and stores it in DynamoDB. Here's how it works:

### 1. **Imports**:
   - `context`: Allows the Lambda function to receive a context object for managing timeouts and cancellations.
   - `encoding/json`: Used for encoding and decoding JSON data.
   - `fmt`: Standard library for formatted I/O.
   - `log`: Standard library for logging.
   - `time`: Used to add a timestamp to the log entry.
   - `github.com/aws/aws-lambda-go/lambda`: AWS SDK for Go Lambda, used to start the Lambda function.
   - `github.com/aws/aws-sdk-go/aws` and `github.com/aws/aws-sdk-go/service/dynamodb`: AWS SDK for Go to interact with DynamoDB.

### 2. **VisitorLog Struct**:
   This struct defines the structure of the log data (IP address and timestamp).
   
   ```go
   type VisitorLog struct {
       IP        string `json:"IP"`
       Timestamp string `json:"Timestamp"`
   }
   ```

### 3. **Global Variable**:
   `dynamoClient` is initialized to interact with DynamoDB.

   ```go
   var dynamoClient *dynamodb.DynamoDB
   ```

### 4. **init() Function**:
   This function initializes the DynamoDB client when the Lambda function is loaded.

   ```go
   func init() {
       sess := session.Must(session.NewSession(&aws.Config{
           Region: aws.String("us-east-1"),
       }))
       dynamoClient = dynamodb.New(sess)
   }
   ```

### 5. **Handler Function**:
   The `handler` function processes the incoming event:

   - Decodes the incoming JSON body to populate the `VisitorLog` struct.
   - Adds a timestamp to the log.
   - Marshals the `VisitorLog` struct into a DynamoDB item format.
   - Stores the log in DynamoDB using the `PutItem` API.
   - Returns a response with a success message and CORS headers.

   ```go
   func handler(ctx context.Context, event map[string]interface{}) (map[string]interface{}, error) {
       var visitor VisitorLog
       body, _ := json.Marshal(event)
       err := json.Unmarshal(body, &visitor)
       if err != nil {
           log.Printf("Error decoding JSON: %v", err)
           return nil, fmt.Errorf("Invalid request body")
       }

       visitor.Timestamp = time.Now().Format("2006-01-02 15:04:05")
       item, err := dynamodbattribute.MarshalMap(visitor)
       if err != nil {
           log.Printf("Failed to marshal visitor log: %v", err)
           return nil, fmt.Errorf("Internal server error")
       }

       input := &dynamodb.PutItemInput{
           TableName: aws.String("VisitorLogs"),
           Item:      item,
       }
       _, err = dynamoClient.PutItem(input)
       if err != nil {
           log.Printf("Failed to put item in DynamoDB: %v", err)
           return nil, fmt.Errorf("Internal server error")
       }

       return map[string]interface{}{
           "statusCode": 200,
           "body":       "Visitor log stored successfully",
           "headers": map[string]string{
               "Access-Control-Allow-Origin":  "*",
               "Access-Control-Allow-Methods": "GET, POST, OPTIONS",
               "Access-Control-Allow-Headers": "Content-Type",
           },
       }, nil
   }
   ```

### 6. **Main Function**:
   This starts the Lambda function.

   ```go
   func main() {
       log.Println("Lambda Execution Started")
       lambda.Start(handler)
   }
   ```

### 7. **CORS Headers**:
   The response includes CORS headers to allow cross-origin requests from the frontend.

   ```go
   "headers": map[string]string{
       "Access-Control-Allow-Origin":  "*",                  
       "Access-Control-Allow-Methods": "GET, POST, OPTIONS",
       "Access-Control-Allow-Headers": "Content-Type",
   },
   ```

## Conclusion

This project demonstrates how to deploy a serverless application using AWS Lambda, API Gateway, DynamoDB, and S3 using AWS SAM. The Go handler function processes visitor logs and stores them in DynamoDB, while the API Gateway serves as the endpoint for submitting logs.


