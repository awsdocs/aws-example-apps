using System;
using System.Collections.Generic;
using System.Text;
using System.Threading.Tasks;
using System.Windows;
using System.Windows.Controls;

using System.Collections;
using PersistedVariables;
using Amazon.Lambda;
using Amazon.Lambda.Model;
using System.IO;
using System.Runtime.Serialization.Json;
using System.Dynamic;

namespace ChatRoomWPF
{
    /// <summary>
    /// Interaction logic for MainWindow.xaml
    /// </summary>
    public partial class MainWindow : Window
    {
        /*
         * We initially start with 
         */
        public MainWindow()
        {
            InitializeComponent();
            HideControls();
            PrintPosts();
        }

        enum Month { Jan = 1, Feb, Mar, Apr, May, Jun, Jul, Aug, Sep, Oct, Nov, Dec };
        private string accessToken;                     // Access token for user when signed in
        private string signedin_user;                   // Name of user when signed in
        private string timestamp;                       // Timestamp of selected message in list of posts
        private string selected_name;                   // Name of user for selected message in list of posts (so we don't bother calling DeletePost if it's not the user's post)
        private string refreshToken;                    // The refresh token ???
        private int postsToGet = 200;                   // Max # of posts to display
        private SortedList<string, GridItem> download;  // The list of posts

        // Use the DataContractJsonSerializer class to parse the JSON, 
        // which requires the GetPostsResult class and all it's underlying classes to define the JSON data.
        public SortedList<string, GridItem> ProcessPostDataJson(string result)
        {
            download = new SortedList<string, GridItem>();
            object[] cols = new object[4];
            ArrayList rawData = new ArrayList();
            result = result.Replace("Content-Type", "contentType");  //Problem with illegal characters
            GetPostsResult postsResults = new GetPostsResult();
            MemoryStream ms = new MemoryStream(Encoding.UTF8.GetBytes(result));
            DataContractJsonSerializer ser = new DataContractJsonSerializer(postsResults.GetType());
            postsResults = ser.ReadObject(ms) as GetPostsResult;
            PostDataArray[] data = postsResults.body.data;
            string[] row = new string[4];
            foreach (PostDataArray pda in data)
            {
                cols[0] = pda.Alias.S.ToString().PadRight(30);
                cols[1] = pda.Timestamp.S;
                cols[2] = GetDate(cols[1].ToString()).PadRight(30);
                cols[3] = pda.Message.S.ToString().PadRight(150);
                row = new string[] { cols[0].ToString(), cols[1].ToString(), cols[2].ToString(), cols[3].ToString() };
                rawData.Add(row);
                download.Add(cols[1].ToString(), (new GridItem()
                {
                    User = cols[0].ToString(),
                    Timestamp = cols[1].ToString(),
                    Date = cols[2].ToString(),
                    Message = cols[3].ToString()
                }));
            }
            return download;
        }

        public void PrintPosts()
        {
            using (AmazonLambdaClient client = new AmazonLambdaClient())
            {
                GetPostsRequest postsRequest = new GetPostsRequest()
                { SortBy = "timestamp", SortOrder = "descending", PostsToGet = postsToGet };
                string payload = GetJson(typeof(GetPostsRequest), postsRequest);

                InvokeRequest iRequest = new InvokeRequest()
                {
                    FunctionName = "GetPosts",
                    Payload = payload
                };
                InvokeResponse response = client.Invoke(iRequest);
                var sr = new StreamReader(response.Payload);
                string result = sr.ReadToEnd();
                if (result.Contains("\"result\":\"failure\""))
                {
                    GetPostsFail(result);
                    return;
                }
                                
                

                download = ProcessPostDataJson(result);
                IList<GridItem> items = download.Values;
                dataGrid.ItemsSource = items;
                dataGrid.ColumnWidth = DataGridLength.SizeToCells;
                int itemCount = dataGrid.Items.Count;
                dataGrid.ScrollIntoView(dataGrid.Items[itemCount - 1]);
            }

        }

        private void GetPostsFail(string result)
        {
            result = result.Replace("Content-Type", "contentType");
            PostsFailure deserializedError = new PostsFailure();
            MemoryStream ms = new MemoryStream(Encoding.UTF8.GetBytes(result));
            DataContractJsonSerializer ser = new DataContractJsonSerializer(deserializedError.GetType());
            deserializedError = ser.ReadObject(ms) as PostsFailure;
            error.Text = deserializedError.body.error;
            error.SetValue(VisibilityProperty, Visibility.Visible);
            errorFlag.SetValue(VisibilityProperty, Visibility.Visible);
        }

