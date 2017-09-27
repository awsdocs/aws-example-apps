package com.amazonaws.samples;

import com.amazonaws.services.lambda.invoke.LambdaFunction;

public interface ChatService {
    @LambdaFunction(functionName="GetPosts")
    GetPostsResponse getPosts (GetPostsRequest input);
    
    @LambdaFunction(functionName="SignInCognitoUser")
    SignInResponse signIn (SignInRequest input);
    
    @LambdaFunction(functionName="AddPost")
    ChatAppDefaultResponse addPost (AddPostRequest input);
    
    @LambdaFunction(functionName="StartAddingPendingCognitoUser")
    ChatAppResponse registerUser (RegisterUserRequest input);
    
    @LambdaFunction(functionName="FinishAddingPendingCognitoUser")
    ChatAppResponse verifyPendingUser (VerifyUserRequest input);
    
    @LambdaFunction(functionName="StartChangingForgottenCognitoUserPassword")
    ChatAppResponse startResetPw (ResetPasswordRequest input);
    
    @LambdaFunction(functionName="FinishChangingForgottenCognitoUserPassword")
    ChatAppResponse finishResetPw (ResetPasswordRequest input);
    
    @LambdaFunction(functionName="DeleteCognitoUser")
    ChatAppDefaultResponse deleteAccount (DeleteAccountRequest input);
    
    @LambdaFunction(functionName="DeletePost")
    ChatAppDefaultResponse deletePost (DeletePostRequest input);
}
