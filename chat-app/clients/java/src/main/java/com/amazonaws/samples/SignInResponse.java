package com.amazonaws.samples;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;

@JsonIgnoreProperties(ignoreUnknown = true)
public class SignInResponse {
    private SignInBody body;
    private Long statusCode;
    
    public SignInBody getBody() { return body; }
    public void setBody(SignInBody value) { body = value; }
    
    public Long getStatusCode() { return statusCode; }
    public void setStatusCode(Long value) { statusCode = value; }
}

