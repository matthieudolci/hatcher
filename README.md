# hatcher	

[![Build Status](https://travis-ci.com/matthieudolci/hatcher.svg?branch=master)](https://travis-ci.com/matthieudolci/hatcher) [![Go Report Card](https://goreportcard.com/badge/github.com/matthieudolci/hatcher)](https://goreportcard.com/report/github.com/matthieudolci/hatcher)

Hatcher is a slack bot written in go. It can:

- Send a survey to users to ask how they are doing, and provides an API to export those results.
- Send and save standup notes 

## Slack App Creation:

- Go to the following url to create a new app: https://api.slack.com/apps
- Retrieve the token https://api.slack.com/apps/{app_id}/install-on-team?

    ```Bot User OAuth Access Token: xoxb-xxxxxxxxx-xxxxxxxxxxxxx```

- Create an environment variable name SLACK_TOKEN with the value of the token you just created:

    ``` export SLACK_TOKEN=xoxb-xxxxxxxxx-xxxxxxxxxxxxx```

## How to start it:

- Start ngrok

    ``` ngork http 9191```

- Copy and past the ngrok url into https://api.slack.com/apps/{app_id}/interactive-messages?

    ``` https://xxxxxx.ngrok.io/slack```

- Start the stack with:

    ```docker-compose up```

## How to use it:

Your users will have to interact a first time with Hatcher by either sending a DM saying `hello` or by saying `@hatcher hello`, `hello @hatcher` in any channels that hatcher belongs to.

It will trigger few questions that need to be answered before the user can use the bot.

You can find out all the bot commands available by sending `help` to Hatcher or in a channel `@hatcher help`.

## How to use the API
Hatcher API documentation is available [here](https://documenter.getpostman.com/view/3454833/RWM9uVgF) 

## Resources:
https://blog.gopheracademy.com/advent-2017/go-slackbot/

https://github.com/sebito91/nhlslackbot

https://github.com/tcnksm/go-slack-interactive

https://api.slack.com/interactive-messages

https://www.calhoun.io/

https://medium.com/aubergine-solutions/how-i-handled-null-possible-values-from-database-rows-in-golang-521fb0ee267

https://flaviocopes.com/golang-tutorial-rest-api/
