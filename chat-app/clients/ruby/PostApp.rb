=begin
Copyright 2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License").
You may not use this file except in compliance with the License.
A copy of the License is located at

http://aws.amazon.com/apache2.0/
=end

require 'aws-sdk'
require 'os'
require 'json'

if OS.windows?
  Aws.use_bundled_cert!
end

def debug_print(debug, s)
  if debug
    puts s
  end
end

USAGE = <<DOC

Usage: ruby PostApp.rb [-r REGION] [-t TIMEZONE] [-n MAXMSGS] [-d] [-h]

If REGION is not supplied, defaults to 'us-west-2'
If TIMEZONE is not supplied, defaults to 'UTC'
If MAXMSGS is not supplied, defaults to 20

-d     Display additional information
-h     Shows this message and quits

DOC

# Class to store credential profile information in
class CredsProfile
  # A credentials profile can have a name, access key, and secret access key
  def set_name(name)
    @name = name
  end

  def set_access_key(access_key)
    @access_key = access_key
  end

  def set_secret_key(secret_key)
    @secret_key = secret_key
  end

  def name
    @name
  end

  def access_key
    @access_key
  end

  def secret_key
    @secret_key
  end
end

def get_creds()
  begin
    filename = '/Users/soosung/.aws/credentials'

    creds = CredsProfile.new

    i = 1

    File.readlines(filename).each do |line|
      # Split input line at '='
      values = line.split("=")

      case values.length
        when 1
          k = values[0].strip
          creds.set_name(k)

        when 2
          k = values[0].strip
          v = values[1].strip

          if k == 'aws_access_key_id'
            creds.set_access_key(v)
          end

          if k == 'aws_secret_access_key'
            creds.set_secret_key(v)
          end
      end

      i += 1
    end

    creds
  end
rescue ArgumentError
  # Dir.home raises ArgumentError when ENV['home'] is not set
  nil
end

def hide_all_but_chars(str, show_last, c)
  if (str == nil) || (str == '')
    return nil
  end

  # Get length of string
  len = str.length    # 10
  min_len = len / 2

  if show_last > min_len   # Since 4 !> 5, showLast == 4
    show_last = min_len
  end

  rest_len = len - show_last   # 10 - 4 == 6

  # Last showLast chars of string
  str_end = str.slice(rest_len, (len - 1))  # "7890"

  # Pad first rest_len chars
  str_end.rjust(len, c)
end

class Message
  def initialize(name, time, msg)
    @name = name
    @time = time
    @msg  = msg
  end

  def name
    return @name
  end

  def time
    return @time
  end

  def msg
    return @msg
  end
end

def clear_screen()
  Gem.win_platform? ? (system 'cls') : (system 'clear')
end

def get_day_of_week(day)
  case day
    when 0
      return 'Sunday'
    when 1
      return 'Monday'
    when 2
      return 'Tuesday'
    when 3
      return 'Wednesday'
    when 4
      return 'Thursday'
    when 5
      return 'Friday'
    when 6
      return 'Saturday'
    else
      return 'unknown'
  end
end

def get_month(month)
  case month
    when 1
      return 'January'
    when 2
      return 'February'
    when 3
      return 'March'
    when 4
      return 'April'
    when 5
      return 'May'
    when 6
      return 'June'
    when 7
      return 'July'
    when 8
      return 'August'
    when 9
      return 'September'
    when 10
      return 'October'
    when 11
      return 'November'
    when 12
      return 'December'
    else
      return 'unknown'
  end
end

def get_date_time(debug, time, tz)
  date_time = Array.new

  # Construct date
  the_time = Time.at(time.to_i)

  ampm = 'am'

  hour = the_time.hour # 0-23

  if hour > 12
    hour = hour - 12
    ampm = 'pm'
  end

  dow = get_day_of_week(the_time.wday) # Monday is 1
  month = get_month(the_time.month) # January is 1

  date = "#{dow}, #{month} #{the_time.day}, #{the_time.year}"
  time = "#{hour}:#{the_time.min}:#{the_time.sec}"
  
  date_time << date
  date_time << time

  return date_time
end

