package com.amazonaws.samples;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;

@JsonIgnoreProperties(ignoreUnknown = true)
public class ChatAppDefaultResponseBody {

    private String result;
    private String error;

    public String getResult() { return result; }
    public void setResult(String value) { result = value; }
    
    public String getError() { return error; }
    public void setError(String value) { error = value; }
}