# Apply Intelligence to Gmail Inbox

## Install

1. git clone 
2. make build
3. install the python packages by [README](classifier/README.md) and start the local gRPC service
4. go to google clound, create a project, and create a credential, download the JSON file, save it to the root of the project.
5. create a config.json at the root of the project, in below format

````json
{
    "gmail": {
      "credentials": ".....apps.googleusercontent.com.json",
      "token": "gmail_token.json"
    },
    "rejectionCheck": {
      "url": "localhost:50051"
    }
  }
````

## Run

First start the gRPC server if not already running it.

````sh
cd classifier
python3 grpcserver.py
````

then in a different terminal, start the gmail process

````
bin/gmailai-macos-amd64 --config config.json label-rejection
````

the first time you run the program, the program will print out a link for you to copy to browser to give permission to access your gmail from your google project. After you grant permission, the program will create the access token and save in the "gmail_token.json" file.


## Contribution


### Generate Go code from proto file

````sh
protoc --proto_path=classifier --go_out=integrations --go_opt=paths=source_relative --go_opt=Mclassifier.proto=github.com/jyouturer/gmail-ai/integration --go-grpc_out=./integrations --go-grpc_opt=paths=source_relative --go-grpc_opt=Mclassifier.proto=github.com/jyouturer/gmail-ai/integration classifier.proto 
````

above command will generate two files under the "integrations" folder, with package "integration"