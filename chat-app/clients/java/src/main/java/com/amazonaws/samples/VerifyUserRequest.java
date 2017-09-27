package com.amazonaws.samples;

import com.fasterxml.jackson.annotation.JsonProperty;

public class VerifyUserRequest {

    private String username;
    private String confCode;
    
    @JsonProperty("UserName")
    public String getUserName() { return username; }
    public void setUserName(String value) { username = value; }

    @JsonProperty("ConfirmationCode")
    public String getConfirmationCode() { return confCode; }
    public void setConfirmationCode(String value) { confCode = value; }

}