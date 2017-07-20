package main

/*
  Workflow:

  1. User hits /, which takes them to StartServer
     Their initial status is 'Not logged in'.
     Since they aren't logged in, they have four options:
     a. Don't do anything
        They can only see the posts.
        TBD: do we want to have a "refresh" option? Auto? Button?
     b. Log in with username and password.
        i.   Button -> posts username and password to /login (LoginServer)
        ii.  LoginServer calls log_in_user with username and password
        iii. If log_in_user returns '' (empty token),
             LoginServer calls / with status 'Not logged in' (start over).
             Otherwise it sets the status to 'Logged in'
             and calls HomeServer.
     c. Register with username, email address, and password
        i.   Button -> posts username, email address, and password to /register (RegisterServer)
        ii.  RegisterServer calls startRegisterUser with username, email address, and password
        iii. If startRegisterUser is successful, RegisterServer sets the status to "Registering",
             and calls itself.
        iv.

     d. Reset password with username field and Submit button.
        Button takes them to /reset.
        At /reset_password they see a form with two text fields and a button.
        i. Finish resetting password with confirmation code and new password.
           Button takes them to finish_reset.
           If resetting is successful, they go to / with status loggedin.
  2. Now that they are logged in, / shows N forms:
     a. Create post with a message field and Submit button.
        Button takes them to / with status loggedin.
     b. Logout with Submit button.
        Button takes them to / with status !loggedin.

  Note that every page displays a message at the top of the page as.

 */

import (
	"encoding/json"
    "errors"
    "flag"
    "fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
    "text/template"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
)

// Used for status
type StatusType uint8

const (
    NOT_LOGGED_IN StatusType = iota
    LOGGED_IN
    LOGGING_IN
    LOGIN_FAILED
    MESSAGE_DELETED
    MESSAGE_DELETE_FAILED
    MESSAGE_POSTED
    MESSAGE_FAILED
    // REGISTERED -> LOGGED_IN
    REGISTERING
    REGISTRATION_FAILED
    // RESET -> LOGGED_IN
    RESETTING
    RESET_FAILED
)

// Status
var status StatusType

func getStatusValue() string {
    value := ""

    switch status {
    case NOT_LOGGED_IN:
        value = "Not logged in"
    case LOGGED_IN:
        value = "Logged in"
    case LOGGING_IN:
        value = "Logging in"
    case LOGIN_FAILED:
        value = "Login failed"
    case MESSAGE_DELETED:
        value = "Message deleted"
    case MESSAGE_DELETE_FAILED:
        value = "Failed to delete message"
    case MESSAGE_POSTED:
        value = "Message posted"
    case MESSAGE_FAILED:
        value = "Failed to post message"
    case REGISTERING:
        value = "Registering"
    case REGISTRATION_FAILED:
        value = "Registration failed"
    case RESETTING:
        value = "Resetting password"
    case RESET_FAILED:
        value = "Resetting password failed"
    }

    return value
}

// Global variables
// Log
var Debug *log.Logger

type Configuration struct {
    Region      string
    Timezone    string
    MaxMessages int
    RefreshSeconds int
    Debug bool
}

// Configuration
var configuration Configuration

// User token
var token string

var username string
var password string

// Templates
var templates *template.Template

// For -h option
func usage() {
    fmt.Println("")
    fmt.Println("Usage:")
    fmt.Println("")

    // Re-enable once the functionality is added
    //fmt.Println("go run PostApp.go [-r REGION] [-t TIMEZONE] [-n MAX_MESSAGES] [-f REFRESH] [-d] [-h]")

    fmt.Println("go run PostApp.go [-r REGION] [-n MAX_MESSAGES] [-d] [-h]")
    fmt.Println("")

    // Re-enable once the functionality is added
    // fmt.Println("If TIMEZONE is omitted, defaults to UTC")
    // fmt.Println("If REFRESH is omitted, defaults to 30 (seconds)")

    fmt.Println("If REGION is omitted, defaults to us-west-2")
    fmt.Println("If MAX_MESSAGES is omitted, defaults to 20")

    fmt.Println("Use -d (debug) to display additional information")
    fmt.Println("Use -h (help) to display this message and quit")

    os.Exit(0)
}

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