        protected string GetJson(Type type, Object obj)
        {
            MemoryStream stream1 = new MemoryStream();
            DataContractJsonSerializer ser = new DataContractJsonSerializer(type);
            ser.WriteObject(stream1, obj);
            stream1.Position = 0;
            StreamReader sr = new StreamReader(stream1);
            return sr.ReadToEnd();
        }

        public string GetDate(string timestamp)
        {
            Int64 thedatetime = Int64.Parse(timestamp);
            System.DateTime dtobj = new DateTime(1970, 1, 1, 0, 0, 0, 0, System.DateTimeKind.Utc);
            dtobj = dtobj.AddSeconds(thedatetime).ToLocalTime();

            // Monday through Sunday:
            string dow = dtobj.DayOfWeek.ToString();

            // Month, where 1 represents January
            int month = dtobj.Month;

            // Year
            string year = dtobj.Year.ToString();

            // Date of month 1-whatever
            string day = dtobj.Day.ToString();

            // Time (note that hours are in 24-hour format) as hours:minutes:seconds:
            string time = dtobj.Hour + ":" + dtobj.Minute + ":" + dtobj.Second;

            Month mo = (Month)month;

            string date = mo + " " + day + " " + year + " " + time;
            return date;
        }

        public string CleanupJson(string input)
        {
            string result = input.Replace("{\"", " ");
            result = result.Replace("\"", " ");
            result = result.Replace("}", " ");
            result = result.Replace("S : ", "^");
            result = result.Replace(",", "");
            result = result.Replace("[", "");
            result = result.Replace("Alias :", "");
            result = result.Replace("Message :", "");
            result = result.Replace("Timestamp :", "");
            return result;
        }

        private void SetPostVisibility(Visibility v)
        {
            messageLabel.SetValue(VisibilityProperty, v);
            message.SetValue(VisibilityProperty, v);
            postMessage.SetValue(VisibilityProperty, v);
            delete.SetValue(VisibilityProperty, v);
            delete_account.SetValue(VisibilityProperty, v);
        }

        private void MakePostInvisible()
        {
            SetPostVisibility(Visibility.Hidden);
        }

        private void MakePostVisible()
        {
            SetPostVisibility(Visibility.Visible);
        }

        private void HideControls()
        {
            confirmLabel.SetValue(VisibilityProperty, Visibility.Hidden);
            email.SetValue(VisibilityProperty, Visibility.Hidden);
            addUser.SetValue(VisibilityProperty, Visibility.Hidden);
            delete.SetValue(VisibilityProperty, Visibility.Hidden);
            delete_account.SetValue(VisibilityProperty, Visibility.Hidden);
            errorFlag.SetValue(VisibilityProperty, Visibility.Hidden);
            error.SetValue(VisibilityProperty, Visibility.Hidden);
            confirmLabel.SetValue(VisibilityProperty, Visibility.Hidden);
            email.SetValue(VisibilityProperty, Visibility.Hidden);
            emailLabel.SetValue(VisibilityProperty, Visibility.Hidden);
            messageLabel.SetValue(VisibilityProperty, Visibility.Hidden);
            message.SetValue(VisibilityProperty, Visibility.Hidden);
            postMessage.SetValue(VisibilityProperty, Visibility.Hidden);
            codeLabel.SetValue(VisibilityProperty, Visibility.Hidden);
            confirmCode.SetValue(VisibilityProperty, Visibility.Hidden);
            confirm.SetValue(VisibilityProperty, Visibility.Hidden);
            newPwdLabel.SetValue(VisibilityProperty, Visibility.Hidden);
            newPwd.SetValue(VisibilityProperty, Visibility.Hidden);
            newConfirmLabel.SetValue(VisibilityProperty, Visibility.Hidden);
            newConfirmCode.SetValue(VisibilityProperty, Visibility.Hidden);
            updatePwd.SetValue(VisibilityProperty, Visibility.Hidden);

            // Set global vars
            accessToken = "";                     // Access token for user when signed in
            signedin_user = "";                   // Name of user when signed in
            timestamp = "";                       // Timestamp of selected message in list of posts
            selected_name = "";                   // Name of user for selected message in list of posts (so we don't bother calling DeletePost if it's not the user's post)
            refreshToken = "";
        }