def print_posts(debug, messages, timezone)
  # Rats, the newest post is first, I want it last
  rev_msgs = Array.new

  # Create new Message obj and add message to it
  messages.each do |m|
    new_msg = Message.new(m["Alias"]["S"], m["Timestamp"]["S"], m["Message"]["S"])
    rev_msgs << new_msg
  end

  # Reverse array so latest post is last
  rev_msgs.reverse!

  # Placeholder for current day
  day = ''

  rev_msgs.each do |msg|
    name = msg.name
    time = msg.time
    msg  = msg.msg

    # Returns a two-element string--the date and the time
    date_time = get_date_time(debug, time, timezone)

    if day != date_time[0]
      day = date_time[0]
      puts
      puts '=== ' + day + ' ==='
      puts
    end

    puts "#{name}@#{date_time[1]} <#{time}>:"
    puts msg
  end

end

def get_all_posts(debug, client, timezone, maxmsgs)
  my_hash = {:SortBy => 'timestamp', :SortOrder => 'descending', :PostsToGet => maxmsgs}
  payload = JSON.generate(my_hash)

  debug_print(debug, 'JSON request for posts:')
  debug_print(debug, payload)

  resp = client.invoke({
                           function_name: 'GetPosts',
                           invocation_type: 'RequestResponse',
                           log_type: 'None',
                           payload: payload
                       })

  debug_print(debug, 'JSON response for posts:')
  debug_print(debug, resp.payload.string)
  debug_print(debug, "")

  my_hash = JSON.parse(resp.payload.string) # , symbolize_names: true)

  # statusCode (should be 200)
  status_code = my_hash["statusCode"]
  debug_print(debug, "Status code: #{status_code}")

  if status_code == 200
    result = my_hash["body"]["result"]

    if result == "success"
      debug_print(debug, "Successfully retrieved posts")

      # Print out messages
      messages = my_hash["body"]["data"]

      return messages
    end
  end

  return nil
end

def notify_signed_in(registered)
  if registered
    puts 'You are already signed in, which means you are already registered.'
  else
    puts 'You are already signed in.'
  end

  puts 'If you want to register or sign in as another user,'
  puts 'you must sign out first and then register or sign in'
  puts
end

def get_string_value(prompt)
  puts prompt
  value = STDIN.gets.chomp

  return value
end

def sign_in_user(debug, client, name, password)
  results = Array.new

  my_hash = {:UserName => name, :Password => password}
  payload = JSON.generate(my_hash)

  debug_print(debug, 'JSON request for signing in user:')
  debug_print(debug, payload)

  resp = client.invoke({
                           function_name: 'SignInCognitoUser',
                           invocation_type: 'RequestResponse',
                           log_type: 'None',
                           payload: payload
                       })

  debug_print(debug, 'JSON response for signing in user:')
  debug_print(debug, resp.payload.string)
  debug_print(debug, "")

  my_hash = JSON.parse(resp.payload.string) # , symbolize_names: true)

  # statusCode (should be 200)
  status_code = my_hash["statusCode"]
  debug_print(debug, "Status code: #{status_code}")

  if status_code == 200
    result = my_hash["body"]["result"]

    if result == "success"
      debug_print(debug, "Successfully retrieved posts")

      # Return access token
      token = my_hash["body"]["data"]["AuthenticationResult"]["AccessToken"]

      results << token
      results << ""
    end
  else
    results << ""
    results << my_hash["body"]["error"]["message"]
  end

  return results
end

def post_from_signed_in_user(debug, client, token, msg)
  my_hash = {:AccessToken => token, :Message => msg}
  payload = JSON.generate(my_hash)

  debug_print(debug, 'JSON request for posting message:')
  debug_print(debug, payload)

  resp = client.invoke({
                           function_name: 'AddPost',
                           invocation_type: 'RequestResponse',
                           log_type: 'None',
                           payload: payload
                       })

  debug_print(debug, 'JSON response for posting message:')
  debug_print(debug, resp.payload.string)
  debug_print(debug, "")

  my_hash = JSON.parse(resp.payload.string) # , symbolize_names: true)

  # statusCode (should be 200)
  status_code = my_hash["statusCode"]
  debug_print(debug, "Status code: #{status_code}")

  if status_code == 200
    result = my_hash["body"]["result"]

    if result == "success"
      debug_print(debug, "Successfully posted message")
    end
  end

end

