/*
 Copyright 2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.

 Licensed under the Apache License, Version 2.0 (the "License"). You may not use this file except in compliance with the License. A copy of the License is located at

     http://aws.amazon.com/apache2.0/

 or in the "license" file accompanying this file. This file is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
*/

const express = require('express'),
      router = express.Router(),
      aws = require('./awsUtils'),
      multer = require('multer'),
      utils = require('./utils');

const storage = multer.diskStorage({
    // no destination specified, so use default OS temp dir
    filename: function (req, file, callback) {
        callback(null, utils.getImageFileName(file));
    }
});

const upload = multer({ storage: storage });

// example: GET http://<host>:3000/images?category=lolcats
router.get('/', function (req, res) {
    var result = aws.getImageKeysByCategory(req.query.category);
    result.then(function(data) {
        res.json(data);
    }).catch(function(err) {
        res.status(500).json(err);
    });
})

// example: GET http://<host>:3000/images/lolcat_1491233250126.jpg
router.get('/:objectKey', function (req, res) {
    var result = aws.getImageData(req.params.objectKey);
    result.then(function(data) {
        res.json(data);
    }).catch(function(err) {
        res.status(404).json(err);
    });
})
// example: POST http://<host>:3000/images?category=lolcats
//          You can also include an image file in the body of the request;
//          use multipart/form-data encoding.
router.post('/', upload.single("photo"), function (req, res) {
    var result = aws.postImage(req.file, req.query.category);
    result.then(function(data) {
        res.json(data);
    }).catch(function(err) {
        res.status(400).json(err);
    });
})

// example: DELETE http://<host>:3000/images/lolcat_1491233250126.jpg
router.delete('/:objectKey', function (req, res) {
    var result = aws.removeImage(req.params.objectKey);
    result.then(function(data) {
        res.json(data);
    }).catch(function(err) {
        res.status(500).json(err);
    });
})

module.exports = router;
