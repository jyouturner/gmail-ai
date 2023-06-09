# Apply Intelligence to Gmail Inbox

## How It Works

This Go project is intended to run at your local, polling gmail messages, and do something (intelligent) to them.

### Architecture

````mermaid
flowchart LR
    subgraph Go CLI
        A[Read Configuration File]
        B[Poll Gmail API]
        C[Extract Sentences]
        D[Handlers]
    end
    subgraph Handlers
        E[Rejection Check]
        F[Other Handlers]
    end
    subgraph gRPC Service
        G[Receive Email Data]
        H[Check Rejection]
    end
    subgraph Python Program
        I[Train SVM Model]
    end
    A --> B
    B --> C
    C --> D
    D --> E
    D --> F
    E --> G
    G --> H
    I --> H


````

### use case 1 - label "Rejection" emails

Imagine you're searching for a job, and despite your exceptional qualifications, you receive numerous rejection emails in response to your applications. These automated messages hold no value for you, and you don't even want to spend time reading them! Let's employ a machine learning model to identify and eliminate them.

In this project, we utilize natural language processing (NLP) to extract the three most important sentences from the email body. Then, we invoke a local Python gRPC service to determine if the message is a rejection or not. For this task, the Python gRPC service employs a machine learning technique called One-Class SVM.


## Machine Learning with Help from ChatGPT

### One-Class SVM

One-Class SVM is an unsupervised learning algorithm that learns the decision boundary for the normal class ("no_reject"). We can then use this decision boundary to determine if a given email is a rejection email or not. Scikit-learn provides an implementation of One-Class SVM that we can use.

```

Initially, I utilized ChatGPT to identify rejection emails, which worked well. However, privacy 
and security concerns led me to create a locally-running machine learning utility instead. I 
attempted to use PyTorch and trained a model with a CSV file, but the results were not satisfactory, 
possibly due to the training data.

Since machine learning is frequently implemented in Python, we employed gRPC for inter-service 
communication to facilitate this process.
```

### Natural Language Process

We use natural language processing (NLP) to extract the relevant text from the message body. NLP can help you identify and extract the most important information from the message, such as the reason for rejection or the specific language used to convey the rejection. In our project we use the github.com/jdkato/prose module.

```
Because we train the model with simple text like "We appreciate all the time and energy you
invested in your application. Unfortunately, we have decided to move forward with different 
candidates at this time. ", we can not use the full email text body. Gmail API does provide 
Snippet which sometimes is not enough, so I decided to use NLP to extract the important sentences
and send to ML service to check.
```

## Project Structure

  ---
    classifier (the Python ML service)
        Python machine learning code
        proto file
        gRPC service
        http web service
    cmd
        main
    automation (the business logic)
        Config
        gmail handler including the rejection check
    integration
        code to integrate with Gmail, ChatGPT etc
    internal
      logging
      NLP code to extract top sentences from email body

## Install

1. git clone
2. make build
3. install the python packages by [README](classifier/README.md) and start the local gRPC service at port 50051
4. go to google clound, create a project, and create a credential, download the JSON file, save it to the root of the project.
5. create a config.json at the root of the project, in below format

````json
{
    "gmail": {
      "credentials": ".....apps.googleusercontent.com.json",
      "token": "gmail_token.json"
    },
    "grpcService": {
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
bin/gmailai-macos-amd64 --help
bin/gmailai-macos-amd64 --config config.json poll
````

the first time you run the program, it will print out a link for you to copy to browser to give permission to access your gmail from your google project. After you grant permission, the program will create the access token and save in the "gmail_token.json" file.

## Contribution

### Generate Go code from proto file

````sh
protoc --proto_path=classifier --go_out=integration --go_opt=paths=source_relative --go_opt=Mclassifier.proto=github.com/jyouturer/gmail-ai/integration --go-grpc_out=./integration --go-grpc_opt=paths=source_relative --go-grpc_opt=Mclassifier.proto=github.com/jyouturer/gmail-ai/integration classifier.proto 
````

above command will generate two files under the "integration" folder, with package "integration"
