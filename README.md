# hatcher	

[![Build Status](https://travis-ci.com/matthieudolci/hatcher.svg?branch=master)](https://travis-ci.com/matthieudolci/hatcher)

Hatcher is a slack bot written in go. It can:

- Send a survey to users to ask how they are doing, and provides an API to export those results.
- (WIP) Send and save standup notes 

## Slack App Creation:

- Go to the following url to create a new app: https://api.slack.com/apps
- Retrieve the token https://api.slack.com/apps/{app_id}/install-on-team?

    ```Bot User OAuth Access Token: xoxb-xxxxxxxxx-xxxxxxxxxxxxx```

- Create an environment variable name SLACK_TOKEN with the value of the token you just created:

    ``` export SLACK_TOKEN=xoxb-xxxxxxxxx-xxxxxxxxxxxxx```

## How to use it:

- Start ngrok

    ``` ngork http 9191```

- Copy and past the ngrok url into https://api.slack.com/apps/{app_id}/interactive-messages?

    ``` https://xxxxxx.ngrok.io/slack```

- Start the stack with:

    ```docker-compose up```

## Resources:
https://blog.gopheracademy.com/advent-2017/go-slackbot/

https://github.com/sebito91/nhlslackbot

https://github.com/tcnksm/go-slack-interactive

https://api.slack.com/interactive-messages

https://www.calhoun.io/

https://medium.com/aubergine-solutions/how-i-handled-null-possible-values-from-database-rows-in-golang-521fb0ee267

https://flaviocopes.com/golang-tutorial-rest-api/
