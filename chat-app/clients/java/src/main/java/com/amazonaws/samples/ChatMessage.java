package com.amazonaws.samples;

import com.fasterxml.jackson.annotation.JsonProperty;

public class ChatMessage {
    private ChatString Message;
    private ChatString Alias;
    private ChatString Timestamp;
    
    @JsonProperty("Message")
    public ChatString getMessage() { return Message; }
    public void setMessage(ChatString value) { Message = value; }
    
    @JsonProperty("Alias")
    public ChatString getAlias() { return Alias; }
    public void setAlias(ChatString value) { Alias = value; }
    
    @JsonProperty("Timestamp")
    public ChatString getTimestamp() { return Timestamp; }
    public void setTimestamp(ChatString value) { Timestamp = value; }
}
