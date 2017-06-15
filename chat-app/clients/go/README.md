# AWS SDK Docs Chat App in Go

This folder contains the Go source code of a command-line app that uses the Lambda functions in
*../../setup/lambda* to implement a simple Chat app.

## Version Info

The Go source was developed on Go v1.8 using the AWS SDK for Go v1.8.21.

## Configuring the App

You can modify the following entries in *conf.json*:

* `Region` defines the default region, currently **us-west-2**.
* `Timezone` defines the default timezone, currently **UTC**
* `MaxMessages` defines the number of most-recent messages to download, currently **20**

## Command-line args

You can modify the following settings from the command line,
overriding those set in *conf.json*:

| Command | Option     | Description |
| ------- | ---------- | ----------------------------------------------- |
| **-t**  | *TIMEZONE* | changes timezone to *TIMEZONE* (not implemented) |
| **-r**  | *REGION*   | changes region to *REGION* |
| **-n**  | *MAXMSGS*  | changes maxMsgs to *MAXMSGS* |
| **-d**  | | enables debugging (emits out a lot of info) |
| **-h**  | | displays help and quits |

## Running the app

`go run PostApp.go`

## Workflow

1. Presents user with options
2. Reads input
3. Calls associated lambda function
4. Gets response and updates display as recessary
5. Repeats steps 2-4 until input == [q | Q]