type PostEntry struct {
    Date string
    Message string
    Timestamp string
}

func SetConfiguration() {
    if configuration == (Configuration{}) {
        // Get configuration values
        file, _ := os.Open("conf.json")
        decoder := json.NewDecoder(file)

        err := decoder.Decode(&configuration)

        if err != nil {
            // Set configuration to default values
            configuration.Debug = false
            configuration.MaxMessages = 20
            configuration.Region = "us-west-2"
            configuration.RefreshSeconds = 30
            configuration.Timezone = "UTC"
        }
    }
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

// Get all posts as an array of postEntry items
func getAllPosts() ([]PostEntry) {
    svc := getLambdaClient()

    var resp getPostsResponse
    var posts []PostEntry

    // Flags and their default values:
    maxMessages := configuration.MaxMessages

	// Get the latest maxMessages posts
	request := getPostsRequest{"timestamp", "descending", maxMessages}

	payload, err := json.Marshal(request)

	if err != nil {
        log.Fatal("Error marshalling GetPosts request: " + err.Error())
	}

	result, err := svc.Invoke(&lambda.InvokeInput{FunctionName: aws.String("GetPosts"), Payload: payload})

	if err != nil {
        log.Fatal("Error calling Lambda function GetPosts: " + err.Error())
	}

	err = json.Unmarshal(result.Payload, &resp)

	if err != nil {
        log.Fatal("Error unmarshalling GetPosts response: " + err.Error())
	}

	// Check the status code
	if resp.StatusCode != 200 {
        log.Fatal("Error getting posts, StatusCode: " + strconv.Itoa(resp.StatusCode))
	}

	numPosts := len(resp.Body.Data)

	if numPosts > 0 {
		var origDate FormatAsDate

        var post PostEntry

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
                    var blankPost PostEntry
                    blankPost.Date = "=== " + theDate.String() + " ==="
                    blankPost.Message = ""
                    blankPost.Timestamp = ""

                    posts = append(posts, blankPost)

					origDate = theDate
				}

				post.Date = p.Alias.S + "@" + theTime.String()
				post.Message = p.Message.S
				post.Timestamp = p.Timestamp.S
			} else {
				post.Date = p.Alias.S + "@??? "
				post.Message = p.Message.S
				post.Timestamp = p.Timestamp.S
			}

            posts = append(posts, post)
		}
	}

	return posts
}

func ParseTemplates() {
    var allFiles []string

    files, err := ioutil.ReadDir(".")

    if err != nil {
        log.Fatal("Error getting files in current folder: " + err.Error())
    }

    for _, file := range files {
        filename := file.Name()

        if strings.HasSuffix(filename, ".tmpl") {
            allFiles = append(allFiles, "./"+filename)
        }
    }

    // Parse all .tmpl files in this folder
    templates, err = template.ParseFiles(allFiles...)

    if err != nil {
        log.Fatal("Error parsing templates: " + err.Error())
    }
}

type HeaderContext struct {
    Message string
    Title string
}

type PostsContext struct {
    Posts []PostEntry
}

// See the following web page for info on automatically refreshing the posts
//     https://blog.markvincze.com/programmatically-refreshing-a-browser-tab-from-a-golang-application/

/*
We always start here, as we are not logged in.

Every *Server function uses the following templates,
in the order listed:

  1. header.tmpl
     Contains the opening HTML tags and a status message
  2. posts.tmpl
     Contains the list of posts
  3. *.tmpl (matching the first part of the function name
     Contains text fields and button(s) to login in, register, logout, ...
  4. footer.tmpl
     Contains the closing HTML tags
 */
