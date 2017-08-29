# AWS SDK Docs Chat App in Go with Templates

This folder contains the Go source code of a GUI app (using templates)
that uses the Lambda functions in
*../../setup/lambda* to implement a simple chat app.

## Version Info

The Go source was developed on Go v1.8 using the AWS SDK for Go v1.8.21.

## Configuring the App

You can modify the following entries in *conf.json*:

* `Region` - Defines the default region, currently **us-west-2**.
* `Timezone` - Defines the default time zone, currently **UTC**.
* `MaxMessages` - Defines the number of most recent messages to download, currently
**20**.
* `RefreshSeconds` - Defines the interval, in seconds, between refreshing the list
of posts, currently **30**.
* `Debug` - Defines whether to emit information about what's going on in the code,
currently **false**.

## Command Line Options

You can use the following options from the command line,
overriding those set in *conf.json*.

| Command | Option     | Description |
| ------- | ---------- | ----------------------------------------------- |
| **-r**  | *REGION*   | Changes Region to *REGION* |
| **-t**  | *TIMEZONE* | Changes Timezone to *TIMEZONE* (not implemented) |
| **-n**  | *MAXMSGS*  | Changes MaxMessages to *MAXMSGS* |
| **-f**  | *REFRESH*  | Changes RefreshSeconds to *REFRESH* (not implemented) |
| **-d**  | | Enables debugging |
| **-h**  | | Displays help and quits |

## Running the App

Use the following command.

`go run main.go [OPTION+]`

## Workflow

1. Present a list of posts and forms for:
   * Logging in.
   * Registering.
   * Resetting your password.
2. If you log in, register, or reset your password,
   you can:
   * Post a message.
   * Log out.
   * Delete your account.
3. If you log out or delete your account,
   you are taken back to step 1.
