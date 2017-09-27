package com.amazonaws.samples;

import com.amazonaws.services.lambda.AWSLambdaClientBuilder;
import com.amazonaws.services.lambda.invoke.LambdaInvokerFactory;

public class ChatLib {

    //create a lambda client with the chat srv class
    private final static ChatService chatSrv = LambdaInvokerFactory.builder()
            .lambdaClient(AWSLambdaClientBuilder.defaultClient())
            .build(ChatService.class);
    
    public static ChatMessage[] getPosts(Long postsToGet) throws ChatExceptions {
        System.out.println("getting posts...");
        
        //create input json
        GetPostsRequest input = new GetPostsRequest();
        input.setSortBy("timestamp");
        input.setSortOrder("ascending");
        input.setPostsToGet(postsToGet);
        
        //get output json and evaluate
        GetPostsResponse output = chatSrv.getPosts(input);

        if (output.getStatusCode() != 200) {
            //print error from lambda and bail
            String errorMsg = output.getBody().getError();
            throw new ChatExceptions ("Could not get posts: " + errorMsg);
        } else {
            return output.getBody().getData();
        }
    }
    
    public static String signIn(String username, String pw) throws ChatExceptions {
        System.out.println("signing in...");
        
        //create input json
        SignInRequest input = new SignInRequest();
        input.setUserName(username);
        input.setPassword(pw);
        
        //get output json and evaluate
        SignInResponse output = chatSrv.signIn(input);

        if (output.getStatusCode() != 200) {
            //print error from lambda and bail
            String errorMsg = output.getBody().getError().getMessage();
            throw new ChatExceptions ("Could not sign in user " + username + ": " + errorMsg);
        } else {
            return output.getBody().getData().getAuthenticationResult().getAccessToken();
        } 
    }

    public static void registerUser(String username, String password, String email) throws ChatExceptions{
        System.out.println("registering you...");
        
        //create input json
        RegisterUserRequest input = new RegisterUserRequest();
        input.setUserName(username);
        input.setPassword(password);
        input.setEmail(email);
        
        //get output json and evaluate
        ChatAppResponse output = chatSrv.registerUser(input);

        if (output.getStatusCode() != 200) {
            //print error from lambda and bail
            String errorMsg = output.getBody().getError().getMessage();
            throw new ChatExceptions ("Could not register user: " + errorMsg);
        } 
    }
    
    public static void verifyUser(String username, String confCode) throws ChatExceptions {
        System.out.println("verifying you...");
        
        //create input json
        VerifyUserRequest input = new VerifyUserRequest();
        input.setUserName(username);
        input.setConfirmationCode(confCode);
        
        //get output json and evaluate
        ChatAppResponse output = chatSrv.verifyPendingUser(input);

        if (output.getStatusCode() != 200) {
            //print error from lambda and bail
            String errorMsg = output.getBody().getError().getMessage();
            throw new ChatExceptions ("Could not verify user: " + errorMsg);
        }
    }
    
    public static void resetPasswordRequest(String username) throws ChatExceptions {
        System.out.println("Submitting reset password request");
        
        //create input json
        ResetPasswordRequest input = new ResetPasswordRequest();
        input.setUserName(username);
        
        //get output json and evaluate
        ChatAppResponse output = chatSrv.startResetPw(input);

        if (output.getStatusCode() != 200) {
            //print error from lambda and bail
            String errorMsg = output.getBody().getError().getMessage();
            throw new ChatExceptions ("Could not start resetting password: " + errorMsg);
        } 
    }
    
    public static void resetPassword(String username, String confCode, String newPw) throws ChatExceptions {
        System.out.println("Changing password...");
        
        //create input json
        ResetPasswordRequest input = new ResetPasswordRequest();
        input.setUserName(username);
        input.setConfirmationCode(confCode);
        input.setNewPassword(newPw);
        
        //get output json and evaluate
        ChatAppResponse output = chatSrv.finishResetPw(input);

        if (output.getStatusCode() != 200) {
            //print error from lambda and bail
            String errorMsg = output.getBody().getError().getMessage();
            throw new ChatExceptions ("Could not start resetting password: " + errorMsg);
        }
    }
    
    public static void postMessage(String token, String message) throws ChatExceptions {
        System.out.println("posting your message...");
        
        //create input json
        AddPostRequest input = new AddPostRequest();
        input.setAccessToken(token);
        input.setMessage(message);
        
        //get output json and evaluate
        ChatAppDefaultResponse output = chatSrv.addPost(input);

        if (output.getStatusCode() != 200) {
            //print error from lambda and bail
            String errorMsg = output.getBody().getError();
            throw new ChatExceptions ("Could not post your message: " + errorMsg);
        } 
    }
    
    public static void deleteAccount(String token) throws ChatExceptions{
        System.out.println("deleting your account...");
        
        //create input json
        DeleteAccountRequest input = new DeleteAccountRequest();
        input.setAccessToken(token);
        
        //get output json and evaluate
        ChatAppDefaultResponse output = chatSrv.deleteAccount(input);

        if (output.getStatusCode() != 200) {
            //print error from lambda and bail
            String errorMsg = output.getBody().getError();
            throw new ChatExceptions ("Could not delete user: " + errorMsg);
        } 
    }
    
    public static void deletePost(String token, String postId) throws ChatExceptions {
        System.out.println("deleting your post...");
        
        //create input json
        DeletePostRequest input = new DeletePostRequest();
        input.setAccessToken(token);
        input.setTimestampOfPost(postId);
        
        //get output json and evaluate
        ChatAppDefaultResponse output = chatSrv.deletePost(input);

        if (output.getStatusCode() != 200) {
            //print error from lambda and bail
            String errorMsg = output.getBody().getError();
            throw new ChatExceptions ("Could not delete post: " + errorMsg);
        } 
    }
}
