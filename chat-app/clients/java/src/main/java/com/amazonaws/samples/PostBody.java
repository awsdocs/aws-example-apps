package com.amazonaws.samples;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;

@JsonIgnoreProperties(ignoreUnknown = true)
public class PostBody {

    private String result;
    private ChatMessage[] data;
    private String error;

    public String getResult() { return result; }
    public void setResult(String value) { result = value; }
    
    public ChatMessage[] getData() { return data; }
    public void setData(ChatMessage[] value) { data = value; }
    
    public String getError() { return error; }
    public void setError(String value) { error = value; }
}
