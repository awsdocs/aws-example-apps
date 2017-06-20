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
require 'yaml'
require 'sinatra/base'

if OS.windows?
  Aws.use_bundled_cert!
end

class ConfigLoader
  def initialize
    load_config_data
  end

  def [](name)
    config_for(name)
  end

  def config_for(name)
    @config_data[name]
  end

  private

  def load_config_data
    config_file_path = 'config.yml'
    begin
      config_file_contents = File.read(config_file_path)
    rescue Errno::ENOENT
      $stderr.puts "missing config file"
      raise
    end
    @config_data = YAML.load(config_file_contents)
  end
end

class DefaultConfig
  CONFIG = ConfigLoader.new.config_for("Default")

  def initialize_config
    @region = CONFIG['region']
    @timezone = CONFIG['timezone']
    @maxmsgs = CONFIG['maxMsgs']
    @debug = CONFIG['debug']
  end

  def region
    @region
  end

  def timezone
    @timezone
  end

  def maxmsgs
    @maxmsgs
  end

  def debug
    @debug
  end
end

class Message
  def initialize(name, time, msg)
    @name = name
    @time = time
    @msg  = msg
  end

  def name
    @name
  end

  def time
    @time
  end

  def msg
    @msg
  end
end

class Post
  def initialize(time, msg, timestamp)
    @time = time
    @msg = msg
    @timestamp = timestamp
  end

  def time
    @time
  end

  def msg
    @msg
  end

  def timestamp
    @timestamp
  end
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

def get_date_time(time)
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

def get_region()
  config = DefaultConfig.new
  config.initialize_config

  config.region
end

def get_timezone()
  config = DefaultConfig.new
  config.initialize_config

  config.timezone
end

def get_debug()
  config = DefaultConfig.new
  config.initialize_config

  config.debug
end

def get_maxmsgs()
  config = DefaultConfig.new
  config.initialize_config

  config.maxmsgs
end

def get_client()
  region = get_region

  # Create Lambda service client
  Aws::Lambda::Client.new(region: region)
end

def get_all_posts()
  # List of posts
  posts = Array.new

  client = get_client

  maxmsgs = get_maxmsgs

  my_hash = {:SortBy => 'timestamp', :SortOrder => 'descending', :PostsToGet => maxmsgs}
  payload = JSON.generate(my_hash)

  resp = client.invoke({
                           function_name: 'GetPosts',
                           invocation_type: 'RequestResponse',
                           log_type: 'None',
                           payload: payload
                       })

  my_hash = JSON.parse(resp.payload.string) # , symbolize_names: true)

  # statusCode (should be 200)
  status_code = my_hash["statusCode"]

  if status_code == 200
    result = my_hash["body"]["result"]

    if result == "success"
      # Print out messages
      messages = my_hash["body"]["data"]

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
        date_time = get_date_time(time)

        if day != date_time[0]
          day = date_time[0]
          new_post = Post.new('=== ' + day + ' ===', '', 0)
          posts << new_post
        end

        new_post = Post.new("#{name}@#{date_time[1]}:", msg, time)
        posts << new_post
      end
    end
  end

  posts
end

def log_in_user(username, password)
  if username == nil || password == nil
    return ''
  end

  token = ''

  client = get_client

  my_hash = {:UserName => username, :Password => password}
  payload = JSON.generate(my_hash)

  resp = client.invoke({
                           function_name: 'SignInCognitoUser',
                           invocation_type: 'RequestResponse',
                           log_type: 'None',
                           payload: payload
                       })

  my_hash = JSON.parse(resp.payload.string)

  # statusCode (should be 200)
  if my_hash["statusCode"] == 200
    if my_hash["body"]["result"] == "success"
      # Return access token
      token = my_hash["body"]["data"]["AuthenticationResult"]["AccessToken"]
    end
  end

  token
end

def post_message(token, message)
  client = get_client

  my_hash = {:AccessToken => token, :Message => message}
  payload = JSON.generate(my_hash)

  resp = client.invoke({
                           function_name: 'AddPost',
                           invocation_type: 'RequestResponse',
                           log_type: 'None',
                           payload: payload
                       })

  my_hash = JSON.parse(resp.payload.string) # , symbolize_names: true)

  if my_hash['statusCode'] == 200
    if my_hash['body']['result'] == 'success'
      reply = 'Posted message'
    end
  else
    reply = 'Could not post message: ' + my_hash['body']['error']['message']
  end

  reply
end

