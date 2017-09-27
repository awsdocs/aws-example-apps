# ChatApp Client: Java #
This lightweight commandline application demonstrates how to interact with Amazon Web Services using the [AWS SDK for Java](https://aws.amazon.com/sdk-for-java/).

## Usage ##

### Build ###    

mvn package

### Configure ###

    export JAVA_SDK_HOME=<path to sdk installation>
    export CLASSPATH=<path to myJar.jar file>:$JAVA_SDK_HOME/lib/*:$JAVA_SDK_HOME/third-party/lib/*:<path to json-simple.jar>

### Run ###

    java com.amazonaws.samples.ChatClient 

## TODO ##
* Add UI 