def start_register_user(debug, client, name, password, email)
  my_hash = {:UserName => name, :Password => password, :Email => email}
  payload = JSON.generate(my_hash)

  debug_print(debug, 'JSON request for adding Cognito user:')
  debug_print(debug, payload)

  resp = client.invoke({function_name: 'StartAddingPendingCognitoUser',
                        invocation_type: 'RequestResponse',
                        log_type: 'None',
                        payload: payload
                       })

  debug_print(debug, 'JSON response for adding Cognito user:')
  debug_print(debug, resp.payload.string)
  debug_print(debug, "")

  my_hash = JSON.parse(resp.payload.string) # , symbolize_names: true)

  # statusCode (should be 200)
  status_code = my_hash["statusCode"]
  debug_print(debug, "Status code: #{status_code}")

  if status_code == 200
    result = my_hash["body"]["result"]

    if result == "success"
      debug_print(debug, "Successfully started registering user")
      return ""
    end
  end

  return my_hash["body"]["error"]["message"]
end

def finish_register_user(debug, client, name, code)
  my_hash = {:UserName => name, :ConfirmationCode => code}
  payload = JSON.generate(my_hash)

  debug_print(debug, 'JSON request for finishing adding Cognito user:')
  debug_print(debug, payload)

  resp = client.invoke({
                           function_name: 'FinishAddingPendingCognitoUser',
                           invocation_type: 'RequestResponse',
                           log_type: 'None',
                           payload: payload
                       })

  debug_print(debug, 'JSON response for finishing adding Cognito user:')
  debug_print(debug, resp.payload.string)
  debug_print(debug, "")

  my_hash = JSON.parse(resp.payload.string) # , symbolize_names: true)

  # statusCode (should be 200)
  status_code = my_hash["statusCode"]
  debug_print(debug, "Status code: #{status_code}")

  if status_code == 200
    result = my_hash["body"]["result"]

    if result == "success"
      debug_print(debug, "Successfully registered user")
      return ""
    end
  end

  return my_hash["body"]["error"]["message"]
end

def start_reset_password(debug, client, name)
  my_hash = {:UserName => name}
  payload = JSON.generate(my_hash)

  debug_print(debug, 'JSON request for starting password reset:')
  debug_print(debug, payload)

  resp = client.invoke({
                           function_name: 'StartChangingForgottenCognitoUserPassword',
                           invocation_type: 'RequestResponse',
                           log_type: 'None',
                           payload: payload
                       })

  debug_print(debug, 'JSON response for starting password reset:')
  debug_print(debug, resp.payload.string)
  debug_print(debug, "")

  my_hash = JSON.parse(resp.payload.string) # , symbolize_names: true)

  # statusCode (should be 200)
  status_code = my_hash["statusCode"]
  debug_print(debug, "Status code: #{status_code}")

  if status_code == 200
    result = my_hash["body"]["result"]

    if result == "success"
      debug_print(debug, "Successfully starte password reset")
      return true
    end
  end

  return false
end

def finish_reset_password(debug, client, name, code, password)
  my_hash = {:UserName => name, :ConfirmationCode => code, :NewPassword => password}
  payload = JSON.generate(my_hash)

  debug_print(debug, 'JSON request for finishing password reset:')
  debug_print(debug, payload)

  resp = client.invoke({
                           function_name: 'FinishChangingForgottenCognitoUserPassword',
                           invocation_type: 'RequestResponse',
                           log_type: 'None',
                           payload: payload
                       })

  debug_print(debug, 'JSON response for finishing password reset:')
  debug_print(debug, resp.payload.string)
  debug_print(debug, "")

  my_hash = JSON.parse(resp.payload.string) # , symbolize_names: true)

  # statusCode (should be 200)
  status_code = my_hash["statusCode"]
  debug_print(debug, "Status code: #{status_code}")

  if status_code == 200
    result = my_hash["body"]["result"]

    if result == "success"
      debug_print(debug, "Successfully reset password")
      return true
    end
  end

  return false
end

def delete_post(debug, client, access_token, timestamp)
  my_hash = {:AccessToken => access_token, :TimestampOfPost => timestamp}
  payload = JSON.generate(my_hash)

  debug_print(debug, 'JSON request for deleting post:')
  debug_print(debug, payload)

  resp = client.invoke({
                           function_name: 'DeletePost',
                           invocation_type: 'RequestResponse',
                           log_type: 'None',
                           payload: payload
                       })

  debug_print(debug, 'JSON response for deleting post:')
  debug_print(debug, resp.payload.string)
  debug_print(debug, "")

  my_hash = JSON.parse(resp.payload.string) # , symbolize_names: true)

  # statusCode (should be 200)
  status_code = my_hash["statusCode"]
  debug_print(debug, "Status code: #{status_code}")

  if status_code == 200
    result = my_hash["body"]["result"]

    if result == "success"
      debug_print(debug, "Successfully deleted post")
      return ""
    end
  end

  if my_hash['body'] != nil
    if my_hash['body']['error'] != nil
      if my_hash['body']['error']['message'] != nil
        return my_hash['body']['error']['message']
      end
    end
  end

  return 'Error: ' + resp.payload.string

