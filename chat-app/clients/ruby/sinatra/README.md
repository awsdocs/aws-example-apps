# AWS SDK Docs Chat App in Ruby

This folder contains the Ruby source code of a Sinatra app that uses the Lambda functions in
*../../../setup/lambda* to implement a simple Chat app.

## Version Info

The Ruby source was developed on Ruby v2.3 using the AWS SDK for Ruby v2.9.

## Configuring the App

You can modify the following entries in *config.yml*:

* `region` defines the default region, currently **us-west-2**.
* `timezone` defines the default timezone, currently **UTC**
* `maxMsgs` defines the number of most-recent messages to download, currently **25**
* `debug` enables the display of settings (see *views/layout.erb*)

## Running the app locally

1. Start the server with `ruby myapp.rb`

2. Navigate to `http://localhost:4567`

## Workflow

1. User navigates to `http://localhost:4567`
   * They immediately see the **maxMsgs** latest posts
   * They can:
     * Login
     * Register
     * Reset their password
2. Once they login they can:
   * Post a message
   * Delete one of their posts
   * Delete their account, which takes them back to #1
   * Logout, which takes them back to #1
3. If they register successfully, they are automatically logged in (they got to #2)