        private void SetNewUserVisibility(Visibility v)
        {
            confirmLabel.SetValue(VisibilityProperty, v);
            email.SetValue(VisibilityProperty, v);
            addUser.SetValue(VisibilityProperty, v);
            confirmLabel.SetValue(VisibilityProperty, v);
            emailLabel.SetValue(VisibilityProperty, v);
        }

        private void MakeNewUserVisible()
        {
            SetNewUserVisibility(Visibility.Visible);
        }

        private void MakeNewUserInvisible()
        {
            SetNewUserVisibility(Visibility.Hidden);
        }

        private void SetConfirmVisibility(Visibility v)
        {
            codeLabel.SetValue(VisibilityProperty, v);
            email.SetValue(VisibilityProperty, v);
            emailLabel.SetValue(VisibilityProperty, v);
            addUser.SetValue(VisibilityProperty, v);
            confirmCode.SetValue(VisibilityProperty, v);
            confirmLabel.SetValue(VisibilityProperty, v);
            confirm.SetValue(VisibilityProperty, v);
        }

        private void MakeConfirmInvisible()
        {
            SetConfirmVisibility(Visibility.Hidden);
        }

        private void MakeConfirmVisible()
        {
            SetConfirmVisibility(Visibility.Visible);
        }

        private void ProcessLogInFail(string result)
        {
            Failure deserializedError = new Failure();
            MemoryStream ms = new MemoryStream(Encoding.UTF8.GetBytes(result));
            DataContractJsonSerializer ser = new DataContractJsonSerializer(deserializedError.GetType());
            deserializedError = ser.ReadObject(ms) as Failure;
            error.Text = deserializedError.body.error.code + ": " + deserializedError.body.error.message;
            error.SetValue(VisibilityProperty, Visibility.Visible);
            errorFlag.SetValue(VisibilityProperty, Visibility.Visible);
        }

        private void dataGrid_SelectionChanged(object sender, SelectionChangedEventArgs e)
        {
            // Clear any old error message
            error.SetValue(VisibilityProperty, Visibility.Visible);
            errorFlag.SetValue(VisibilityProperty, Visibility.Hidden);
            error.Text = "";

            // Store timestamp and user name from selected post
            DataGrid dg = (DataGrid)sender;
            GridItem gi = (GridItem)dg.SelectedItem;

            // gi is null once we delete the item
            if (gi == null)
            {
                timestamp = "";
                selected_name = "";
            }
            else
            {
                timestamp = gi.Timestamp.Trim();
                selected_name = gi.User.Trim();
            }
        }

        // Sign in with Lambda function SignInCognitoUser
        private void logIn_Click(object sender, RoutedEventArgs e)
        {
            {
                StartCognitoLogin start = new StartCognitoLogin()
                { UserName = userName.Text, Password = pin.Password };
                string payload = GetJson(typeof(StartCognitoLogin), start);
                using (AmazonLambdaClient iClient = new AmazonLambdaClient())
                {
                    error.SetValue(VisibilityProperty, Visibility.Hidden);
                    errorFlag.SetValue(VisibilityProperty, Visibility.Hidden);
                    InvokeRequest iRequest = new InvokeRequest()
                    {
                        FunctionName = "SignInCognitoUser",
                        Payload = payload
                    };
                    var response = iClient.Invoke(iRequest);
                    if (response.HttpStatusCode == System.Net.HttpStatusCode.OK)
                    {
                        if (null != response && response.StatusCode == 200)
                        {
                            var sr = new StreamReader(response.Payload);
                            string result = sr.ReadToEnd();
                            if (result.Contains("failure"))
                            {
                                ProcessLogInFail(result);
                                return;
                            }
                            result = result.Replace("Content-Type", "contentType");
                            OutPut deserializedUser = new OutPut();
                            MemoryStream ms = new MemoryStream(Encoding.UTF8.GetBytes(result));
                            DataContractJsonSerializer ser = new DataContractJsonSerializer(deserializedUser.GetType());
                            deserializedUser = ser.ReadObject(ms) as OutPut;
                            string status = deserializedUser.statusCode;
                            string contentType = deserializedUser.headers.contentType;
                            Data data = deserializedUser.body.data;
                            AuthenticationResult aR = data.AuthenticationResult;
                            accessToken = aR.AccessToken;
                            refreshToken = aR.RefreshToken;
                            ms.Close();
                            //return deserializedUser;
                            if (result.Contains("success"))
                            {
                                MakePostVisible();
                                MakeConfirmInvisible();
                                postMessage.Focus();

                                // Cache user name
                                signedin_user = start.UserName;

                                return;
                            }

                            else
                            {
                                error.Text = "Add User failed";
                                error.SetValue(VisibilityProperty, Visibility.Visible);
                                errorFlag.SetValue(VisibilityProperty, Visibility.Visible);
                            }
                        }
                    }
                }
            }
        }

