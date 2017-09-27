package com.amazonaws.samples;

import com.fasterxml.jackson.annotation.JsonProperty;

public class DeleteAccountRequest {

    private String accessToken;
    
    @JsonProperty("AccessToken")
    public String getAccessToken() { return accessToken; }
    public void setAccessToken(String value) { accessToken = value; }
}
