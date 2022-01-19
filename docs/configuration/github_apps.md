# GitHub app Setup

## Creation

A github app must be installed in each org that is going to be monitored.  

To create one, see [this doc](https://docs.github.com/en/developers/apps/building-github-apps/creating-a-github-app).  
No `Callback URL` is needed, as we're only going to receive push events from Github.

## Configuration

In order to be able to run in multiple github orgs, the following variables must be suffixed by a numerical ID :

Each Github app needs to have ENV variables set :

* `GITHUB_ORG` - The name of the github org, as it appears in the URL (ex: <https://github.com/heroku> would be `heroku`)
* `GITHUB_APPID` - ID of the installed Github App
* `GITHUB_INSTALLID` - InstallID of the GitHub App, can be extracted from the URL when accessing the app's configuration
* `GITHUB_PRIVATE_KEY` - Created while creating the app
* `GITHUB_SECRET` - secret required from the GitHub App, to validate incoming payloads (can be ommited in `dev` enviromnent)
* `GITHUB_SLACK_APPID` - The ID of the Slack App to post messages to.

All those variables need to be suffixed by a numerical ID, to be able to have multiple orgs :  
`GITHUB_ORG_1`, `GITHUB_APPID_1`, ... 

The only hard requirement is that numerical IDs are only digits. They don't necessarily have to be in sequence. One can have `GITHUB_ORG_1`, `GITHUB_ORG_1337`, `GITHUB_ORG_42`... 

## App installation

The Webhook URL is set to this web application's URL.  
The app is configured to receive all Push events from the org:
![Github event config](../medias/gh-events.png)