def delete_message(token, msg_timestamp)
  if token == nil || token == ''
    return 'You must include a token to delete a message'
  end

  if msg_timestamp == nil || msg_timestamp == ''
    return 'You must select a message if you want to delete it'
  end

  client = get_client

  my_hash = {:AccessToken => token, :TimestampOfPost => msg_timestamp}
  payload = JSON.generate(my_hash)

  resp = client.invoke({
                           function_name: 'DeletePost',
                           invocation_type: 'RequestResponse',
                           log_type: 'None',
                           payload: payload
                       })

  my_hash = JSON.parse(resp.payload.string) # , symbolize_names: true)

  if my_hash["statusCode"] == 200
    if my_hash["body"]["result"] == "success"
      return 'Successfully deleted message'
    end
  end

  return 'Error trying to delete message: ' + my_hash["body"]["error"]["message"]
end

def delete_account(token)
  client = get_client

  my_hash = {:AccessToken => token}
  payload = JSON.generate(my_hash)

  resp = client.invoke({
                           function_name: 'DeleteCognitoUser',
                           invocation_type: 'RequestResponse',
                           log_type: 'None',
                           payload: payload
                       })

  my_hash = JSON.parse(resp.payload.string) # , symbolize_names: true)

  if my_hash["statusCode"] == 200
    if my_hash["body"]["result"] == "success"
      reply = 'Successfully deleted user account'
    end
  else
    reply = 'Could not delete the account: ' + my_hash['body']['error']['message']
  end

  reply
end

def start_reset_password(name)
  reply = ''

  client = get_client

  my_hash = {:UserName => name}
  payload = JSON.generate(my_hash)

  resp = client.invoke({
                           function_name: 'StartChangingForgottenCognitoUserPassword',
                           invocation_type: 'RequestResponse',
                           log_type: 'None',
                           payload: payload
                       })

  my_hash = JSON.parse(resp.payload.string)

  if my_hash["statusCode"] == 200
    if my_hash["body"]["result"] == "success"
      reply = 'success'
    end
  else
    reply = 'Error starting changing password: ' + my_hash['body']['error']['message']
  end

  reply
end

def finish_reset_password(name, code, password)
  token = ''

  client = get_client

  my_hash = {:UserName => name, :ConfirmationCode => code, :NewPassword => password}
  payload = JSON.generate(my_hash)

  resp = client.invoke({
                           function_name: 'FinishChangingForgottenCognitoUserPassword',
                           invocation_type: 'RequestResponse',
                           log_type: 'None',
                           payload: payload
                       })

  my_hash = JSON.parse(resp.payload.string)

  if my_hash["statusCode"] == 200
    if my_hash["body"]["result"] == "success"
      token = log_in_user(name, password)
    end
  end

  token
end

def delete_user_account(token)
  reply = ''

  client = get_client

  my_hash = {:AccessToken => token}
  payload = JSON.generate(my_hash)

  resp = client.invoke({
                           function_name: 'DeleteCognitoUser',
                           invocation_type: 'RequestResponse',
                           log_type: 'None',
                           payload: payload
                       })

  my_hash = JSON.parse(resp.payload.string) # , symbolize_names: true)

  if my_hash["statusCode"] == 200
    if my_hash["body"]["result"] == "success"
      reply = 'success'
    end
  else
    reply = 'Could not delete account: ' + my_hash['body']['error']['message']
  end

  reply
end

def start_register_user(name, password, email)
  if name == nil || name == ''
    return 'You must supply a user name to register'
  end

  if password == nil || password == ''
    return 'You must supply a password to register'
  end

  if email == nil || email == ''
    return 'You must supply an email address to register'
  end

  reply = ''

  client = get_client

  my_hash = {:UserName => name, :Password => password, :Email => email}
  payload = JSON.generate(my_hash)

  resp = client.invoke({function_name: 'StartAddingPendingCognitoUser',
                        invocation_type: 'RequestResponse',
                        log_type: 'None',
                        payload: payload
                       })

  my_hash = JSON.parse(resp.payload.string) # , symbolize_names: true)

  if my_hash['statusCode'] == 200
    if my_hash['body']['result'] == 'success'
      reply = 'success'
    end
  else
    reply = 'Could not start adding user: ' + my_hash['body']['error']['message']
  end

  reply
end

def finish_register_user(name, code, password)
  if name == nil || name == ''
    return 'You must supply a user name to register'
  end

  if code == nil || code == ''
    return 'You must supply a confirmation code to register'
  end

  if password == nil || password == ''
    return 'You must supply a password to register'
  end

  token = ''

  client = get_client


  my_hash = {:UserName => name, :ConfirmationCode => code}
  payload = JSON.generate(my_hash)

  resp = client.invoke({
                           function_name: 'FinishAddingPendingCognitoUser',
                           invocation_type: 'RequestResponse',
                           log_type: 'None',
                           payload: payload
                       })

  my_hash = JSON.parse(resp.payload.string) # , symbolize_names: true)

  if my_hash["statusCode"] == 200
    if my_hash["body"]["result"] == "success"
      token = log_in_user(name, password)
    end
  end

  token
end

def get_short_token(token)
  # Get first and last 4 chars of token
  first = token[0..3]
  last = token[(token.length() - 4),4]

  return first + '...' + last
end
