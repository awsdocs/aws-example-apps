package com.amazonaws.samples;

import com.fasterxml.jackson.annotation.JsonProperty;

public class ResetPasswordRequest {

    private String username;
    private String confCode;
    private String newPassword;
    
    @JsonProperty("UserName")
    public String getUserName() { return username; }
    public void setUserName(String value) { username = value; }
    
    @JsonProperty("ConfirmationCode")
    public String getConfirmationCode() { return confCode; }
    public void setConfirmationCode(String value) { confCode = value; }
    
    @JsonProperty("NewPassword")
    public String getNewPassword() { return newPassword; }
    public void setNewPassword(String value) { newPassword = value; }
}
