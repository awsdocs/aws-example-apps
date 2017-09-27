package com.amazonaws.samples;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.annotation.JsonProperty;

@JsonIgnoreProperties(ignoreUnknown = true)
public class ChatError {

    private String message;
    private String code;
    
    public String getMessage() { return message; }
    public void setMessage(String value) { message = value; }
    
    public String getCode() { return code; }
    public void setCode(String value) { code = value; }
    
}