        private void initiateNewUser_Click(object sender, RoutedEventArgs e)
        {
            MakePostInvisible();
            MakeNewUserVisible();
        }

        private void addUser_Click(object sender, RoutedEventArgs e)
        {
            StartAddCognitoUser addUser = new StartAddCognitoUser()
            { UserName = userName.Text, Password = pin.Password, Email = email.Text };

            string payload = GetJson(typeof(StartAddCognitoUser), addUser);

            using (AmazonLambdaClient iClient = new AmazonLambdaClient())
            {
                InvokeRequest iRequest = new InvokeRequest()
                {
                    FunctionName = "StartAddingPendingCognitoUser",
                    Payload = payload
                };
                var response = iClient.Invoke(iRequest);
                if (response.HttpStatusCode == System.Net.HttpStatusCode.OK)
                {
                    if (null != response && response.StatusCode == 200)
                    {
                        var sr = new StreamReader(response.Payload);
                        string result = sr.ReadToEnd();
                        if (result.Contains("success"))
                        {
                            error.Text = "Pending enail confirmation";
                            error.SetValue(VisibilityProperty, Visibility.Visible);
                            errorFlag.SetValue(VisibilityProperty, Visibility.Hidden);
                            MakeConfirmVisible();
                            return;
                        }

                        else
                        {
                            ProcessLogInFail(result);

                        }
                    }
                }
            }
        }

        private void confirm_Click(object sender, RoutedEventArgs e)
        {
            using (AmazonLambdaClient iClient = new AmazonLambdaClient())
            {
                FinishAddCognitoUser finish = new FinishAddCognitoUser()
                { UserName = userName.Text, ConfirmationCode = confirmCode.Text };

                string payload = GetJson(typeof(FinishAddCognitoUser), finish);

                InvokeRequest iRequest = new InvokeRequest()
                {
                    FunctionName = "FinishAddingPendingCognitoUser",
                    Payload = payload
                };

                var response = iClient.Invoke(iRequest);

                if (response.HttpStatusCode == System.Net.HttpStatusCode.OK)
                {
                    if (null != response && response.StatusCode == 200)
                    {
                        var sr = new StreamReader(response.Payload);
                        string result = sr.ReadToEnd();

                        if (result.Contains("success"))
                        {
                            error.Text = "New user confirmed";
                            error.SetValue(VisibilityProperty, Visibility.Visible);
                            errorFlag.SetValue(VisibilityProperty, Visibility.Hidden);
                            initiateNewUser.SetValue(VisibilityProperty, Visibility.Hidden);
                            MakeConfirmInvisible();
                            logIn.Focus();

                            return;
                        }
                        else
                        {
                            error.Text = "Add User failed";
                            error.SetValue(VisibilityProperty, Visibility.Visible);
                            errorFlag.SetValue(VisibilityProperty, Visibility.Visible);
                        }
                    }
                    else
                    {
                        error.Text = "Add User failed";
                        error.SetValue(VisibilityProperty, Visibility.Visible);
                        errorFlag.SetValue(VisibilityProperty, Visibility.Visible);
                    }
                }
            }
        }