end

def delete_user_account(debug, client, access_token)
  my_hash = {:AccessToken => access_token}
  payload = JSON.generate(my_hash)

  debug_print(debug, 'JSON request for deleting user account:')
  debug_print(debug, payload)

  resp = client.invoke({
                           function_name: 'DeleteCognitoUser',
                           invocation_type: 'RequestResponse',
                           log_type: 'None',
                           payload: payload
                       })

  debug_print(debug, 'JSON response for deleting user account:')
  debug_print(debug, resp.payload.string)
  debug_print(debug, "")

  my_hash = JSON.parse(resp.payload.string) # , symbolize_names: true)

  # statusCode (should be 200)
  status_code = my_hash["statusCode"]
  debug_print(debug, "Status code: #{status_code}")

  if status_code == 200
    result = my_hash["body"]["result"]

    if result == "success"
      debug_print(debug, "Successfully deleted user account")
      return ""
    end
  end

  return my_hash["body"]["error"]["message"]
end


# main routime starts here
debug = false

# Get config values from conf.json
file = File.read 'conf.json'
data = JSON.parse(file)

region = data['Region']
timezone = data['Timezone']
max_msgs = data['MaxMessages']

i = 0

while i < ARGV.length
  case ARGV[i]
    when '-r'
      i += 1
      region = ARGV[i]

    when '-t'
      i += 1
      timezone = ARGV[i]

    when '-n'
      i += 1
      max_msgs = ARGV[i].to_i

    when '-d'
      debug = true

    when '-h'
      puts USAGE
      exit 0

    else
      puts 'Unrecognized option: ' + ARGV[i]
      puts USAGE
      exit 1
  end

  i += 1
end

cursor = '(anonymous)> '

# When false, stop the app
keep_going = true

# The name of the current user
user_name = ''

# True if signed in (required to post)
signed_in = false

creds = get_creds()

if debug
# Barf out user info
  puts('Access key ID: ' + hide_all_but_chars(creds.access_key, 4, '-'))
  puts('Secret key:    ' + hide_all_but_chars(creds.secret_key, 4, '-'))

  client = Aws::IAM::Client.new(region: region, access_key_id: creds.access_key, secret_access_key: creds.secret_key)
  iam = Aws::IAM::Resource.new(client: client)

  # Show user info
  user = iam.current_user
  name = user.user_name
  puts "User name:  #{name}"
  puts "    ARN:          #{user.arn}"
  arn_parts = user.arn.split(':')
  puts "    Account ID:   #{arn_parts[4]}"
  puts "    User ID:      #{user.user_id}"
  puts "    Sign-in link: https://#{arn_parts[4]}.signin.aws.amazon.com/console"
end

# Create Lambda service client
client = Aws::Lambda::Client.new(region: region, access_key_id: creds.access_key, secret_access_key: creds.secret_key)

# Replace scanner in args with:
# value = STDIN.gets.chomp

input_value = ''

# So we can adjust prompt for two-stage processes
# set to true once start* completes successfully; reset to false after finish* completes successfully
past_step1 = false
register_prompt = '3: Register as new user'
reset_password_prompt = '4: Reset password (you must be signed in)'

outcome = false
name = ''
password = ''
access_token = ''

