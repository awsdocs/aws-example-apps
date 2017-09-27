package com.amazonaws.samples;

import java.io.IOException;
import java.io.InputStream;

import java.util.*;
import java.text.*;

import com.amazonaws.util.StringUtils;

public class ChatClient {
    
    private static Long POSTS_TO_GET_DEFAULT = 100L;
    private static Long postsToGet = POSTS_TO_GET_DEFAULT; //default is 100 
    private static String currentUser;
    private static Boolean signedIn = false;
    private static Boolean registeringNewUser = false;
    private static Boolean resettingPassword = false;
    private static String token = "";
    private static String signInOut = "2: Sign in \r\n";
    private static String registerPrompt = "3: Register as a new user \r\n";
    private static String resetPasswordPrompt = "4: Reset password (you must be signed in) \r\n";

    public static void main(String[] args) throws Exception {        
        String inputValue = "";
        Boolean keepGoing = true;
        
        Scanner scanner = new Scanner(System.in);
        postsToGet = getPropValues();
        
        while (keepGoing) {
            String menu = "\r\nEnter a value from 1 to 7 to perform an action or q/Q to quit: \r\n"
                + "1: List last " + postsToGet + " posts \r\n"
                + signInOut
                + registerPrompt
                + resetPasswordPrompt
                + "5: Post a message (you must be signed in) \r\n"
                + "6: Delete your account (you must be signed in) \r\n"
                + "7: Delete a post (you must be signed in and it must be your post) \r\n"
                + "q (or Q): Quit\r\n";
            
            if (!StringUtils.isNullOrEmpty(currentUser)) {
                menu = menu + "\r\nSigned in as " + currentUser;
            }
            
            System.out.println(menu);
            inputValue = scanner.nextLine();    
            
            String username="", pw="", email="";
            
            switch (inputValue) {
                case "1": //get posts
                    try {
                        ChatMessage[] messages = ChatLib.getPosts(postsToGet);
                        printPosts(messages);
                    }
                    catch (ChatExceptions ex) {
                        System.out.println(ex.getMessage());
                    }
                    break;
                case "2": //if signed in, sign out
                    if (signedIn) {
                        currentUser = "";
                        signedIn = false;
                        token = "";
                        signInOut = "2: Sign in \r\n";
                        System.out.println("You are logged out.");
                    }
                    else {
                        System.out.println("Enter Username:");
                        username = scanner.nextLine();
                        
                        System.out.println("Enter password: ");
                        pw = scanner.nextLine();
                        
                        try {
                            token = ChatLib.signIn(username, pw);
                            signedIn = true;
                            currentUser = username;
                            signInOut = "2: Sign out \r\n";
                        } catch (ChatExceptions ex) {
                            System.out.println(ex.getMessage());
                        }
                    }
                    break;
                case "3": //register new user    
                    if (registeringNewUser){
                        System.out.println("Enter Username:");
                        username = scanner.nextLine();
                        
                        System.out.println("Enter verification code for " + username + ":");
                        String verifyCode = scanner.nextLine();
                        
                        try {
                            ChatLib.verifyUser(username, verifyCode);
                            registeringNewUser = false;
                            registerPrompt = "3: Register as a new user \r\n";
                            System.out.println("User, " + username + " verification successful.");
                        } catch (ChatExceptions ex) {
                            System.out.println(ex.getMessage());
                        }
                    }
                    else {
                        System.out.println("Enter Username:");
                        username = scanner.nextLine();
                        
                        System.out.println("Enter password: ");
                        pw = scanner.nextLine();
                        
                        System.out.println("Enter email: ");
                        email = scanner.nextLine();
                                                
                        try {
                            ChatLib.registerUser(username, pw, email);
                            registerPrompt = "3: Finish registering as a new user \r\n";
                            registeringNewUser = true;
                            System.out.println("User " + username + " registration started.");
                            System.out.println("Please check email for verification code to continue registration.");
                        } catch (ChatExceptions ex) {
                            System.out.println(ex.getMessage());
                        }
                    }
                    
                    break;
                case "4": //reset password
                    if (signedIn && !StringUtils.isNullOrEmpty(currentUser)) {
                        if (resettingPassword) {
                            System.out.println("Enter confirmation code:");
                            String confCode = scanner.nextLine();
                            
                            System.out.println("Enter new password:");
                            String newPw = scanner.nextLine();
                            
                            try {
                                ChatLib.resetPassword(currentUser, confCode, newPw);
                                resetPasswordPrompt = "4: Reset password (you must be signed in) \r\n";
                                resettingPassword = false;
                                System.out.println("You've successfully reset your password.");
                            }
                            catch (ChatExceptions ex) {
                                System.out.println(ex.getMessage());
                            }
                        }
                        else {                            
                            try {
                                ChatLib.resetPasswordRequest(currentUser);
                                resetPasswordPrompt = "4: Finish resetting password (you must be signed in) \r\n";
                                resettingPassword = true;
                                System.out.println("Please check your email for verification code to continue with password change.");
                            }
                            catch (ChatExceptions ex) {
                                System.out.println(ex.getMessage());
                            }
                        }
                    }
                    else {
                        System.out.println("You must be signed in to change password.");
                    }
                    
                    break;
                case "5": //post message
                    if (signedIn) {
                        System.out.println("Enter message");
                        String message = scanner.nextLine();
                        
                        try {
                            ChatLib.postMessage(token, message);
                            System.out.println("Your message has been posted.");
                            ChatMessage[] posts = ChatLib.getPosts(postsToGet);
                            printPosts(posts);
                        } catch (ChatExceptions ex) {
                            System.out.println(ex.getMessage());
                        }
                    }
                    else {
                        System.out.println("You must be signed in to post a message.");
                    }
                    
                    
                    break;
                case "6": //delete account
                    if (signedIn && !StringUtils.isNullOrEmpty(currentUser)) {
                        try {
                            System.out.println("Press enter to confirm deleting account " + currentUser);
                            scanner.nextLine();

                            ChatLib.deleteAccount(token);
                            System.out.println("Account " + currentUser + " deleted.");
                            currentUser = "";
                            signedIn = false;
                            token = "";
                            signInOut = "2: Sign in \r\n";
                            
                        } catch (ChatExceptions ex) {
                            System.out.println(ex.getMessage());
                        }
                    }
                    else {
                        System.out.println("You must be signed in to delete an account.");
                    }
                    
                    break;
                case "7": //delete post
                    if (signedIn) {
                        System.out.println("Enter the post ID:");
                        String postId = scanner.nextLine();

                        try {
                            ChatLib.deletePost(token, postId);
                            System.out.println("Post " + postId + " deleted.");
                        } catch (ChatExceptions ex) {
                            System.out.println(ex.getMessage());
                        }
                    }
                    else {
                        System.out.println("You must be signed in to delete a post.");
                    }
                    
                    break;
                case "q":
                case "Q":
                    keepGoing = false;
                    break;
                default: 
                    System.out.println("Invalid option. Please choose from the menu.");
                    break;
            }
        }
        
        scanner.close();
        System.out.println("Thanks for chatting, goodbye...");
    }
    
