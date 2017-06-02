/*
 Copyright 2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.

 Licensed under the Apache License, Version 2.0 (the "License"). You may not use this file except in compliance with the License. A copy of the License is located at

     http://aws.amazon.com/apache2.0/

 or in the "license" file accompanying this file. This file is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
*/

const AWS = require('aws-sdk'),
      CONF = require('./config'),
      utils = require('./utils'),
      fs = require('fs');

var s3 = new AWS.S3({region: CONF.region});
var ddbDocClient = new AWS.DynamoDB.DocumentClient({region: CONF.region});

module.exports = {
    getImageData: function(key) {
        var params = { Bucket: CONF.bucket, Key: key };
        return s3.headObject(params).promise();
    },
    getImageKeysByCategory: function(category) {
        var params = {
            TableName: CONF.tableName,
            IndexName: CONF.indexName,
            KeyConditionExpression: "Category = :CatgoryAttribute",
            ExpressionAttributeValues: {
                ":CatgoryAttribute": category
            },
            // we only care about the bucket key, so we'll use
            // ProjectionExpression to return only S3ObjectKey
            ProjectionExpression: "S3ObjectKey"
        };
        return new Promise(function(resolve, reject) {
            var allResults = [];
            ddbDocClient.query(params).eachPage(function(err, data) {
                if (err) {
                    reject(err);
                    return false; // stop pagination
                } else if (data) {
                    allResults = allResults.concat(data.Items);
                } else {
                    resolve(allResults);
                }
            });
        });
    },
    postImage: function(file, category) {
        var params = { Bucket: CONF.bucket,
                       Key: file.filename,
                       Body: fs.createReadStream(file.path) };
        return s3.upload(params).promise()
            .then(function(data) {
                var ddbParams = {
                    Item: {
                        ImageLocation: data.Location,
                        Category: category,
                        S3ObjectKey: params.Key
                   },
                   ReturnConsumedCapacity: "TOTAL",
                   TableName: CONF.tableName
                };
                return ddbDocClient.put(ddbParams).promise();
            });
    },
    removeImage: function(key) {
        var params = { Bucket: CONF.bucket, Key: key };
        return s3.deleteObject(params).promise()
            .then(function(data) {
                var ddbParams = {
                    Key: { S3ObjectKey: params.Key },
                    TableName: CONF.tableName
                };
                return ddbDocClient.delete(ddbParams).promise();
            });
    }
};
