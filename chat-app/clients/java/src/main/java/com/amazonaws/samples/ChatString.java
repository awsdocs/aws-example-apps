package com.amazonaws.samples;

import com.fasterxml.jackson.annotation.JsonProperty;

public class ChatString {

    private String value;
    
    @JsonProperty("S")
    public String getValue() { return value; }
    public void setValue(String val) { value = val; }
}