        private void postMessage_Click(object sender, RoutedEventArgs e)
        {
            AddPost post = new AddPost()
            { AccessToken = accessToken, Message = message.Text };

            string payload = GetJson(typeof(AddPost), post);

            using (AmazonLambdaClient iClient = new AmazonLambdaClient())
            {
                InvokeRequest iRequest = new InvokeRequest()
                {
                    FunctionName = "AddPost",
                    Payload = payload
                };
                var response = iClient.Invoke(iRequest);
                if (response.HttpStatusCode == System.Net.HttpStatusCode.OK)
                {
                    if (null != response && response.StatusCode == 200)
                    {
                        var sr = new StreamReader(response.Payload);
                        string result = sr.ReadToEnd();

                        if (result.Contains("success"))
                        {
                            error.Text = "";
                            error.SetValue(VisibilityProperty, Visibility.Hidden);
                            errorFlag.SetValue(VisibilityProperty, Visibility.Hidden);
                            //dataGrid.Items.Clear();
                            PrintPosts();
                            return;
                        }
                        else
                        {
                            if (result.Contains("error"))
                            {

                                string errMsg = result.Substring(result.IndexOf("error") + 8);
                                errMsg = errMsg.Substring(0, errMsg.IndexOf("\""));
                                error.Text = errMsg;
                                error.SetValue(VisibilityProperty, Visibility.Visible);
                                errorFlag.SetValue(VisibilityProperty, Visibility.Visible);
                            }
                            else
                            {
                                error.Text = "Add User failed";
                                error.SetValue(VisibilityProperty, Visibility.Visible);
                                errorFlag.SetValue(VisibilityProperty, Visibility.Visible);
                            }
                        }
                    }
                }
            }
        }

        // Start getting new password using Lambda function StartChangingForgottenCognitoUserPassword
        private void forgotPwd_Click(object sender, RoutedEventArgs e)
        {
            Username newUserName = new Username() { UserName = userName.Text };

            string payload = GetJson(typeof(Username), newUserName);

            using (AmazonLambdaClient iClient = new AmazonLambdaClient())
            {
                InvokeRequest iRequest = new InvokeRequest()
                {
                    FunctionName = "StartChangingForgottenCognitoUserPassword",
                    Payload = payload
                };

                var response = iClient.Invoke(iRequest);

                if (response.HttpStatusCode == System.Net.HttpStatusCode.OK)
                {
                    if (null != response && response.StatusCode == 200)
                    {
                        var sr = new StreamReader(response.Payload);
                        string result = sr.ReadToEnd();

                        if (result.Contains("success"))
                        {
                            error.SetValue(VisibilityProperty, Visibility.Visible);
                            errorFlag.SetValue(VisibilityProperty, Visibility.Hidden);
                            error.Text = "Confirmation code sent to address on file";
                            MakePostInvisible();
                            newPwdLabel.SetValue(VisibilityProperty, Visibility.Visible);
                            newPwd.SetValue(VisibilityProperty, Visibility.Visible);
                            newConfirmLabel.SetValue(VisibilityProperty, Visibility.Visible);
                            newConfirmCode.SetValue(VisibilityProperty, Visibility.Visible);
                            updatePwd.SetValue(VisibilityProperty, Visibility.Visible);
                            logIn.SetValue(VisibilityProperty, Visibility.Hidden);
                            return;
                        }
                        else
                        {
                            if (result.Contains("failure"))
                            {
                                ProcessLogInFail(result);
                            }
                            else
                            {
                                error.Text = "New password failed";
                                error.SetValue(VisibilityProperty, Visibility.Visible);
                                errorFlag.SetValue(VisibilityProperty, Visibility.Visible);
                            }
                        }
                    }
                }
            }
        }

