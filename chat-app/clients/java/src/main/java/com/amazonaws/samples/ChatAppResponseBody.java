package com.amazonaws.samples;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;

@JsonIgnoreProperties(ignoreUnknown = true)
public class ChatAppResponseBody {

    private String result;
    private ChatError error;

    public String getResult() { return result; }
    public void setResult(String value) { result = value; }
    
    public ChatError getError() { return error; }
    public void setError(ChatError value) { error = value; }
}