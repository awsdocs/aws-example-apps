# AWS SDK Docs Chat App in Ruby

This folder contains the Ruby source code of a command-line app that uses the Lambda functions in
*../../setup/lambda* to implement a simple Chat app.

## Version Info

The Ruby source was developed on Ruby v2.3 using the AWS SDK for Ruby v2.9.

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

`ruby PostApp.rb`

## Workflow

1. Presents user with options
2. Reads input
3. Calls associated lambda function
4. Gets response and updates display as recessary
5. Repeats steps 2-4 until input == [q | Q]
