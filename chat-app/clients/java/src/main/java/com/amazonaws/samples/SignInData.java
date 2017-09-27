package com.amazonaws.samples;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.annotation.JsonProperty;

@JsonIgnoreProperties(ignoreUnknown = true)
public class SignInData {

    private AuthenticationResult authResult;
    
    @JsonProperty("AuthenticationResult")
    public AuthenticationResult getAuthenticationResult() { return authResult; }
    public void setAuthenticationResult(AuthenticationResult value) { authResult = value; }
}
