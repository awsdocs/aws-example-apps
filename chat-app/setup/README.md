# Chat app setup

Experiment with this example chat app to learn how to automate AWS services with code. 

This app features a single chat room and supports multiple participants. It allows participants to sign up and sign in to the chat room as well as read and add posts in the chat room. It uses Amazon DynamoDB, AWS Lambda, Amazon Cognito, and AWS Identity and Access Management (AWS IAM).

THIS EXAMPLE APP IS FOR LEARNING PURPOSES ONLY. DO NOT USE THIS EXAMPLE APP IN PRODUCTION ENVIRONMENTS.

## Scripts for the app ##

In this folder, find scripts you can run in your AWS account to create the app's required AWS resources. These include resources such as Amazon DynamoDB tables, AWS Lambda functions, Amazon Cognito user pools, and AWS IAM resource access policies.

The scripts are in the following subfolders:

- `cognito`: Scripts for creating the required Amazon Cognito user pool.
- `dynamodb`: Scripts for creating the required Amazon DynamoDB tables.
- `iam`: Scripts for creating the required AWS IAM resource access policies.
- `lambda`: Scripts for creating the required AWS Lambda functions. 

After you create the app's required AWS resources, you should be able to run the app's code on your local workstation with few or no modifications. The code is one level up in the `clients` subfolder, organized by programming language.

## We welcome feedback! ##

If you have a suggestion for improvement, please submit an issue by choosing **Issues, New Issue**.

If you find an error and have a fix, or want to contribute code, please submit a pull request by choosing **Pull requests, New pull request**.

