package com.amazonaws.samples;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;

@JsonIgnoreProperties(ignoreUnknown = true)
public class GetPostsResponse {
    private PostBody body;
    private Long statusCode;
    
    public PostBody getBody() { return body; }
    public void setBody(PostBody value) { body = value; }
    
    public Long getStatusCode() { return statusCode; }
    public void setStatusCode(Long value) { statusCode = value; }
}
