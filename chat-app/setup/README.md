# Chat app setup

Experiment with this example chat app to learn how to automate AWS services with code. 

This app features a single chat room and supports multiple participants. It allows participants to sign up and sign in to the chat room, and read and add posts in the chat room. It uses Amazon DynamoDB, AWS Lambda, Amazon Cognito, and AWS Identity and Access Management (IAM).

THIS EXAMPLE APP IS FOR LEARNING PURPOSES ONLY. DO NOT USE THIS EXAMPLE APP IN PRODUCTION ENVIRONMENTS.

## Scripts for the app ##

In this folder, find scripts to run in your AWS account to create the app's required AWS resources. These resources include DynamoDB tables, Lambda functions, Amazon Cognito user pools, and IAM resource access policies.

The scripts are in the following subfolders:
- `cognito`: A script for creating the required Amazon Cognito user pool.
- `dynamodb`: A script for creating the required DynamoDB table.
- `iam`: A script for creating the required IAM service role.
- `lambda`: Scripts for creating the required Lambda functions. 

After you create the app's required AWS resources, you should be able to run the app's code on your local workstation with very few modifications. The code is one level up in the `clients` subfolder, organized by programming language.

## Running the scripts ##

Follow these steps to run the scripts for the app. See [More info](#more-info) to learn more about using the AWS CloudFormation console, creating Amazon S3 buckets, and adding objects to S3 buckets.

**Step 1.** In the CloudFormation console, run these template files in the following order. You must create a separate stack for each template file, choose a meaningful value for **Stack name** for each stack, and accept the default settings on the stack's **Options** page.

1. `dynamodb/PostsTable.yaml`
2. `cognito/ChatRoomPool.yaml` - When the stack shows **CREATE_COMPLETE**, choose 
    the box next to the stack. On the **Outputs** tab, note the value for the **Keys** named 
    **UserPoolClientID** (for example, `8715f0bte9brn79tj9221j2lEX`), **UserPoolID** (for example, `us-west-2_oa6IReZEX`), and **UserPoolIDShortURL** 
    (for example, `https://cognito-idp.us-west-2.amazonaws.com/us-west-2_oa6IReZEX`). You need these values for later steps.
3. `iam/LambdaChatAppRole.yaml` - On the **Review** page, select **I acknowledge 
   that AWS CloudFormation might create IAM resources with custom names.** When the stack shows **CREATE_COMPLETE**, choose the box next to the stack. On the **Outputs** tab, note the value for the 
   **Key** named **RoleARN** (for example, `arn:aws:iam::YOUR_AWS_ACCOUNT_ID:role/LambdaChatAppRole`, where `YOUR_AWS_ACCOUNT_ID` is your 12-digit AWS account ID). You need this value for later steps.

**Step 2.** Unzip the following .zip file: `lambda/VerifyCognitoSignIn.zip`. Change the value of the `iss` variable in the `index.js` file to the **UserPoolIDShortURL** value you noted earlier. Save your changes, and 
re-ZIP the file. 

**Step 3.** Upload the following .zip files to a single Amazon S3 bucket. Note the bucket's name. You need this name for later steps. 

* `lambda/AddPost.zip`
* `lambda/DeleteCognitoUser.zip`
* `lambda/DeletePost.zip`
* `lambda/FinishAddingPendingCognitoUser.zip`
* `lambda/FinishChangingForgottenCognitoUserPassword.zip`
* `lambda/GetPosts.zip`
* `lambda/SignInCognitoUser.zip`
* `lambda/StartAddingPendingCognitoUser.zip`
* `lambda/StartChangingForgottenCognitoUserPassword.zip`
* The `VerifyCognitoSignIn.zip` file you changed in the previous step

**Step 4.** In the CloudFormation console, run these template files in the following 
order. You must create a separate stack for each template file. For each stack, choose a meaningful value for **Stack name**. 
For each stack, for **IAMRoleARN**, type the **RoleARN** value you noted earlier. For **S3BucketName**, type the bucket name you noted earlier. Accept the default settings on the stack's **Options** page.

* `lambda/StartAddingPendingCognitoUser.yaml`
* `lambda/FinishAddingPendingCognitoUser.yaml`
* `lambda/SignInCognitoUser.yaml`
* `lambda/VerifyCognitoSignIn.yaml`
* `lambda/StartChangingForgottenCognitoUserPassword.yaml`
* `lambda/FinishChangingForgottenCognitoUserPassword.yaml`
* `lambda/GetPosts.yaml`
* `lambda/AddPost.yaml`
* `lambda/DeletePost.yaml`
* `lambda/DeleteCognitoUser.yaml`

**Step 5.** In the Lambda console, change the value of `ClientId` in these functions' code to match the **UserPoolClientID** value you noted earlier.

* `StartAddingPendingCognitoUser`
* `FinishAddingPendingCognitoUser`
* `SignInCognitoUser`
* `StartChangingForgottenCognitoUserPassword`
* `FinishChangingForgottenCognitoUserPassword`

**Step 6.** In the Lambda console, change the value of `UserPoolID` in the `SignInCognitoUser` function's code 
to match the **UserPoolID** value you noted earlier.
  
**To edit a function's code** 

1. [Open the AWS Lambda console](https://console.aws.amazon.com/lambda).
2. In the navigation bar, choose the AWS Region where you created the function (for example, **US West (Oregon)**).
4. If you see **Get Started Now**, choose it.
5. In the navigation pane, you should see **Functions** is already chosen.
6. In the **Function name** column, choose the name of the function you want to edit (for example, **StartAddingPendingCognitoUser**).
7. The **Code** tab should already be chosen. Find the existing `ClientId` value in the code, and replace it with the **UserPoolClientID** value 
   you noted earlier.
8. Choose **Save**.

## More info

* [Creating a Stack on the AWS CloudFormation Console](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/cfn-console-create-stack.html) in the *AWS CloudFormation User Guide*
* [Create a Bucket](http://docs.aws.amazon.com/AmazonS3/latest/gsg/CreatingABucket.html) in the *Amazon Simple Storage Service (S3) Getting Started Guide*
* [Add an Object to a Bucket](http://docs.aws.amazon.com/AmazonS3/latest/gsg/PuttingAnObjectInABucket.html) in the *Amazon Simple Storage Service (S3) Getting Started Guide* 

## We welcome feedback! ##

If you have a suggestion for improvement, please submit an issue by choosing **Issues, New Issue**.

If you find an error and have a fix, or want to contribute code, please submit a pull request by choosing **Pull requests, New pull request**.