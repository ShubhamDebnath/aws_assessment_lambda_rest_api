package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
)

type CupcakeDoc struct {
	Count       int    `json:"count"`
	Date        string `json:"date"`
	EvtType     string `json:"evtType"`
	Last_update string `json:"last_update"`
}

var sess *session.Session

// Basic information for the Amazon Elasticsearch Service domain
var domain = "https://search-cupcake-domain-001-bdj74ottahj7ttzw3szp4e6tea.ap-south-1.es.amazonaws.com"
var index = "cupcake-index-002"
var endpoint = domain + "/" + index + "/" + "_search"
var region = "ap-south-1" // e.g. us-east-1
var service = "es"
var signer *v4.Signer
var client = &http.Client{}

// this runs only the very first time
// this lambda function is created
// will be used for initial setup
func init() {
	sess = configureAWS()
	signer = v4.NewSigner(sess.Config.Credentials)
}

func configureAWS() *session.Session {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("ap-south-1"),
		Credentials: credentials.NewStaticCredentials("XXXXXXXXX", "XXXXXXXX", ""),
	})

	if err != nil {
		log.Fatal(err)
	}

	// adding logging handler to session
	sess.Handlers.Send.PushFront(func(r *request.Request) {
		// Log every request made and its payload
		fmt.Printf("Request: %s/%s, Payload: %s",
			r.ClientInfo.ServiceName, r.Operation, r.Params)
	})

	// fmt.Println(sess)

	return sess
}

func show(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	searchReqParams := `{
		"size": 200,
		"sort": { "Month": "asc", "last_update": "asc"},
		"query": {
		   "match_all": {}
		}
	 }`
	body := strings.NewReader(searchReqParams)

	searchReq, err := http.NewRequest(http.MethodPost, endpoint, body)
	if err != nil {
		fmt.Print(err)
	}

	searchReq.Header.Add("Content-Type", "application/json")

	signer.Sign(searchReq, body, service, region, time.Now())
	resp, err := client.Do(searchReq)

	if err != nil {
		fmt.Print(err)
	}

	var dataMap map[string]interface{}

	// creating map out of response body
	// will use selective fields out of this
	err = json.NewDecoder(resp.Body).Decode(&dataMap)

	if err != nil {
		fmt.Println(err.Error())
	}

	respSlice := make([]CupcakeDoc, 0)

	// Iterate the document "hits" returned by API call
	for _, hit := range dataMap["hits"].(map[string]interface{})["hits"].([]interface{}) {

		// Parse the attributes/fields of the document
		doc := hit.(map[string]interface{})
		source := doc["_source"].(map[string]interface{})
		// fmt.Println("doc _source:", reflect.TypeOf(source))

		// // Get the document's _id and print it out along with _source data
		// docID := doc["_id"]
		// fmt.Println("docID:", docID)
		// // fmt.Println("_source:", source, "\n")

		count, _ := strconv.Atoi(source["Cupcake"].(string))
		respSlice = append(respSlice, CupcakeDoc{Count: count, Date: source["Month"].(string), EvtType: source["evtType"].(string), Last_update: source["last_update"].(string)})

	} // end of response iteration

	// The APIGatewayProxyResponse.Body field needs to be a string, so
	// we marshal the book record into JSON.
	js, err := json.Marshal(respSlice)
	if err != nil {
		return serverError(err)
	}

	// Return a response with a 200 OK status and the JSON book record
	// as the body.
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(js),
		Headers: map[string]string{"Access-Control-Allow-Headers": "Content-Type",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "*"},
	}, nil
}

// Add a helper for handling errors. This logs any error to os.Stderr
// and returns a 500 Internal Server Error response that the AWS API
// Gateway understands.
func serverError(err error) (events.APIGatewayProxyResponse, error) {
	fmt.Println(err.Error())

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusInternalServerError,
		Body:       http.StatusText(http.StatusInternalServerError),
		Headers: map[string]string{"Access-Control-Allow-Headers": "Content-Type",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "*"},
	}, nil
}

// Similarly add a helper for send responses relating to client errors.
func clientError(status int) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: status,
		Body:       http.StatusText(status),
		Headers: map[string]string{"Access-Control-Allow-Headers": "Content-Type",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "*"},
	}, nil
}

func main() {

	lambda.Start(show)

}
