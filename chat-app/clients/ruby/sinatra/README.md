# AWS SDK Docs Chat App in Sinatra

This folder contains the source code of a Sinatra app that uses the Lambda functions in
*../../../setup/lambda* to implement a simple chat app.

## Version Info

The Ruby source was developed on Ruby v2.3 using the AWS SDK for Ruby v2.9.

## Configuring the App

You can modify the following entries in *config.yml*:

* `region` - Defines the default region, currently **us-west-2**.
* `timezone` - Defines the default time zone, currently **UTC**.
* `maxMsgs` - Defines the number of most recent messages to download, currently **25**.
* `debug` - Enables the display of settings (see *views/layout.erb*).

## Running the App Locally

1. Start the server with `ruby myapp.rb`.

2. Navigate to `http://localhost:4567`.

## Workflow

1. The user navigates to `http://localhost:4567`.
   * They immediately see the latest posts (up to **maxMsgs**).
   * They can:
     * Log in.
     * Register.
     * Reset their password.
2. Once they log in, they can:
   * Post a message.
   * Delete one of their posts.
   * Delete their account, which takes them back to step 1.
   * Log out, which takes them back to step 1.
3. If they register successfully, they are automatically logged in (they go to step
2).