func StartServer(w http.ResponseWriter, req *http.Request) {
    Debug.Println("The status in StartServer is: " + getStatusValue())

    message := "You must be logged in (or registered, which automatically logs you in) before you can post, delete a post, or delete your account."

    // Make sure they didn't get here on accident
    switch status {
        case LOGGED_IN:
        Debug.Println("Calling HomeServer from StartServer")
        HomeServer(w, req)

    case RESETTING:
        message = "Enter your confirmation code and click <b>Submit</b> to finish resetting your password"
        var headerContext HeaderContext
        headerContext = HeaderContext{Message: message, Title: "Chat App"}
        s1 := templates.Lookup("header.tmpl")
        s1.Execute(w, headerContext)

        var postContext PostsContext
        posts := getAllPosts()
        postContext = PostsContext{Posts: posts}
        s2 := templates.Lookup("posts.tmpl")
        s2.Execute(w, postContext)

        s3 := templates.Lookup("reset.tmpl")
        s3.Execute(w, nil)

        s4 := templates.Lookup("footer.tmpl")
        s4.Execute(w, nil)

    case REGISTERING:
        message = "Enter your confirmation code and click <b>Submit</b> to finish registering"
        var headerContext HeaderContext
        headerContext = HeaderContext{Message: message, Title: "Chat App"}
        s1 := templates.Lookup("header.tmpl")
        s1.Execute(w, headerContext)

        var postContext PostsContext
        posts := getAllPosts()
        postContext = PostsContext{Posts: posts}
        s2 := templates.Lookup("posts.tmpl")
        s2.Execute(w, postContext)

        s3 := templates.Lookup("register.tmpl")
        s3.Execute(w, nil)

        s4 := templates.Lookup("footer.tmpl")
        s4.Execute(w, nil)

    default:
        // Change message if attempt to login, register, or reset password failed
        if status == LOGIN_FAILED {
            message = "<b>Login failed!</b> " + message
        }

        if status == REGISTRATION_FAILED {
            message = "<b>Registration failed!</b>! " + message
        }

        if status == RESET_FAILED {
            message = "<b>Resetting password failed!</b> " + message
        }

        status = NOT_LOGGED_IN

        // Beginning HTML tags, includinge common message (paragraph)
        var headerContext HeaderContext
        headerContext = HeaderContext{Message: message, Title: "Chat App"}
        s1 := templates.Lookup("header.tmpl")
        s1.Execute(w, headerContext)

        // Display the posts
        posts := getAllPosts()

        numMsgs := len(posts)

        Debug.Println("Got: " + strconv.Itoa(numMsgs) + " posts")

        var postContext PostsContext
        postContext = PostsContext{Posts: posts}
        s2 := templates.Lookup("posts.tmpl")
        s2.Execute(w, postContext)

        // Forms for log in, register, reset password
        s3 := templates.Lookup("start.tmpl")
        s3.Execute(w, nil)

        // Closing HTML tags
        s4 := templates.Lookup("footer.tmpl")
        s4.Execute(w, nil)
    }
}

func AboutServer(w http.ResponseWriter, req *http.Request) {
    Debug.Println("The status in AboutServer is: " + getStatusValue())

    message := ""

    var headerContext HeaderContext
    headerContext = HeaderContext{Message: message, Title: "About the Chat App"}

    s1 := templates.Lookup("header.tmpl")
    s1.Execute(w, headerContext)

    s2 := templates.Lookup("about.tmpl")
    s2.Execute(w, nil)

    s3 := templates.Lookup("footer.tmpl")
    s3.Execute(w, nil)
}

func ContactServer(w http.ResponseWriter, req *http.Request) {
    Debug.Println("The status in ContactServer is: " + getStatusValue())

    message := ""

    var headerContext HeaderContext
    headerContext = HeaderContext{Message: message, Title: "Contact info for the Chat App"}

    s1 := templates.Lookup("header.tmpl")
    s1.Execute(w, headerContext)

    s2 := templates.Lookup("contact.tmpl")
    s2.Execute(w, nil)

    s3 := templates.Lookup("footer.tmpl")
    s3.Execute(w, nil)
}

func HomeServer(w http.ResponseWriter, req *http.Request) {
    Debug.Println("")
    Debug.Println("The status in HomeServer is: " + getStatusValue())

    switch status {

    case NOT_LOGGED_IN:
        Debug.Println("Calling StartServer from HomeServer")
        StartServer(w, req)

    default:
        message := getStatusValue()
        status = LOGGED_IN
        Debug.Println("Setting status to " + getStatusValue() + " in HomeServer")

        var headerContext HeaderContext
        headerContext = HeaderContext{Message: message, Title: "Chat App"}
        s1 := templates.Lookup("header.tmpl")
        s1.Execute(w, headerContext)

        var postContext PostsContext
        posts := getAllPosts()
        postContext = PostsContext{Posts: posts}
        s2 := templates.Lookup("posts.tmpl")
        s2.Execute(w, postContext)

        // Form for submitting a post and
        // buttons for deleting a selected post, logging out, deleting account
        s3 := templates.Lookup("home.tmpl")
        s3.Execute(w, nil)

        s4 := templates.Lookup("footer.tmpl")
        s4.Execute(w, nil)
    }
}

