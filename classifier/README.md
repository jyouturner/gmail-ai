# API to classify text

## Use Case

* Rejections

Given a text for example email, classify whether it is a rejection of job application.

## Install

````sh
pip install -r requirements.txt 
python -m spacy download en_core_web_sm
````

### gRPC Service

run below command to start the gRPC service at port 

````sh
pyton3 grpcserver.py
````

### Http Service

Run below command will start a Flask web service (port 5000)

````sh
python3 httpserver.py
````

to test

````sh
curl --location 'http://localhost:5000/predict' \
--header 'Content-Type: application/json' \
--data '{
    "text": "Thank you for your time and interest in whatever and our Director, Core Services position. We appreciate your time discussing the role with us, but we have decided to move forward with other individuals who more closely meet our requirements at this time."
}'
````

### Generate Python Proto

````sh
python -m grpc_tools.protoc -I. --python_out=. --grpc_python_out=. classifier.proto
````