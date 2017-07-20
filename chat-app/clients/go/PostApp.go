/*  Copyright 2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License").
 *  You may not use this file except in compliance with the License.
 *  A copy of the License is located at
 *
 *  http://aws.amazon.com/apache2.0/
 */

package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
)

type Configuration struct {
    Region      string
    Timezone    string
    MaxMessages int
    RefreshSeconds int
    Debug bool
}

// Configuration
var configuration Configuration

func SetConfiguration() error {
    var myError error

    if configuration == (Configuration{}) {
        // Get configuration values
        file, _ := os.Open("conf.json")
        decoder := json.NewDecoder(file)

        err := decoder.Decode(&configuration)

        if err != nil {
			// Set the values to something reasonable
			configuration.Region = "us-west-2"
			configuration.Timezone = "UTC"
			configuration.MaxMessages = 20
            configuration.RefreshSeconds = 30
            configuration.Debug = false

            myError = errors.New("Error parsing config file: " + err.Error())
            return myError
        }
    }

    return myError
}

var client *lambda.Lambda

func getLambdaClient() (*lambda.Lambda)  {
	if client == nil { // *(lambda.Lambda{}) {
		// Create Lambda service client
		sess := session.Must(session.NewSessionWithOptions(session.Options{
			SharedConfigState: session.SharedConfigEnable,
		}))

		client = lambda.New(sess, &aws.Config{Region: aws.String(configuration.Region)})
	}

	return client
}

