package com.amazonaws.samples;

import com.fasterxml.jackson.annotation.JsonProperty;

public class GetPostsRequest {

    private String sortBy;
    private String sortOrder;
    private Long postsToGet;
    
    @JsonProperty("SortBy")
    public String getSortBy() { return sortBy; }
    public void setSortBy(String value) { sortBy = value; }

    @JsonProperty("SortOrder")
    public String getSortOrder() { return sortOrder; }
    public void setSortOrder(String value) { sortOrder = value; }
    
    @JsonProperty("PostsToGet")
    public Long getPostsToGet() { return postsToGet; }
    public void setPostsToGet(Long value) { postsToGet = value; }
}
