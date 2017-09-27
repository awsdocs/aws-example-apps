package com.amazonaws.samples;

import com.fasterxml.jackson.annotation.JsonProperty;

public class DeletePostRequest {
    
    private String accessToken;
    private String timestampOfPost;
    
    @JsonProperty("AccessToken")
    public String getAccessToken() { return accessToken; }
    public void setAccessToken(String value) { accessToken = value; }
    
    @JsonProperty("TimestampOfPost")
    public String getTimestampOfPost() { return timestampOfPost; }
    public void setTimestampOfPost(String value) { timestampOfPost = value; }
}
