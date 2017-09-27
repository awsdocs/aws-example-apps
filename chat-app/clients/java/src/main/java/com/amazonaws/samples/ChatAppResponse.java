package com.amazonaws.samples;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;

@JsonIgnoreProperties(ignoreUnknown = true)
public class ChatAppResponse {
    private ChatAppResponseBody body;
    private Long statusCode;
    
    public ChatAppResponseBody getBody() { return body; }
    public void setBody(ChatAppResponseBody value) { body = value; }
    
    public Long getStatusCode() { return statusCode; }
    public void setStatusCode(Long value) { statusCode = value; }
}