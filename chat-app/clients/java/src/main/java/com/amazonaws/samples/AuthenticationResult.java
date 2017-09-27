package com.amazonaws.samples;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.annotation.JsonProperty;

@JsonIgnoreProperties(ignoreUnknown = true)
public class AuthenticationResult {

    private String accessToken;
    private String refreshToken;
    private String idToken;
    
    @JsonProperty("AccessToken")
    public String getAccessToken() { return accessToken; }
    public void setAccessToken(String value) { accessToken = value; }
    
    @JsonProperty("RefreshToken")
    public String getRefreshToken() { return refreshToken; }
    public void setRefreshToken(String value) { refreshToken = value; }
    
    @JsonProperty("IdToken")
    public String getIdToken() { return idToken; }
    public void setIdToken(String value) { idToken = value; }
}