while keep_going
  # Menu
  puts
  puts 'Enter a value between 1 and 8 to perform the indicated action or q (or Q) to quit:'
  puts
	puts '1: List all posts'
	puts '2: Sign in'
	puts register_prompt
	puts reset_password_prompt
	puts '5: Post a message (you must be signed in)'
  puts '6: Sign out'
  puts '7: Delete your account (you must be signed in)'
  puts '8: Delete a post (you must be signed in and it must be your post)'
	puts 'q (or Q): Quit'
  puts

  input_value = get_string_value(cursor)
  input_value.strip!

  # Don't clear the screen if we are quitting or debugging
  if input_value != 'q' && input_value != 'Q' && ! debug
    clear_screen
  end

  case input_value
    when '1'
      # list all posts
      debug_print(debug, 'Calling get_all_posts')
      posts = get_all_posts(debug, client, timezone, max_msgs)

      if posts == nil
        puts 'Could not retrieve any posts'
      else
        print_posts(debug, posts, timezone)
      end

    when '2'
      # Sign in
      debug_print(debug, 'Calling sign_in')

      if signed_in
        notify_signed_in(false)
      else
        # Get user name as we use it for the prompt
        name = get_string_value("Enter your user name")
        puts
        password = get_string_value("Enter your password")
        puts

        debug_print(debug, 'Calling sign_in_user')

        results = sign_in_user(debug, client, name, password)

        debug_print(debug, 'results[0] == ' + results[0])
        debug_print(debug, 'results[1] == ' + results[1])

        if results[0] == ""
          puts 'Could not sign in user'
          puts results[1]
        else
          access_token = results[0]
          signed_in = true
          cursor = '(' + name + ')> '
        end
      end

    when '3'
      # Register user

      if signed_in
        notify_signed_in(true)
        next
      end

      if past_step1
        # Finish registering
        code = get_string_value("Enter your confirmation code")

        debug_print(debug, 'Calling finish_register_user')

        outcome = finish_register_user(debug, client, name, code)

        if outcome == ""
          # Automatic sign in
          access_token = sign_in_user(debug, client, name, password)
          signed_in = true
          cursor = '(' + name + ')> '
          past_step1 = false
          register_prompt = '3: Register as new user'
        else
          puts 'Error registering, did you use the wrong code?'
          puts outcome
          next
        end
      else
        debug_print(debug, "Calling start_register_user")

        # Get their name and a password
        name = get_string_value('Enter your user name')
        puts
        password = get_string_value('Enter a password with at least 6 characters')
        puts

        if password.length < 6
          puts('You password is too short, try again')
          next
        end

        email = get_string_value('Enter your email address')
        puts

        outcome = start_register_user(debug, client, name, password, email)

        if outcome == ""
          # Change registration prompt
          register_prompt = '3: Finish registering (you still must sign in)'
          past_step1 = true
        else
          puts "Could not start registering user"
          puts outcome
        end
      end

    when '4'
      # Reset password
      if past_step1
        # Finish resetting password
        debug_print(debug, 'Finishing resetting password')
        debug_print(debug, "For user #{name}")

        code = get_string_value('Enter the confirmation code')
        password = get_string_value('Enter your new password')

        result = finish_reset_password(debug, client, name, code, password)

        if result
          debug_print(debug, "Successfully reset password")
          # user is signed in with new password
          user_name = name
          signed_in = true
          cursor = "(" + user_name + ")> "
          past_step1 = false
          reset_password_prompt = "4: Reset password"
        end
      else
        # Must be signed in to reset
        if ! signed_in
          puts 'You must be signed in to reset your password'
          next
        end

        result = start_reset_password(debug, client, user_name)

        if result
          past_step1 = true
          reset_password_prompt = '4: Finish resetting password'
        end
      end

    when '5'
      # Post message
      # Must be signed in
      if ! signed_in
        puts 'You must sign in to post'
        next
      end

      message = get_string_value('Enter a message to post:')
      post_from_signed_in_user(debug, client, access_token, message)

    when '6'
      # Sign out
      signed_in = false
      user_name = ''
      access_token = ''
      cursor = '(anonymous)> '

    when '7'
      # Delete account
      debug_print(debug, 'Deleting account')

      if ! signed_in
        puts 'You must sign in to delete your account'
        next
      end

      result = delete_user_account(debug, client, access_token)

      if result == ""
        signed_in = false
        user_name = ""
        access_token = ""
        cursor = "(anonymous)> "
        puts 'Your account has been deleted'
      else
        puts 'Could not delete account'
        puts result
      end

    when '8'
      # Delete post
      debug_print(debug, 'Deleting post')

      if ! signed_in
        puts 'You must sign in to delete a postt'
        next
      end

      timestamp = get_string_value('Enter the ID of the post to delete (the ID is the long number at the end of the first line):')

      result = delete_post(debug, client, access_token, timestamp)

      if result == ""
        puts 'The post has been deleted'
      else
        puts 'Could not delete post'
        puts result
      end

    when 'q'
      keep_going = false

    when 'Q'
      keep_going = false

    else
      puts 'Unrecognized option: ' + input_value
  end

end
