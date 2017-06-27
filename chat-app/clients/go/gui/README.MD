# AWS SDK Docs Chat App in Go with Templates

This folder contains the Go source code of a GUI app (using templates)
that uses the Lambda functions in
*../../setup/lambda* to implement a simple Chat app.

## Version Info

The Go source was developed on Go v1.8 using the AWS SDK for Go v1.8.21.

## Configuring the App

You can modify the following entries in *conf.json*:

* `Region` defines the default region, currently **us-west-2**.
* `Timezone` defines the default timezone, currently **UTC**
* `MaxMessages` defines the number of most-recent messages to download, currently **20**
* `RefreshSeconds` defines the interval, in seconds, between refreshing the list of posts, currently **30**
* `Debug` defines whether to emit information about what's going on under the covers, currently **false**

## Command-line options

You can use the following options from the command line,
overriding those set in *conf.json*:

| Command | Option     | Description |
| ------- | ---------- | ----------------------------------------------- |
| **-r**  | *REGION*   | changes Region to *REGION* |
| **-t**  | *TIMEZONE* | changes Timezone to *TIMEZONE* (not implemented) |
| **-n**  | *MAXMSGS*  | changes MaxMessages to *MAXMSGS* |
| **-f**  | *REFRESH*  | changes RefreshSeconds to *REFRESH* (not implemented) |
| **-d**  | | enables debugging |
| **-h**  | | displays help and quits |

## Running the app

`go run main.go [OPTION+]`

## Workflow

1. Presents list of posts and forms for:
   * Logging in
   * Registering
   * Resetting your password
2. If you log in, register, or reset your password,
   you can:
   * Post a message
   * Logout
   * Delete your account
3. If you logout or delete your account,
   you are taken back to step 1.
