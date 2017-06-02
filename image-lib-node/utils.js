/*
 Copyright 2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.

 Licensed under the Apache License, Version 2.0 (the "License"). You may not use this file except in compliance with the License. A copy of the License is located at

     http://aws.amazon.com/apache2.0/

 or in the "license" file accompanying this file. This file is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
*/

function buildFileName(file, ext) {
    var fname = file.originalname.replace(/\.[^/.]+$/, ""); // remove extension
    return fname + "_"  + Date.now() + ext;
}

module.exports = {
    getImageFileName: function(file) {
        switch(file.mimetype) {
            case "image/jpeg":
                return buildFileName(file, ".jpg");
            case "image/png":
                return buildFileName(file, ".png");
            case "image/gif":
                return buildFileName(file, ".gif");
            default:
                return buildFileName(file, "");
        }
    }
};
