# Apply Intelligence to Gmail Inbox

## Install

1. git clone 
2. make build
3. get ChatGPT API key
4. go to google clound, create a project, and create a credential, download the JSON file, save it to the root of the project.
5. create a config.json at the root of the project, in below format

````json
{
    "gmail": {
      "credentials": ".....apps.googleusercontent.com.json",
      "token": "gmail_token.json"
    },
    "chatgpt": {
      "api_key": "..."
    }
  }
````

## Run

Example

````sh
bin/gmailai-macos-amd64 --config config.json label-rejection
````

the first time you run the program, the program will print out a link for you to copy to browser to give permission to access your gmail from your google project. After you grant permission, the program will create the access token and save in the "gmail_token.json" file.
