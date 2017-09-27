package com.amazonaws.samples;

import com.fasterxml.jackson.annotation.JsonProperty;

public class RegisterUserRequest {

    private String username;
    private String password;
    private String email;
    
    @JsonProperty("UserName")
    public String getUserName() { return username; }
    public void setUserName(String value) { username = value; }

    @JsonProperty("Password")
    public String getPassword() { return password; }
    public void setPassword(String value) { password = value; }
    
    @JsonProperty("Email")
    public String getEmail() { return email; }
    public void setEmail(String value) { email = value; }

}