func clearScreen() {
	switch runtime.GOOS {
	case "linux":
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()

	case "windows":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

// Global log
var Debug *log.Logger

func initLog(debugHandle io.Writer) {
	Debug = log.New(debugHandle, "", 0) // To add info about date/time:
	//		"DEBUG: ",
	//		log.Ldate|log.Ltime|log.Lshortfile)
}

type FormatAsTime time.Time

func (t FormatAsTime) String() string {
	const format = `3:04:05 PM MST`
	return time.Time(t).Format(format)
}

type FormatAsDate time.Time

func (t FormatAsDate) String() string {
	const format = `Monday, January _2 2006`
	return time.Time(t).Format(format)
}

func (t FormatAsDate) Equals(t2 FormatAsDate) bool {
	return t.String() == t2.String()
}

type getPostsRequest struct {
	SortBy     string
	SortOrder  string
	PostsToGet int
}

type getPostsResponseAlias struct {
	S string
}

type getPostsResponseTimestamp struct {
	S string
}

type getPostsResponseMessage struct {
	S string
}

type getPostsResponseData struct {
	Alias     getPostsResponseAlias
	Timestamp getPostsResponseTimestamp
	Message   getPostsResponseMessage
}

type getPostsResponseBody struct {
	Result string                 `json:"result"`
	Data   []getPostsResponseData `json:"data"`
}

type getPostsResponseHeaders struct {
	ContentType string `json:"Content-Type"`
}

type getPostsResponse struct {
	StatusCode int                     `json:"statusCode"`
	Headers    getPostsResponseHeaders `json:"headers"`
	Body       getPostsResponseBody    `json:"body"`
}

func getAllPosts(maxMessages int) (getPostsResponse, error) {
	svc := getLambdaClient()

	var myError error
	var resp getPostsResponse

	// Get the latest maxMessages posts
	request := getPostsRequest{"timestamp", "descending", maxMessages}

	payload, err := json.Marshal(request)

	if err != nil {
		myError = errors.New("Error marshalling GetPosts request")
		return resp, myError
	}

	result, err := svc.Invoke(&lambda.InvokeInput{FunctionName: aws.String("GetPosts"), Payload: payload})

	if err != nil {
		myError = errors.New("Error calling GetPosts")
		return resp, myError
	}

	Debug.Println("")
	Debug.Println("Raw response:")
	Debug.Println(string(result.Payload))
	Debug.Println("")

	err = json.Unmarshal(result.Payload, &resp)

	if err != nil {
		myError = errors.New("Error unmarshalling GetPosts response")
		return resp, myError
	}

	return resp, myError
}

func listAllPosts(resp getPostsResponse) error {
	var myError error

	// Check the status code
	if resp.StatusCode != 200 {
		myError = errors.New("Error getting posts, StatusCode: " + strconv.Itoa(resp.StatusCode))
		return myError
	}

	numPosts := len(resp.Body.Data)

	if numPosts > 0 {
		var origDate FormatAsDate

		msg := fmt.Sprintf("%s %d %s", "Got", numPosts, "posts:\n")
		// WAS: debugPrint(debug, msg)
		Debug.Println(msg)

		for i := range resp.Body.Data {
			p := resp.Body.Data[len(resp.Body.Data)-i-1]
			// Doug @ 4:45 PM PST <ID>:
			// Where is the meeting today?

			// Convert date/time from UTC
			numTime, err := strconv.ParseInt(p.Timestamp.S, 10, 64)

			if err == nil {
				thisTime := time.Unix(numTime, 0)

				theDate := FormatAsDate(thisTime)
				theTime := FormatAsTime(thisTime)

				// If we have a new date, show it
				if !origDate.Equals(theDate) {
					fmt.Println("=== " + theDate.String() + " ===")
					fmt.Println("")

					origDate = theDate
				}

				fmt.Println(p.Alias.S + "@" + theTime.String() + " <" + p.Timestamp.S + ">:")
				fmt.Println(p.Message.S)
				fmt.Println("")
			} else {
				fmt.Println(p.Alias.S + "@??? <" + p.Timestamp.S + ">:")
				fmt.Println(p.Message.S)
				fmt.Println("")
			}
		}
	}

	return myError
}

type user struct {
	UserName string
	Password string
}

type userResponseHeaders struct {
	ContentType string `json:"Content-Type"`
}

type userResponseAuthenticationResult struct {
	AccessToken  string
	ExpiresIn    int
	TokenType    string
	RefreshToken string
	IdToken      string
}

type userResponseData struct {
	ChallengeParameters  interface{} // {} is always returned
	AuthenticationResult userResponseAuthenticationResult
}

type userResponseBody struct {
	Result string           `json:"result"`
	Data   userResponseData `json:"data"`
}

type userResponseBodyFailureError struct {
    Message string `json:"message"`
}

type userResponseBodyFailure struct {
	Result string `json:"result"`
	Error  userResponseBodyFailureError `json:"error"`
}

type userResponse struct {
	StatusCode int                 `json:"statusCode"`
	Headers    userResponseHeaders `json:"headers"`
	Body       userResponseBody    `json:"body"`
}

type userResponseFailure struct {
	StatusCode int                     `json:"statusCode"`
	Headers    userResponseHeaders     `json:"headers"`
	Body       userResponseBodyFailure `json:"body"`
}

func signInUser(userName string, password string) (string, error) {
	svc := getLambdaClient()

	var myError error

	Debug.Println("Creating payload for user " + userName)

	user := user{userName, password}

	payload, err := json.Marshal(user)

	if err != nil {
		myError = errors.New("Error marshalling SignInCognitoUser request")
		return "", myError
	}

	Debug.Println("Signing in user:")
	Debug.Println(string(payload))

	result, err := svc.Invoke(&lambda.InvokeInput{FunctionName: aws.String("SignInCognitoUser"), Payload: payload})

	if err != nil {
		myError = errors.New("Error calling SignInCognitoUser: " + err.Error())
		return "", myError
	}

	Debug.Println("")
	Debug.Println("Raw response:")
	Debug.Println(string(result.Payload))
	Debug.Println("")

	var resp userResponse
	err = json.Unmarshal(result.Payload, &resp)

	if err != nil {
		myError = errors.New("Error unmarshalling SignInCognitoUser response")
		return "", myError
	}

	// Did we not get a 200?
	if resp.StatusCode != 200 {
        var respFailure userResponseFailure
        err = json.Unmarshal(result.Payload, &respFailure)

        if err != nil {
            myError = errors.New("Got status code: " + strconv.Itoa(resp.StatusCode) + " trying to sign in user")
            return "", myError
        } else {
            myError = errors.New("Got error: " + respFailure.Body.Error.Message + " trying to sign in user")
            return "", myError
        }
	}

	if resp.Body.Result != "success" {
		var respFailure userResponseFailure

		err = json.Unmarshal(result.Payload, &respFailure)

		if err != nil {
			myError = errors.New("Error unmarshalling SignInCognitoUser response")
			return "", myError
		}

		if respFailure.Body.Error.Message != "" {
			myError = errors.New("Failed to sign in Cognito user. Error: " + respFailure.Body.Error.Message)
			return "", myError
		}

		return "", myError
	}

	return resp.Body.Data.AuthenticationResult.AccessToken, myError
}

type deleteAccountRequest struct {
	AccessToken string
}

func deleteUserAccount(accessToken string) error {
	svc := getLambdaClient()

	var myError error

	token := deleteAccountRequest{accessToken}

	payload, err := json.Marshal(token)

	if err != nil {
		myError = errors.New("Error marshalling request for DeleteCognitoUser")
		return myError
	}

	//payload = []byte(payload)

	Debug.Println("Calling DeleteCognitoUser")

	result, err := svc.Invoke(&lambda.InvokeInput{FunctionName: aws.String("DeleteCognitoUser"), Payload: payload})

	if err != nil {
		myError = errors.New("Error calling DeleteCognitoUser")
		return myError
	}

	Debug.Println("")
	Debug.Println("Raw response:")
	Debug.Println(string(result.Payload))
	Debug.Println("")

	var resp userResponse
	err = json.Unmarshal(result.Payload, &resp)

	if err != nil {
		myError = errors.New("Error unmarshalling DeleteCognitoUser response")
		return myError
	}

	// Did we not get a 200?
	if resp.StatusCode != 200 {
		myError = errors.New("Expected a 200 status code, but got: " + strconv.Itoa(resp.StatusCode))
	}

	return myError
}

type deletePostRequest struct {
	AccessToken     string
	TimestampOfPost string
}

func deletePost(accessToken string, timestamp string) error {
	svc := getLambdaClient()

	var myError error

	req := deletePostRequest{accessToken, timestamp}

	payload, err := json.Marshal(req)

	Debug.Println("Raw request:")
	Debug.Println(string(payload))
	Debug.Println("")

	if err != nil {
		myError = errors.New("Error marshalling request for DeletePost")
		return myError
	}

	Debug.Println("Calling DeletePost")

	result, err := svc.Invoke(&lambda.InvokeInput{FunctionName: aws.String("DeletePost"), Payload: payload})

	if err != nil {
		myError = errors.New("Error calling DeletePost")
		return myError
	}

	Debug.Println("")
	Debug.Println("Raw response:")
	Debug.Println(string(result.Payload))
	Debug.Println("")

	var resp userResponse
	err = json.Unmarshal(result.Payload, &resp)

	if err != nil {
		myError = errors.New("Error unmarshalling response from DeletePost")
		return myError
	}

	// Did we get anything but a 200?
	if resp.StatusCode != 200 {
		var respFailure userResponseFailure
		err = json.Unmarshal(result.Payload, &respFailure)

		if err != nil {
			myError = errors.New(err.Error())
			return myError
		}

		myError = errors.New(respFailure.Body.Error.Message)
		return myError
	}

	return myError
}

type postRequest struct {
	AccessToken string
	Message     string
}

type postResponseBody struct {
	Result string
}

type finishSigninResponse struct {
	StatusCode int                 `json:"statusCode"`
	Headers    userResponseHeaders `json:"headers"`
	Body       postResponseBody    `json:"body"`
}

type defaultHeaders struct {
    ContentType string `json:"Content-Type"`
}

type defaultError struct {
    Message string `json:"message"`
}

type defaultFailureBody struct {
    Result string `json:"result"`
    Error defaultError `json:"error"`
}

type defaultFailureResponse struct {
    StatusCode int `json:"statusCode"`
    Headers defaultHeaders `json:"headers"`
    Body defaultFailureBody `json:"body"`
}

func postFromSignedInUser(accessToken string, message string) error {
	svc := getLambdaClient()

	var myError error

	request := postRequest{accessToken, message}

	payload, err := json.Marshal(request)

	if err != nil {
		myError = errors.New("Error marshalling request for AddPost")
		return myError
	}

	Debug.Println("Raw request to AddPost:")
	Debug.Println(string(payload))

	result, err := svc.Invoke(&lambda.InvokeInput{FunctionName: aws.String("AddPost"), Payload: payload})

	if err != nil {
		myError = errors.New("Error calling AddPost")
		return myError
	}

	Debug.Println("")
	Debug.Println("Raw response from AddPost:")
	Debug.Println(string(result.Payload))
	Debug.Println("")

	var resp finishSigninResponse

	err = json.Unmarshal(result.Payload, &resp)

	if err != nil {
		myError = errors.New("Error unmarshalling AddPost response")
		return myError
	}

	// Make sure we got a 200 and success
	if resp.StatusCode != 200 {
		// Try to get the error
        var respFailure defaultFailureResponse

        err = json.Unmarshal(result.Payload, &respFailure)

        if err != nil {
            myError = errors.New("Got unexpected status code: " + strconv.Itoa(resp.StatusCode))
            return myError
        } else {
            myError = errors.New("Message not posted: " + respFailure.Body.Error.Message)
            return myError
        }
	}

	if resp.Body.Result != "success" {
		myError = errors.New("Got unexpected result: " + resp.Body.Result)
		return myError
	}

	return myError
}

type startRegisterUserRequest struct {
	UserName string
	Password string
	Email    string
}

type startRegisterUserResponseData struct {
	UserConfirmed bool
}

type startRegisterUserResponseBody struct {
	Result string                        `json:"result"`
	Data   startRegisterUserResponseData `json:"data"`
}

type registerUserResponse struct {
	StatusCode int                           `json:"statusCode"`
	Headers    userResponseHeaders           `json:"headers"`
	Body       startRegisterUserResponseBody `json:"body"`
}

type registerUserResponseError struct {
	Message    string `json:"message"`
	Code       string
	Time       string
	RequestId  string
	StatusCode int
	Retryable  bool
	RetryDelay float64
}

type registerUserResponseFailureBody struct {
	Result string                    `json:"result"`
	Error  registerUserResponseError `json:"error"`
}

type registerUserResponseFailure struct {
	StatusCode int                             `json:"statusCode"`
	Headers    userResponseHeaders             `json:"headers"`
	Body       registerUserResponseFailureBody `json:"body"`
}

func startRegisterUser(name string, password string, email string) error {
	svc := getLambdaClient()

	var myError error

	request := startRegisterUserRequest{name, password, email}

	payload, err := json.Marshal(request)

	if err != nil {
		myError = errors.New("Error marshalling StartAddingPendingCognitoUser request")
		return myError
	}

	Debug.Println("Getting info about user:")
	Debug.Println(string(payload))

	result, err := svc.Invoke(&lambda.InvokeInput{FunctionName: aws.String("StartAddingPendingCognitoUser"), Payload: payload})

	if err != nil {
		myError = errors.New("Error calling StartAddingPendingCognitoUser")
		return myError
	}

	Debug.Println("")
	Debug.Println("Raw response:")
	Debug.Println(string(result.Payload))
	Debug.Println("")

	var resp registerUserResponse
	err = json.Unmarshal(result.Payload, &resp)

	if err != nil {
		myError = errors.New("Error unmarshalling StartAddingPendingCognitoUser response")
		return myError
	}

	if resp.StatusCode == 200 {
		// Got a valid response, was it success?
		if resp.Body.Result == "success" {
			return myError
		} else {
			myError = errors.New("Got result: " + resp.Body.Result)
			return myError
		}
	} else {
		var resp registerUserResponseFailure

		err = json.Unmarshal(result.Payload, &resp)

        if err != nil {
            myError = errors.New("Could not register user")
            return myError
        } else {
			myError = errors.New("Got status code: " + strconv.Itoa(resp.StatusCode) + " and error message: " +resp.Body.Error.Message)
			return myError
		}
	}

	return myError
}

type finishRegisterRequest struct {
	UserName         string
	ConfirmationCode string
}

type finishRegisterResponseBody struct {
	Result string `json:"result"`
}

type finishRegisterResponse struct {
	StatusCode int                        `json:"statusCode"`
	Headers    userResponseHeaders        `json:"headers"`
	Body       finishRegisterResponseBody `json:"body"`
}

func finishRegisterUser(name string, code string) error {
	svc := getLambdaClient()

	var myError error

	request := finishRegisterRequest{name, code}

	payload, err := json.Marshal(request)

	if err != nil {
		myError = errors.New("Error marshalling request for FinishAddingPendingCognitoUser")
		return myError
	}

	Debug.Println("Raw payload to finish registering user:")
	Debug.Println(string(payload))

	result, err := svc.Invoke(&lambda.InvokeInput{FunctionName: aws.String("FinishAddingPendingCognitoUser"), Payload: payload})

	if err != nil {
		myError = errors.New("Error calling FinishAddingPendingCognitoUser")
		return myError
	}

	Debug.Println("")
	Debug.Println("Raw response:")
	Debug.Println(string(result.Payload))
	Debug.Println("")

	var resp finishRegisterResponse
	err = json.Unmarshal(result.Payload, &resp)

	if err != nil {
		myError = errors.New("Error unmarshalling FinishAddingPendingCognitoUser response")
		return myError
	}

	if resp.Body.Result != "success" {
		var respFailure registerUserResponseFailure

		err = json.Unmarshal(result.Payload, &respFailure)

		if err != nil {
			myError = errors.New("Error unmarshalling FinishAddingPendingCognitoUser response")
			return myError
		}

		if respFailure.Body.Error.Message != "" {
			myError = errors.New(respFailure.Body.Error.Message)
			return myError
		}
	}

	return myError
}

type resetPasswordRequest struct {
	UserName string
}

type resetPasswordResponseHeaders struct {
	ContentType string `json:"Conten-Type"`
}

type resetPasswordResponseCodeDeliveryDetails struct {
	Destination       string
	DeliveryMechanism string
	AttributeName     string
}

type resetPasswordResponseData struct {
	CodeDeliveryDetails resetPasswordResponseCodeDeliveryDetails
}

type resetPasswordResponseBody struct {
	Result string                    `json:"result"`
	Data   resetPasswordResponseData `json:"data"`
}

type resetPasswordResponse struct {
	StatusCode int                       `json:"statusCode"`
	Headers    userResponseHeaders       `json:"headers"`
	Body       resetPasswordResponseBody `json:"body"`
}

func startResetPassword(userName string) error {
	svc := getLambdaClient()

	var myError error

	// Create request
	request := resetPasswordRequest{userName}

	payload, err := json.Marshal(request)

	if err != nil {
		myError = errors.New("Error marshalling StartChangingForgottenCognitoUserPassword request")
		return myError
	}

	Debug.Println("Raw request for resetting password:")
	Debug.Println(string(payload))

	result, err := svc.Invoke(&lambda.InvokeInput{FunctionName: aws.String("StartChangingForgottenCognitoUserPassword"), Payload: payload})

	if err != nil {
		myError = errors.New("Error calling StartChangingForgottenCognitoUserPassword")
		return myError
	}

	Debug.Println("")
	Debug.Println("Raw response:")
	Debug.Println(string(result.Payload))
	Debug.Println("")

	var resp resetPasswordResponse
	err = json.Unmarshal(result.Payload, &resp)

	if err != nil {
		myError = errors.New("Error unmarshalling StartChangingForgottenCognitoUserPassword response")
		return myError
	}

	if resp.StatusCode == 200 {
		if resp.Body.Result == "success" {
			return myError
		} else {
			myError = errors.New("Result: " + resp.Body.Result)
			return myError
		}
	} else {
		var respFailure registerUserResponseFailure

		err = json.Unmarshal(result.Payload, &respFailure)

		if err != nil {
			myError = errors.New("Error unmarshalling StartChangingForgottenCognitoUserPassword failure response")
			return myError
		}

		if respFailure.Body.Error.Message != "" {
			myError = errors.New(respFailure.Body.Error.Message)
			return myError
		}

	}

	return myError
}

type finishResetRequest struct {
	UserName         string
	ConfirmationCode string
	NewPassword      string
}

func finishResetPassword(userName string, cc string, pw string) error {
	svc := getLambdaClient()

	var myError error

	// Create request
	request := finishResetRequest{userName, cc, pw}

	payload, err := json.Marshal(request)

	if err != nil {
		myError = errors.New("Error marshalling request for FinishChangingForgottenCognitoUserPassword")
		return myError
	}

	Debug.Println("Raw request for final step of resetting password:")
	Debug.Println(string(payload))

	result, err := svc.Invoke(&lambda.InvokeInput{FunctionName: aws.String("FinishChangingForgottenCognitoUserPassword"), Payload: payload})

	if err != nil {
		myError = errors.New("Error calling FinishChangingForgottenCognitoUserPassword")
		return myError
	}

	Debug.Println("")
	Debug.Println("Raw response:")
	Debug.Println(string(result.Payload))
	Debug.Println("")

	var resp resetPasswordResponse
	err = json.Unmarshal(result.Payload, &resp)

	if err != nil {
		myError = errors.New("Error unmarshalling FinishChangingForgottenCognitoUserPassword response")
		return myError
	}

	if resp.StatusCode == 200 {
		if resp.Body.Result == "success" {
			return myError
		} else {
			myError = errors.New("Result: " + resp.Body.Result)
			return myError
		}
	} else {
		var respFailure registerUserResponseFailure

		err = json.Unmarshal(result.Payload, &respFailure)

		if err != nil {
			myError = errors.New("Error unmarshalling FinishChangingForgottenCognitoUserPassword failure response")
			return myError
		}

		if respFailure.Body.Error.Message != "" {
			myError = errors.New(respFailure.Body.Error.Message)
			return myError
		}
	}

	return myError
}

func usage() {
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("")

	// Re-enable once the functionality is added
	//fmt.Println("go run PostApp.go [-t TIMEZONE] [-r REGION] [-d] [-h]")

	fmt.Println("go run PostApp.go [-r REGION] [-d] [-h]")
	fmt.Println("")

	// Re-enable once the functionality is added
	//fmt.Println("If TIMEZONE is omitted, defaults to UTC")
	fmt.Println("If REGION is omitted, defaults to us-west-2")

	fmt.Println("Use -d (debug) to display additional information")
	fmt.Println("Use -h (help) to display this message and quit")

	os.Exit(0)
}

func getStringValue(scanner *bufio.Scanner, prompt string) string {
	fmt.Println(prompt)
	scanner.Scan()
	value := scanner.Text()
	value = strings.TrimSpace(value)

	return value
}

func notifySignedIn(registered bool) {
	if registered {
		fmt.Println("You are already signed in, which means you are already registered")
		fmt.Println("If you want to register or sign in as another user,")
		fmt.Println("you must sign out first and then register or sign in")
		fmt.Println("")
	} else {
		fmt.Println("You are already signed in")
		fmt.Println("If you want to sign in as another user,")
		fmt.Println("you must sign out first and then sign in")
		fmt.Println("")
	}
}

func getAndListAllPosts(maxMessages int) {
	Debug.Println("Calling getAllPosts")
	posts, err := getAllPosts(maxMessages)

	if err == nil {
		listAllPosts(posts)
	} else {
		fmt.Println("Could not get posts: " + err.Error())
	}
}

type logInUserResult struct {
	userName    string
	accessToken string
}

func logInUser(scanner *bufio.Scanner) (logInUserResult, error) {
	var myError error
	var result logInUserResult

	// Get user name
	name := getStringValue(scanner, "Enter your user name")
	fmt.Println("")
	password := getStringValue(scanner, "Enter your password")
	fmt.Println("")

	Debug.Println("Calling signInUser")
	token, err := signInUser(name, password)

	// err means something went wrong;
	// err.Error() has details
	if err == nil {
		result.userName = name
		result.accessToken = token
	} else {
		myError = errors.New("Could not sign in user: " + err.Error())
	}

	return result, myError
}

type registerUserResult struct {
	userName       string
	password       string
	accessToken    string
	signedIn       bool
	cursor         string
	pastStep1      bool
	registerPrompt string
}

func registerUser(scanner *bufio.Scanner, pastStep1 bool, name string, password string) (registerUserResult, error) {
	var result registerUserResult
	var myError error

	if pastStep1 {
		// Finish registering
		code := getStringValue(scanner, "Enter your confirmation code")
		fmt.Println("")

		Debug.Println("Calling finishRegisterUser")

		err := finishRegisterUser(name, code)

		if err != nil {
			myError = errors.New("Could not finish registering user: " + err.Error())
			return result, myError
		}

		// Sign them in
		token, err := signInUser(name, password)

		if err == nil {
			result.userName = name
			result.password = password
			result.accessToken = token
			result.signedIn = true
			result.cursor = "(" + name + ")> "
			result.pastStep1 = false
			result.registerPrompt = "3: Register as new user"

			return result, myError
		} else {
			myError = errors.New("Could not register user: " + err.Error())
			return result, myError
		}
	} else {
		Debug.Println("Calling startRegisterUser")

		// We need the name here
		name = getStringValue(scanner, "Enter your user name")
		password = getStringValue(scanner, "Enter a password with at least 6 characters")

		if len(password) < 6 {
			myError = errors.New("You password is too short, try again")
			return result, myError
		}

		email := getStringValue(scanner, "Enter your email address")
		fmt.Println("")

		err := startRegisterUser(name, password, email)

		if err == nil {
			result.userName = name
			result.password = password
			// Change registration prompt
			result.registerPrompt = "3: Finish registering (automatically signs you in)"
			result.pastStep1 = true
		} else {
			myError = errors.New("Could not start registering user: " + err.Error())
		}

		return result, myError
	}
}

type resetPasswordResult struct {
	cursor              string
	pastStep1           bool
	resetPasswordPrompt string
    signedIn bool
}

func resetPassword(scanner *bufio.Scanner, pastStep1 bool, name string) (resetPasswordResult, error) {
	var result resetPasswordResult
	var myError error

	if pastStep1 {
		// Finish resetting password
		Debug.Println("Calling finishResetPassword")

		// Get confirmation code and new password
		cc := getStringValue(scanner, "Enter the confirmation code")
		fmt.Println("")
		pw := getStringValue(scanner, "Enter your new password")
		fmt.Println("")

		err := finishResetPassword(name, cc, pw)

		if err == nil {
			Debug.Println("Successfully reset password")
			// user is signed in with new password
			result.cursor = "(" + name + ")> "
			result.pastStep1 = false
			result.resetPasswordPrompt = "4: Reset password"
            result.signedIn = true

			return result, myError
		} else {
			myError = errors.New("Could not reset password: " + err.Error())
			return result, myError
		}
	} else {
		Debug.Println("Calling startResetPassword")
		Debug.Println("For user " + name)

		err := startResetPassword(name)

		if err == nil {
			result.pastStep1 = true
			result.resetPasswordPrompt = "4: Finish resetting password"
		} else {
			myError = errors.New("Could not reset password: " + err.Error())
		}

		return result, myError
	}
}

func postMessage(scanner *bufio.Scanner, accessToken string) error {
	var myError error

	// Query for message to post
	message := getStringValue(scanner, "Enter the message to post")

	Debug.Println("Calling postMessage")

	err := postFromSignedInUser(accessToken, message)

	if err == nil {
		fmt.Println("Message posted")
		return myError
	} else {
		myError = errors.New("Message not posted: " + err.Error())
		return myError
	}
}

func deleteAccount(accessToken string) error {
	var myError error

	err := deleteUserAccount(accessToken)

	if err == nil {
		fmt.Println("Your account has been deleted")
	} else {
		myError = errors.New("Could not delete account: " + err.Error())
	}

	return myError
}

func deleteMyPost(scanner *bufio.Scanner, accessToken string) error {
	var myError error

	// Get the ID of the post
	timestamp := getStringValue(scanner, "Enter the ID of the post to delete (the ID is the long number at the end of the first line):")
	fmt.Println("")

	err := deletePost(accessToken, timestamp)

	if err != nil {
		myError = errors.New("Could not delete post: " + err.Error())
	}

	return myError
}

func main() {
	/*

	   The strategy for parsing JSON response is to get the statusCode,
	   headers (and Content-Type), and body (and result).

	   If the statusCode is not 200, the call failed.
	   If result is not "success", the operation failed.

	   In either of these two cases, we reconstitute the JSON as the failed version
	   of the object and display the error message if it exists.

	*/
    SetConfiguration()

	regionPtr := flag.String("r", configuration.Region, "Region to look for services")
	timezonePtr := flag.String("t", configuration.Timezone, "Timezone for displayed date and time")
	maxMsgsPtr := flag.Int("n", configuration.MaxMessages, "Maximum number of messages to download")
	refreshPtr := flag.Int("f", configuration.RefreshSeconds, "Duration, in seconds, between refreshing post list")
	debugPtr := flag.Bool("d", configuration.Debug, "Whether to show debug output")
	helpPtr := flag.Bool("h", false, "Show usage")

    flag.Parse()

    // Save configuration
    configuration.Region = *regionPtr
    configuration.Timezone = *timezonePtr
    configuration.MaxMessages = *maxMsgsPtr
    configuration.RefreshSeconds = *refreshPtr
    configuration.Debug = *debugPtr

    help := *helpPtr

    if help {
        usage()
        os.Exit(0)
    }

    if configuration.Debug {
        initLog(os.Stderr)
    } else {
        initLog(ioutil.Discard)
    }

    Debug.Println("Region:     " + configuration.Region)
    Debug.Println("Timezone:   " + configuration.Timezone)
    Debug.Println("Max # msgs: " + strconv.Itoa(configuration.MaxMessages))
    Debug.Println("Refresh:    " + strconv.Itoa(configuration.RefreshSeconds))

	cursor := "(anonymous)> "

	// When false, stop the app
	keepGoing := true

	// The name of the current user
	userName := ""

	// True if signed in (required to post)
	signedIn := false

	// Initialize a session that the SDK will use to load configuration,
	// credentials, and region from the shared config file. (~/.aws/config).
	Debug.Println("Calling Lambda function in: " + configuration.Region)

	scanner := bufio.NewScanner(os.Stdin)
	inputValue := ""

	// So we can adjust prompt for two-stage processes
	pastStep1 := false // set to true once start* completes successfully; reset to false after finish* completes successfully
	registerPrompt := "3: Register as new user"
	resetPasswordPrompt := "4: Reset password"

	var password string = ""
	var accessToken string = ""

	for keepGoing {
		// Menu
		fmt.Println("")
		fmt.Println("Enter a value between 1 and 8 to perform the indicated action or q (or Q) to quit:")
		fmt.Println("")
		fmt.Println("1: List all posts")
		fmt.Println("2: Sign in")
		fmt.Println(registerPrompt)
		fmt.Println(resetPasswordPrompt)
		fmt.Println("5: Post a message (you must be signed in)")
		fmt.Println("6: Sign out")
		fmt.Println("7: Delete your account (you must be signed in)")
		fmt.Println("8: Delete a post (you must be signed in and it must be your post)")
		fmt.Println("q (or Q): Quit")
		fmt.Println("")

		inputValue = getStringValue(scanner, cursor)

		if inputValue != "q" && inputValue != "Q" && !configuration.Debug {
			clearScreen()
		}

		Debug.Println("Got: '" + inputValue + "'")

		switch inputValue {
		case "1":
			// Get and list all posts
			getAndListAllPosts(configuration.MaxMessages)

		case "2":
			// sign in user

			// if already signed in, tell user
			if signedIn {
				notifySignedIn(false) // true means tell them they are already registered
				continue
			}

			result, err := logInUser(scanner)

			if err != nil {
				fmt.Println(err.Error())
			} else {
				signedIn = true

				userName = result.userName
				accessToken = result.accessToken

				cursor = "(" + userName + ")> "
			}

		case "3":
			// register

			// if already signed in, tell user
			if signedIn {
				notifySignedIn(true) // true means tell them they are already registered
				continue
			}

			result, err := registerUser(scanner, pastStep1, userName, password)

			if err != nil {
				fmt.Println(err.Error())
			} else {
				userName = result.userName
				password = result.password
				accessToken = result.accessToken
				signedIn = result.signedIn
				cursor = result.cursor
				pastStep1 = result.pastStep1
				registerPrompt = result.registerPrompt
			}

		case "4":
			// Reset password
			if !pastStep1 {
				userName = getStringValue(scanner, "Enter your user name:")
				fmt.Println("")
			}

			result, err := resetPassword(scanner, pastStep1, userName)

			if err == nil {
				cursor = result.cursor
				pastStep1 = result.pastStep1
				resetPasswordPrompt = result.resetPasswordPrompt
                signedIn = result.signedIn
			} else {
				fmt.Println(err.Error())
			}

		case "5":
			// post message
			if !signedIn {
				fmt.Println("You must be signed in to post a message")
				continue
			}

			err := postMessage(scanner, accessToken)

			if err != nil {
				fmt.Println(err.Error())
			}

		case "6":
			if !signedIn {
				fmt.Println("Your are not signed in")
				continue
			}

			// sign out
			signedIn = false
			userName = ""
			accessToken = ""
			cursor = "(anonymous)> "

		case "7":
			// delete account
			if !signedIn {
				fmt.Println("You must be signed in to delete your account")
				continue
			}

			err := deleteAccount(accessToken)

			if err == nil {
				signedIn = false
				userName = ""
				accessToken = ""
				cursor = "(anonymous)> "
			} else {
				fmt.Println(err.Error())
			}

		case "8":
			// delete post
			if !signedIn {
				fmt.Println("You must be signed in to delete a post")
				continue
			}

			err := deleteMyPost(scanner, accessToken)

			if err == nil {
				fmt.Println("Post deleted")
			} else {
				fmt.Println(err.Error())
			}

		case "q", "Q":
			// quite
			keepGoing = false

		default:
			fmt.Println("Unrecognized option: " + inputValue)
		}
	}
}
