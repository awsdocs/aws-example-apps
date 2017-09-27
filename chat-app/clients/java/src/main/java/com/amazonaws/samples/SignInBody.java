package com.amazonaws.samples;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;

@JsonIgnoreProperties(ignoreUnknown = true)
public class SignInBody {

    private String result;
    private SignInData data;
    private ChatError error;

    public String getResult() { return result; }
    public void setResult(String value) { result = value; }
    
    public SignInData getData() { return data; }
    public void setData(SignInData value) { data = value; }
    
    public ChatError getError() { return error; }
    public void setError(ChatError value) { error = value; }
}