    private static Long getPropValues() throws IOException {
        InputStream inputStream = null;
        Long maxMsgs = POSTS_TO_GET_DEFAULT;
        
        try {
            Properties prop = new Properties();
            String propFileName = "/config.properties";
 
            inputStream = ChatClient.class.getResourceAsStream(propFileName);
            prop.load(inputStream);
  
            //get the property value and print it out
            maxMsgs = Long.parseLong(prop.getProperty("maxMsgs"));
 
        } catch (Exception e) {
            System.out.println("Exception: " + e);
        } finally {
            if (inputStream != null) {
                inputStream.close();
            }
        }
        
        return maxMsgs;
    }
    
    private static void printPosts(ChatMessage[] messages) {
        
        for (ChatMessage message : messages)
        {
            String timestampStr = getDateTime(message.getTimestamp().getValue());
            System.out.println(message.getAlias().getValue() + "@" + timestampStr + "     <" + message.getTimestamp().getValue() + ">");
            System.out.println(message.getMessage().getValue());
            System.out.println();
        }
    }
    
    private static String getDateTime(String epochTime) {
        Date date = new Date(Long.parseLong(epochTime)*1000);
        DateFormat format = new SimpleDateFormat("MM/dd/yyyy HH:mm:ss");
        String formatted = format.format(date);
        return formatted;
    }
}