        // Finish re-setting password using Lambda function FinishChangingForgottenCognitoUserPassword
        private void updatePwd_Click(object sender, RoutedEventArgs e)
        {
            using (AmazonLambdaClient iClient = new AmazonLambdaClient())
            {
                FinishForgottenPassword finish = new FinishForgottenPassword()
                { UserName = userName.Text, ConfirmationCode = newConfirmCode.Text, NewPassword = newPwd.Text };

                string payload = GetJson(typeof(FinishForgottenPassword), finish);

                InvokeRequest iRequest = new InvokeRequest()
                {
                    FunctionName = "FinishChangingForgottenCognitoUserPassword",
                    Payload = payload
                };

                var response = iClient.Invoke(iRequest);

                if (response.HttpStatusCode == System.Net.HttpStatusCode.OK)
                {
                    if (null != response && response.StatusCode == 200)
                    {
                        var sr = new StreamReader(response.Payload);
                        string result = sr.ReadToEnd();

                        if (result.Contains("success"))
                        {
                            error.Text = "New password confirmed";
                            error.SetValue(VisibilityProperty, Visibility.Visible);
                            errorFlag.SetValue(VisibilityProperty, Visibility.Hidden);
                            MakeConfirmInvisible();
                            initiateNewUser.SetValue(VisibilityProperty, Visibility.Hidden);
                            MakePostVisible();
                            newPwdLabel.SetValue(VisibilityProperty, Visibility.Hidden);
                            newPwd.SetValue(VisibilityProperty, Visibility.Hidden);
                            newPwdLabel.SetValue(VisibilityProperty, Visibility.Hidden);
                            newConfirmCode.SetValue(VisibilityProperty, Visibility.Hidden);
                            newConfirmLabel.SetValue(VisibilityProperty, Visibility.Hidden);
                            updatePwd.SetValue(VisibilityProperty, Visibility.Hidden);
                            messageLabel.SetValue(VisibilityProperty, Visibility.Hidden);
                            message.SetValue(VisibilityProperty, Visibility.Hidden);
                            postMessage.SetValue(VisibilityProperty, Visibility.Hidden);
                            delete.SetValue(VisibilityProperty, Visibility.Hidden);
                            addUser.SetValue(VisibilityProperty, Visibility.Hidden);
                            forgotPwd.SetValue(VisibilityProperty, Visibility.Hidden);
                            logIn.SetValue(VisibilityProperty, Visibility.Visible);
                            logIn.Focus();
                            pin.Password = newPwd.Text;
                            initiateNewUser.SetValue(VisibilityProperty, Visibility.Visible);
                            return;
                        }

                        if (result.Contains("failure"))
                        {
                            ProcessLogInFail(result);
                        }
                        else
                        {
                            error.Text = "Update Password failed";
                            error.SetValue(VisibilityProperty, Visibility.Visible);
                            errorFlag.SetValue(VisibilityProperty, Visibility.Visible);
                        }
                    }
                    else
                    {
                        error.Text = "Update Password failed";
                        error.SetValue(VisibilityProperty, Visibility.Visible);
                        errorFlag.SetValue(VisibilityProperty, Visibility.Visible);
                    }
                }
            }
        }

        // Delete user's account using Lambda function DeleteCognitoUser
        private void delete_account_Click(object sender, RoutedEventArgs e)
        {
            // Make sure they are logged in
            if (signedin_user == null || signedin_user == "")
            {
                error.SetValue(VisibilityProperty, Visibility.Visible);
                errorFlag.SetValue(VisibilityProperty, Visibility.Hidden);
                error.Text = "You must be signed in to delete your account";

                return;
            }

            // DeleteCognitoUser takes one arg, the access token
            DeleteAccount account = new DeleteAccount()
            { AccessToken = accessToken };

            string payload = GetJson(typeof(DeleteAccount), account);

            using (AmazonLambdaClient iClient = new AmazonLambdaClient())
            {
                InvokeRequest iRequest = new InvokeRequest()
                {
                    FunctionName = "DeleteCognitoUser",
                    Payload = payload
                };
                var response = iClient.Invoke(iRequest);
                if (response.HttpStatusCode == System.Net.HttpStatusCode.OK)
                {
                    if (null != response && response.StatusCode == 200)
                    {
                        var sr = new StreamReader(response.Payload);
                        string result = sr.ReadToEnd();

                        if (result.Contains("success"))
                        {
                            HideControls();
                            error.SetValue(VisibilityProperty, Visibility.Visible);
                            errorFlag.SetValue(VisibilityProperty, Visibility.Hidden);
                            error.Text = "Account deleted";

                            // Clear user name, password, etc.
                            userName.Text = "";
                            pin.Password = "";
                            selected_name = "";
                            accessToken = "";
                            timestamp = "";
                        }
                    }
                }
            }
        }

