# Gmail AI Toolbox

Use ChatGTP API to apply intelligence to Gmails inbox

## Handle Job Application Rejections

Use case: when you are applying jobs, chances are you will receive many automatic rejection emails. Those emails have no values, and they often make you feel worse.

We can use Gmail API to read emails, and send to ChatAPI to decide whether it is a rejection or not. If so, we can label the email, or even archive it.

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

````sh
bin/gmailai-macos-amd64 --config config.json label-rejection
````
the first time you run the program, the program will print out a link for you to copy to browser to give permission to access your gmail from your google project. After you grant permission, the program will create the access token and save in the "gmail_token.json" file.
