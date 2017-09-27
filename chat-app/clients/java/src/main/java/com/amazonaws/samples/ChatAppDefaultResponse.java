package com.amazonaws.samples;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;

@JsonIgnoreProperties(ignoreUnknown = true)
public class ChatAppDefaultResponse {
    private ChatAppDefaultResponseBody body;
    private Long statusCode;
    
    public ChatAppDefaultResponseBody getBody() { return body; }
    public void setBody(ChatAppDefaultResponseBody value) { body = value; }
    
    public Long getStatusCode() { return statusCode; }
    public void setStatusCode(Long value) { statusCode = value; }
}