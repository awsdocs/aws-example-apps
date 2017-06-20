=begin
Copyright 2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License").
You may not use this file except in compliance with the License.
A copy of the License is located at

http://aws.amazon.com/apache2.0/
=end

require 'sinatra'
require 'sinatra/reloader' if development?

configure do
  enable :sessions

  set :username, ''
  set :email, ''
  set :password, ''
  set :code, ''
  set :token, ''
  set :token_short, ''
  set :status, 'Not logged in'
  set :item, ''
  set :debug, true
end

=begin
  PostsLib.rb contains most of the logic.

  Workflow:

  1. User hits /, which takes them to /start
     Their initial status is 'Not logged in'.
     Since they aren't logged in, they have four options:
     a. Don't do anything
        They can only see the posts.
        TBD: do we want to have a "refresh" option? Auto? Button?
     b. Log in with username and password.
        i.   Button -> posts username and password to /login
        ii.  /login calls log_in_user with username and password
        iii. If log_in_user returns '' (empty token),
             -> /home with status 'Not logged in' (start over).
             Otherwise we set the status to 'Logged in'
             save the status, username, and token in settings,
             and call home.erb with status and msg == 'You are now logged in'.
     c. Register with username, email address, and password
        i.   Button -> posts username, email address, and password to /register
        ii.  
        i. At /register enter confirmation code and click Submit button.
           Button -> calls  /home.
           If registration is successful, they go to / with status loggedin.
     d. Reset password with username field and Submit button.
        Button takes them to /reset.
        At /reset_password they see a form with two text fields and a button.
        i. Finish resetting password with confirmation code and new password.
           Button takes them to finish_reset.
           If resetting is successful, they go to / with status loggedin.
  2. Now that they are logged in, / shows N forms:
     a. Create post with a message field and Submit button.
        Button takes them to / with status loggedin.
     b. Logout with Submit button.
        Button takes them to / with status !loggedin.

  Note that every page displays the status at the top of the page as:
    Status: @status
=end

require './PostsLib.rb'

get '/' do
  @status = settings.status

  if @status  == 'Logged in'
    @msg = 'You are logged in'

    @posts = get_all_posts
    erb :home
  else
    @msg = 'You must be logged in (or registered, which automatically logs you in) before you can post, delete a post, or delete your account.'

    @posts = get_all_posts
    erb :start
  end

end

get '/about' do
  @title = 'About this web site'
  erb :about
end

get '/contact' do
  @title = 'Contact Info'
  erb :contact
end

post '/login' do
  @username = params[:username]
  @password = params[:password]
  @token = log_in_user(@username, @password)

  if @token == ''
    @status = 'Not logged in'
    @msg = 'Could not log in. Retry your username and password or return to the main page to register or reset password.'

    settings.status = 'Not logged in'

    @posts = get_all_posts
    erb :start
  else
    @status = 'Logged in'
    @msg = 'You are now logged in'

    settings.username = @username
    settings.token = @token
    settings.token_short = get_short_token(@token)
    settings.status = 'Logged in'

    @posts = get_all_posts
    erb :home
  end

end

=begin

Registration takes two steps:
1. Provide a name and email address
2. Sign in with the confirmation code sent to the user's email address

=end

post '/register' do
  if settings.status == 'Not logged in'
    # Our first submit (post from the /start page)
    @username = params[:username]
    @email = params[:email]
    @password = params[:password]

    if @username == ''
      @status = settings.status
      @msg = 'You forgot to supply a user name'

      @posts = get_all_posts
      erb :start
    end

    if @email == ''
      @status = settings.status
      @msg = 'You forgot to supply an email address'

      @posts = get_all_posts
      erb :start
    end

    if @password == ''
      @status = settings.status
      @msg = 'You forgot to supply a password'

      @posts = get_all_posts
      erb :start
    end

    @reply = start_register_user(@username, @password, @email)

    if @reply == 'success'
      @status = 'Finish registering'
      @msg = 'Registration under way'

      settings.status = @status
      settings.username = @username
      settings.password = @password
      settings.email = @email

      @posts = get_all_posts
      erb :register
    else
      @msg = 'Error registering: ' + @reply
      @status = 'Not logged in'

      settings.status = @status
      settings.username = ''
      settings.password = ''
      settings.email = ''

      # start all over
      @posts = get_all_posts
      erb :start
    end

  elsif settings.status = 'Finish registering'
    @username = settings.username
    @code = params[:code]
    @password = settings.password

    @token = finish_register_user(@username, @code, @password)

    if @token == ''
      @status = 'Not logged in'
      @msg = 'Sorry, could not register you'

      settings.username = ''
      settings.password = ''
      settings.status = @status

      # Start all over
      @posts = get_all_posts
      erb :start
    else
      @status = 'Logged in'
      @msg = 'You are now registered'

      settings.status = @status
      settings.username = @username
      settings.token = @token
      settings.token_short = get_short_token(@token)

      @posts = get_all_posts
      erb :home
    end

  end

end

post '/post' do
  @message = params[:message]
  @token = settings.token
  @msg = post_message(@token, @message)

  @posts = get_all_posts
  erb :home
end

post '/delete' do
  @msg_id = params[:message_value]
  @msg = delete_message(settings.token, @msg_id)
  @status = settings.status

  @posts = get_all_posts
  erb :home
end

post '/logout' do
  @msg = 'You are now logged out'
  @status = 'Not logged in'

  settings.username = ''
  settings.email = ''
  settings.password = ''
  settings.token = ''
  settings.token_short = ''
  settings.status = @status

  @posts = get_all_posts
  erb :start
end

post '/reset' do
  if settings.status = 'Not logged in'
    # Our first submit (post from the /start page)
    @username = params[:username]
    @reply = start_reset_password(@username)

    if @reply == 'success'
      @status = 'Resetting'
      @msg = 'Resetting password underway'

      settings.status = @status
      settings.username = @username

      @posts = get_all_posts
      erb :reset
    else
      @msg = 'Error resetting password: ' + @reply
      @status = 'Not logged in'

      settings.status = @status
      settings.username = '' # Shouldn't be necessary

      # start all over
      @posts = get_all_posts
      erb :start
    end

  elsif settings.status == 'Resetting'
    # Our second submit (post from the /reset page)
    @code = params[:code]
    @password = params[:password]
    @token = finish_reset_password(@name, @code, @password)

    if @token == ''
      @msg = 'Resetting your password failed, try again.'
      @status = 'Not logged in'

      settings.status = @status

      @posts = get_all_posts
      erb :start
    else
      @msg = 'You have successfully reset your password.'
      @status = 'Logged in'

      settings.status = @status

      @posts = get_all_posts
      erb :home
    end

  else
    # This shouldn't happen
  end

end

post '/unregister' do
  @token = settings.token

  @reply = delete_user_account(@token)

  if @reply == 'success'
    @msg = 'You have successfully unregistered'
    @status = 'Not logged in'

    settings.status = @status
    settings.username = ''
    settings.password = ''
    settings.token = ''
    settings.token_short = ''

    @posts = get_all_posts
    erb :start
  else
    @msg = 'Could not unregister your account, try again.'
    @status = settings.status

    @posts = get_all_posts
    erb :home
  end

end

not_found do
  @title = 'Link not found'
  erb :not_found
end
