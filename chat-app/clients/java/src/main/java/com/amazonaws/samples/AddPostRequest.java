package com.amazonaws.samples;

import com.fasterxml.jackson.annotation.JsonProperty;

public class AddPostRequest {

    private String AccessToken;
    private String Message;
    
    @JsonProperty("Message")
    public String getMessage() { return Message; }
    public void setMessage(String value) { Message = value; }
    
    @JsonProperty("AccessToken")
    public String getAccessToken() { return AccessToken; }
    public void setAccessToken(String value) { AccessToken = value; } 
}
