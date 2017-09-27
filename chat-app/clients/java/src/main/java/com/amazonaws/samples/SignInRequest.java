package com.amazonaws.samples;

import com.fasterxml.jackson.annotation.JsonProperty;

public class SignInRequest {
    
    private String username;
    private String password;
    
    @JsonProperty("UserName")
    public String getUserName() { return username; }
    public void setUserName(String value) { username = value; }

    @JsonProperty("Password")
    public String getPassword() { return password; }
    public void setPassword(String value) { password = value; }

}