        // Delete post using Lambda function DeletePost
        private void delete_Click(object sender, RoutedEventArgs e)
        {
            // Clear any old message/error text
            this.message.Text = "";
            error.SetValue(VisibilityProperty, Visibility.Visible);
            errorFlag.SetValue(VisibilityProperty, Visibility.Hidden);
            error.Text = "";

            // DeletePost takes two args: access token and timestamp
            // access token is set when user is logged in;
            // timestamp and selected_user are set when user selects a post

            if (selected_name != signedin_user)
            {
                error.Text = "Not your post!";

                return;
            }

            DeletePost post = new DeletePost()
            { AccessToken = accessToken, TimestampOfPost = timestamp };

            string payload = GetJson(typeof(DeletePost), post);

            using (AmazonLambdaClient iClient = new AmazonLambdaClient())
            {
                InvokeRequest iRequest = new InvokeRequest()
                {
                    FunctionName = "DeletePost",
                    Payload = payload
                };

                var response = iClient.Invoke(iRequest);

                if (response.HttpStatusCode == System.Net.HttpStatusCode.OK)
                {
                    if (null != response && response.StatusCode == 200)
                    {
                        var sr = new StreamReader(response.Payload);
                        string result = sr.ReadToEnd();

                        if (result.Contains("success"))
                        {
                            error.Text = "Post deleted";

                            // Show new list of posts
                            PrintPosts();

                            // Clear timestamp and selected_user so we don't accidently cache them
                            timestamp = "";
                            selected_name = "";

                            return;
                        }
                        else
                        {
                            if (result.Contains("failure"))
                            {
                                ProcessLogInFail(result);
                            }
                            else
                            {
                                error.Text = "Post could not be deleted";
                            }
                        }
                    }
                }
            }

            // Clear timestamp and selected_user so we don't accidently cache them
            timestamp = "";
            selected_name = "";
        }
    }

    public class GetPostsRequest
    {
        public string SortBy;
        public string SortOrder;
        public int PostsToGet;
    }

    public class GridItem
    {
        public string User { get; set; }
        public string Timestamp { get; set; }
        public string Date { get; set; }
        public string Message { get; set; }
    }

    public class AuthenticationResult
    {
        public string AccessToken;
        public string ExpiresIn;
        public string TokenType;
        public string RefreshToken;
        public string IdToken;
    }

    public class ChallengeParameters
    {

    }

    public class Username
    {
        public string UserName { get; set; }
    }

    public class Data
    {
        public ChallengeParameters ChallengeParameters;
        public AuthenticationResult AuthenticationResult;
    }

    public class Body
    {
        public string result;
        public Data data;
    }

    public class FailBody
    {
        public string result;
        public Error error;
    }

    public class PostsFailure
    {
        public string statusCode;
        public Headers headers;
        public PostsFailBody body;
    }

    public class PostsFailBody
    {
        public string result;
        public string error;
    }

    public class Error
    {
        public string message;
        public string code;
        public string time;
        public string requestId;
        public string statusCode;
        public string retryable;
        public string retryDelay;
    }

    public class FinishForgottenPassword
    {
        public string UserName;
        public string ConfirmationCode;
        public string NewPassword;
    }

    public class Headers
    {
        public string contentType;
    }

    public class OutPut
    {
        public string statusCode;
        public Headers headers;
        public Body body;
    }

    public class GetPostsResult
    {
        public string statusCode;
        public Headers headers;
        public PostBody body;
    }

    public class PostBody
    {
        public string result;
        public PostDataArray[] data;
    }

    public class PostDataArray
    {
        public Alias Alias;
        public Timestamp Timestamp;
        public Message Message;
    }

    public class Message
    {
        public string S;
    }

    public class Alias
    {
        public string S;
    }

    public class Timestamp
    {
        public string S;
    }

    public class Failure
    {
        public string statusCode;
        public Headers headers;
        public FailBody body;
    }

    public class FinishChangeForgottenPassword
    {
        public string UserName;
        public string ConfirmationCode;
        public string NewPassword;
    }

    public class StartAddCognitoUser
    {
        public string UserName;
        public string Password;
        public string Email;
    }

    public class FinishAddCognitoUser
    {
        public string UserName;
        public string ConfirmationCode;
    }

    public class StartChangeForgottenPassword
    {
        public string UserName;
    }

    public class StartCognitoLogin
    {
        public string UserName;
        public string Password;
    }

    public class AddPost // FinishCognitoLogin
    {
        public string AccessToken;
        public string Message;
    }

    public class Item
    {
        public string  User { get; set; }
        public string Timestamp { get; set; }
        public string Date { get; set; }
        public string Message { get; set; }
    }

    public class DeletePost
    {
        public string AccessToken { get; set; }
        public string TimestampOfPost { get; set; }
    }

    public class DeleteAccount
    {
        public string AccessToken { get; set; }
    }
}
