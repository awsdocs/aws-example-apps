# AWS SDK Docs Chat App in Go

This folder contains the Go source code of a command line app that uses the Lambda
functions in
*../../setup/lambda* to implement a simple chat app.

## Version Info

The Go source was developed on Go v1.8 using the AWS SDK for Go v1.8.21.

## Configuring the App

You can modify the following entries in *conf.json*:

* `Region` - Defines the default region, currently **us-west-2**.
* `Timezone` - Defines the default time zone, currently **UTC**.
* `MaxMessages`- Defines the number of most-recent messages to download, currently
**20**.

## Command Line Args

You can modify the following settings from the command line,
overriding those set in *conf.json*.

| Command | Option     | Description |
| ------- | ---------- | ----------------------------------------------- |
| **-t**  | *TIMEZONE* | Changes timezone to *TIMEZONE* (not implemented) |
| **-r**  | *REGION*   | Changes region to *REGION* |
| **-n**  | *MAXMSGS*  | Changes maxMsgs to *MAXMSGS* |
| **-d**  | | Enables debugging (emits out a lot of info) |
| **-h**  | | Displays help and quits |

## Running the App

Use the following command.

`go run PostApp.go`

## Workflow

1. Present the user with options.
2. Read the input.
3. Call the associated Lambda function.
4. Get the response and update the display as needed.
5. Repeat steps 2-4 until input == [q | Q].