type user struct {
    UserName string
    Password string
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

type userResponse struct {
    StatusCode int                 `json:"statusCode"`
    Headers    userResponseHeaders `json:"headers"`
    Body       userResponseBody    `json:"body"`
}

func logInUser(userName string, password string) (string, error) {
    var newToken = ""
    var myError error

    user := user{userName, password}

    payload, err := json.Marshal(user)

    if err != nil {
        myError = errors.New("Error marshalling SignInCognitoUser request: " + err.Error())
        return newToken, myError
    }

    svc := getLambdaClient()

    result, err := svc.Invoke(&lambda.InvokeInput{FunctionName: aws.String("SignInCognitoUser"), Payload: payload})

    if err != nil {
        myError = errors.New("Error calling SignInCognitoUser: " + err.Error())
        return newToken, myError
    }

    var resp userResponse
    err = json.Unmarshal(result.Payload, &resp)

    if err != nil {
        myError = errors.New("Error unmarshalling SignInCognitoUser response: " + err.Error())
        return newToken,myError
    }

    // Did we not get a 200?
    if resp.StatusCode != 200 {
        var respFailure registerUserResponseFailure

        err = json.Unmarshal(result.Payload, &respFailure)

        Debug.Println("Got a status code of: " + strconv.Itoa(respFailure.StatusCode))
        Debug.Println("and a message of:      " + respFailure.Body.Error.Message)

        myError = errors.New(respFailure.Body.Error.Message)
        return newToken, myError
    }

    if resp.Body.Result != "success" {
        Debug.Println("Got a result of: " + resp.Body.Result)
        myError = errors.New("Got a result of: " + resp.Body.Result)
        return newToken, myError
    }

    return resp.Body.Data.AuthenticationResult.AccessToken, myError
}

func LoginServer(w http.ResponseWriter, req *http.Request) {
    Debug.Println("")
    Debug.Println("LoginServer called")

    username := ""
    password := ""

    switch status {

    case LOGGED_IN:
        // They're already logged in
        Debug.Println("Calling HomeServer from LoginServer")
        HomeServer(w, req)
    default:
        // Get username and password and log them in
        req.ParseForm()    // Parses the request body

        username = req.Form.Get("username")
        password = req.Form.Get("password")

        Debug.Println("Calling logInUser with user name: " + username + " and password: " + password)

        newToken, err := logInUser(username, password)

        if err != nil {
            fmt.Println("Login failed")
            // Login failed, so send them back to start
            status = LOGIN_FAILED
            StartServer(w, req)
        } else {
            token = newToken
            fmt.Println("User is now logged in")
            status = LOGGED_IN
            Debug.Println("Calling HomeServer from LoginServer")
            HomeServer(w, req)
        }
    }
}

func LogoutServer(w http.ResponseWriter, req *http.Request) {
    Debug.Println("")
    Debug.Println("LogoutServer called")
    // This shouldn't happen,
    // but if not logged in,
    // we have nothing to do,
    // so just redirect them to the start
    // Nuke global info
    token = ""
    username = ""
    status = NOT_LOGGED_IN
    StartServer(w, req)
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

type userResponseHeaders struct {
    ContentType string `json:"Content-Type"`
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
    var myError error

    request := startRegisterUserRequest{name, password, email}

    payload, err := json.Marshal(request)

    if err != nil {
        myError = errors.New("Error marshalling StartAddingPendingCognitoUser request: " + err.Error())
        return myError
    }

    svc := getLambdaClient()

    result, err := svc.Invoke(&lambda.InvokeInput{FunctionName: aws.String("StartAddingPendingCognitoUser"), Payload: payload})

    if err != nil {
        myError = errors.New("Error calling StartAddingPendingCognitoUser: " + err.Error())
        return myError
    }

    var resp registerUserResponse
    err = json.Unmarshal(result.Payload, &resp)

    if err != nil {
        myError = errors.New("Error unmarshalling StartAddingPendingCognitoUser response: " + err.Error())
        return myError
    }

    if resp.StatusCode == 200 {
        // Got a valid response, was it success?
        if resp.Body.Result == "success" {
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

func finishRegisterUser(name string, code string, password string) (string, error) {
    newToken := ""
    var myError error

    request := finishRegisterRequest{name, code}

    payload, err := json.Marshal(request)

    if err != nil {
        myError = errors.New("Error marshalling request for FinishAddingPendingCognitoUser: " + err.Error())
        return newToken, myError
    }

    svc := getLambdaClient()

    result, err := svc.Invoke(&lambda.InvokeInput{FunctionName: aws.String("FinishAddingPendingCognitoUser"), Payload: payload})

    if err != nil {
        myError = errors.New("Error calling FinishAddingPendingCognitoUser: " + err.Error())
        return newToken, myError
    }

    var resp finishRegisterResponse
    err = json.Unmarshal(result.Payload, &resp)

    if err != nil {
        myError = errors.New("Error unmarshalling FinishAddingPendingCognitoUser response: " + err.Error())
        return newToken, myError
    }

    if resp.StatusCode == 200 {
        // Got a valid response, was it success?
        if resp.Body.Result == "success" {
            // Log them in and get token
            newToken, err = logInUser(name, password)

            if err != nil {
                myError = errors.New("Error logging user in: " + err.Error())
                return "", myError
            }
        }
    }

    return newToken, myError
}

func RegisterServer(w http.ResponseWriter, req *http.Request) {
    Debug.Println("")
    Debug.Println("RegisterServer called with status: " + getStatusValue())

    switch status {
    case LOGGED_IN:
        // If they are already logged in they are already registered
        Debug.Println("Calling HomeServer from RegisterServer")
        HomeServer(w, req)
    case NOT_LOGGED_IN:
        // Ths first time we're called
        // Get request values
        req.ParseForm()    // Parses the request body

        username = req.Form.Get("username")
        password = req.Form.Get("password")
        email := req.Form.Get("email")

        Debug.Println("Calling startRegisterUser with:")
        Debug.Println("   Username: " + username)
        Debug.Println("   Password: " + password)
        Debug.Println("   Email     " + email)

        err := startRegisterUser(username, password, email)

        if err == nil {
            status = REGISTERING
            StartServer(w, req)
        } else {
            // Start registering failed, so shoot them back to start
            status = REGISTRATION_FAILED
            StartServer(w, req)
        }
    case REGISTERING:
        // The second time through
        Debug.Println("User is finishing registering")

        req.ParseForm()

        code := req.Form.Get("code")

        newToken, err := finishRegisterUser(username, code, password)

        if err != nil {
            status = REGISTRATION_FAILED
            StartServer(w, req)
        } else {
            token = newToken
            status = LOGGED_IN
            HomeServer(w, req)
        }
    }
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
    var myError error

     // Create request
    request := resetPasswordRequest{userName}

    payload, err := json.Marshal(request)

    if err != nil {
        myError = errors.New("Error marshalling StartChangingForgottenCognitoUserPassword request: " + err.Error())
        return myError
    }

    svc := getLambdaClient()

    result, err := svc.Invoke(&lambda.InvokeInput{FunctionName: aws.String("StartChangingForgottenCognitoUserPassword"), Payload: payload})

    if err != nil {
        myError = errors.New("Error calling StartChangingForgottenCognitoUserPassword: " + err.Error())
        return myError
    }

    var resp resetPasswordResponse
    err = json.Unmarshal(result.Payload, &resp)

    if err != nil {
        myError = errors.New("Error unmarshalling StartChangingForgottenCognitoUserPassword response: " + err.Error())
        return myError
    }

    if resp.StatusCode == 200 {
        if resp.Body.Result == "success" {
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

func finishResetPassword(userName string, cc string, pw string) (string, error) {
    var myError error
    theToken := ""

    // Create request
    request := finishResetRequest{userName, cc, pw}

    payload, err := json.Marshal(request)

    if err != nil {
        myError = errors.New("Error marshalling request for FinishChangingForgottenCognitoUserPassword: " + err.Error())
        return theToken, myError
    }

    svc := getLambdaClient()

    result, err := svc.Invoke(&lambda.InvokeInput{FunctionName: aws.String("FinishChangingForgottenCognitoUserPassword"), Payload: payload})

    if err != nil {
        myError = errors.New("Error calling FinishChangingForgottenCognitoUserPassword: " + err.Error())
        return theToken, myError
    }

    var resp resetPasswordResponse
    err = json.Unmarshal(result.Payload, &resp)

    if err != nil {
        myError = errors.New("Error unmarshalling FinishChangingForgottenCognitoUserPassword response: " + err.Error())
        return theToken, myError
    }

    if resp.StatusCode == 200 {
        if resp.Body.Result == "success" {
            theToken, err := logInUser(username, pw)

            if err != nil {
                myError = errors.New("Error loggging in user in ??: " + err.Error())
                return theToken, myError
            }
        }
    } else {
        var failResp registerUserResponseFailure
        err = json.Unmarshal(result.Payload, &failResp)

        if err == nil {
            myError = errors.New("Got error message finishing password reset: " + failResp.Body.Error.Message)
            return theToken, myError
        }
    }

    return theToken, myError
}

func ResetServer(w http.ResponseWriter, req *http.Request) {
    Debug.Println("")
    Debug.Println("ResetServer called with status: " + getStatusValue())

    switch status {
    case LOGGED_IN:
        // If they are already logged in they are already registered
        Debug.Println("Calling HomeServer from ResetServer")
        HomeServer(w, req)
    case NOT_LOGGED_IN:
        // Ths first time we're called
        // Get request values
        req.ParseForm()    // Parses the request body

        username = req.Form.Get("username")

        Debug.Println("Calling startResetPassword with:")
        Debug.Println("   Username: " + username)

        err := startResetPassword(username)

        if err != nil {
            // Start resetting failed, so shoot them back to start
            status = RESET_FAILED
            StartServer(w, req)
        } else {
            status = RESETTING
            // StartServer sees
            // status == RESETTING
            // and creates a new form with reset.tmpl as 3rd item.
            StartServer(w, req)
        }
    case RESETTING:
        // The second time through
        Debug.Println("User is finishing resetting their password")

        req.ParseForm()    // Parses the request body

        password := req.Form.Get("password")
        code := req.Form.Get("code")

        Debug.Println("Calling finishResetPassword with:")
        Debug.Println("   Username:          " + username)
        Debug.Println("   Verification code: " + code)
        Debug.Println("   Password:          " + password)

        theToken, err := finishResetPassword(username, code, password)

        if err == nil && theToken != "" {
            token = theToken
            status = LOGGED_IN
            HomeServer(w, req)
        } else {
            status = RESET_FAILED
            StartServer(w, req)
        }
    }
}

type deleteAccountRequest struct {
    AccessToken string
}

func deleteUserAccount(accessToken string) error {
    var myError error

    token := deleteAccountRequest{accessToken}

    payload, err := json.Marshal(token)

    if err != nil {
        myError = errors.New("Error marshalling request for DeleteCognitoUser")
        return myError
    }

    svc := getLambdaClient()

    result, err := svc.Invoke(&lambda.InvokeInput{FunctionName: aws.String("DeleteCognitoUser"), Payload: payload})

    if err != nil {
        myError = errors.New("Error calling DeleteCognitoUser")
        return myError
    }

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

func UnregisterServer(w http.ResponseWriter, req *http.Request) {
    Debug.Println("")
    Debug.Println("UnregisterServer called")

    err := deleteUserAccount(token)

    if err == nil {
        token = ""
        username = ""
        password = ""

        status = NOT_LOGGED_IN
        StartServer(w, req)
    }
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

func postFromSignedInUser(accessToken string, message string) error {
    var myError error

    request := postRequest{accessToken, message}

    payload, err := json.Marshal(request)

    if err != nil {
        myError = errors.New("Error marshalling request for AddPost: " + err.Error())
        return myError
    }

    svc := getLambdaClient()

    result, err := svc.Invoke(&lambda.InvokeInput{FunctionName: aws.String("AddPost"), Payload: payload})

    if err != nil {
        myError = errors.New("Error calling AddPost: " + err.Error())
        return myError
    }

    var resp finishSigninResponse

    err = json.Unmarshal(result.Payload, &resp)

    if err != nil {
        myError = errors.New("Error unmarshalling AddPost response: " + err.Error())
        return myError
    }

    // Make sure we got a 200 and success
    if resp.StatusCode != 200 {
        myError = errors.New("Got status code: " + strconv.Itoa(resp.StatusCode))
        return myError
    }

    if resp.Body.Result != "success" {
        myError = errors.New("Bad result: " + resp.Body.Result)
        return myError
    }

    return myError
}

func PostServer(w http.ResponseWriter, req *http.Request) {
    Debug.Println("")
    Debug.Println("PostServer called with status: " + getStatusValue())

    req.ParseForm()    // Parses the request body

    message := req.Form.Get("message")

    err := postFromSignedInUser(token, message)

    if err != nil {
        status = MESSAGE_FAILED
    } else {
        status = MESSAGE_POSTED
    }

    HomeServer(w, req)
}

type deletePostRequest struct {
    AccessToken     string
    TimestampOfPost string
}

func deletePost(accessToken string, timestamp string) error {
    var myError error

    req := deletePostRequest{accessToken, timestamp}

    payload, err := json.Marshal(req)

    if err != nil {
        myError = errors.New("Error marshalling request for DeletePost")
        return myError
    }

    svc := getLambdaClient()

    result, err := svc.Invoke(&lambda.InvokeInput{FunctionName: aws.String("DeletePost"), Payload: payload})

    if err != nil {
        myError = errors.New("Error calling DeletePost")
        return myError
    }

    var resp userResponse
    err = json.Unmarshal(result.Payload, &resp)

    if err != nil {
        myError = errors.New("Error unmarshalling response from DeletePost")
        return myError
    }

    // Did we get anything but a 200?
    if resp.StatusCode != 200 {
        myError = errors.New("Got bad status code: " + strconv.Itoa(resp.StatusCode))
    }

    return myError
}

func DeleteServer(w http.ResponseWriter, req *http.Request) {
    Debug.Println("")
    Debug.Println("DeleteServer called with status: " + getStatusValue())

    req.ParseForm()    // Parses the request body

    timestamp := req.Form.Get("message_value")

    err := deletePost(token, timestamp)

    if err == nil {
        status = MESSAGE_DELETED
    } else {
        status = MESSAGE_DELETE_FAILED
    }

    HomeServer(w, req)
}

func main() {
    // Override default value if configuration is parsed correctly
    SetConfiguration()

    regionPtr := flag.String("r", configuration.Region, "Region to look for services")
    timezonePtr := flag.String("t", configuration.Timezone, "Timezone for displayed date and time")
    maxMsgsPtr := flag.Int("n", configuration.MaxMessages, "Maximum number of messages to download")
    refreshPtr := flag.Int("f", configuration.RefreshSeconds, "Duration, in seconds, between refreshing post list")
    debugPtr := flag.Bool("d", configuration.Debug, "Whether to show debug output")
    helpPtr := flag.Bool("h", false, "Show usage")

    flag.Parse()

    // Save configuration if it's changed
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

    ParseTemplates()

    // When we start we aren't logged in, so tell them what to do
    status = NOT_LOGGED_IN

    Debug.Println("The initial status is: " + getStatusValue())

    // The same order as myapp.rb:
    http.HandleFunc("/", StartServer)
    http.HandleFunc("/about", AboutServer)
    http.HandleFunc("/contact", ContactServer)
    http.HandleFunc("/delete", DeleteServer)
    http.HandleFunc("/home", HomeServer)
    http.HandleFunc("/login", LoginServer)
    http.HandleFunc("/logout", LogoutServer)
    http.HandleFunc("/post", PostServer)
    http.HandleFunc("/register", RegisterServer)
    http.HandleFunc("/reset", ResetServer)
    http.HandleFunc("/unregister", UnregisterServer)

	// Get port # from environemt or use 12345
	port := os.Getenv("PORT")

	if port == "" {
		port = ":12345"
	}

    err := http.ListenAndServe(port, nil)

    if err != nil {
        log.Fatal("ListenAndServe returned error: ", err)
    }
}
