# Slack setup

## Environment vars

- `SLACK_APPID` - The ID of the App, found on the "Basic Information" Page
- `SLACK_CHANNEL` - channel ID to post detected secrets to
- `SLACK_TOKEN` - Slack access token to post
- `SLACK_SIGNING_SECRET`- Slack signing secret to validate incoming requests, found under "App Credentials"

As for the Github apps, all those variables need to be suffixed by a numerical ID, to be able to have multiple orgs :

`SLACK_APPID_1`, `SLACK_CHANNEL_1`, ...

## App installation

The Slack interactivity used by this project needs a Slack app to be setup. This is for both receiving notifications about a detected secret, and for the interactivity to allow marking findings as Valid or false postives.

Create a new app in Slack:

- Go to [https://api.slack.com/apps](https://api.slack.com/apps)
- Click **Create New App**
- Select "From an app manifest"
- Pick the Workspace you want to use
- Paste the manifest, after adapting to your needs:

```yaml
_metadata:
  major_version: 1
  minor_version: 1
display_information:
  name: <Name of the app>
  description: Monitors GitHub for secrets
  background_color: "#000000"
features:
  bot_user:
    display_name: <Name of the bot>
    always_online: false
oauth_config:
  scopes:
    bot:
      - chat:write
      - incoming-webhook
settings:
  interactivity:
    is_enabled: true
    request_url: <URL to the Web app - don't forget to end by /slack >
  org_deploy_enabled: false
  socket_mode_enabled: false
  token_rotation_enabled: false
  ```

- Click **Create**
- Request to add it to your workspace

A Slack admin for the workspace will need to authorize the install. Once installed, you will be able to copy the OAuth token (found on the left panel under Features / Oauth & parmissions) and add this to the app configuration app as the `SLACK_TOKEN_*` config var.
You'll also need the Channel ID to which this app should write to.

- create a new channel
- right click on the channel and "copy link"
- paste the link in an editor and extract the channel id. Example <https://myorg.slack.com/archives/AAABBB111> - the channel ID is **AAABBB111**
- save the Channel ID to the `SLACK_CHANNEL_*` config var
- if it's a private channel, you'll need to add the bot to the channel
