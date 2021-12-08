package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/heroku/lobster-pot/config"
	"github.com/heroku/lobster-pot/db"
	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
)

type Job struct {
	FID     string
	Msg     slack.Message
	Retries int
	appID   config.SlackAppID
}

// QueueMessage adds a message to the slack message queue
func QueueMessage(fid string, message slack.Message, slackAppID config.SlackAppID) {
	// need to check if space in queue before inserting
	// this will block until there is space in the queue
	jb := &Job{fid, message, 0, slackAppID}
	messageQueue <- jb
}

var messageQueue = make(chan *Job, 200) // space for queueing 200 slack messages
var rateLimit = 200 * time.Millisecond  // basic rate limit to tick every 200 milliseconds

// StartQueueWorker starts the Slack Queue Worker to monitor a queue for new messages to post to slack
// the worker ensures that messages are rate limited to avoid spamming the channel
func StartQueueWorker(c config.Config) {

	log.Debug("Starting Slack Queue Worker")
	//nolint:staticcheck // this is an endless loop
	limiter := time.Tick(rateLimit)
	go func() {
		for {
			<-limiter            // wait until the limiter ticks
			jb := <-messageQueue // get message from the Queue
			// try send message
			log.WithFields(log.Fields{"Job Message": jb.Msg}).Debug("Posting to slack")
			if messageTs, er := PostToSlack(jb.Msg, jb.appID, c); er != nil {
				log.WithFields(
					log.Fields{
						"message id": messageTs,
						"error":      er,
					}).Debug("Error posting Slack message")
				if rateLimitedError, ok := er.(*slack.RateLimitedError); ok {
					// rate limited
					// update the rate limiter
					rl := rateLimitedError.RetryAfter
					//nolint:staticcheck // this is an endless loop
					limiter = time.Tick(time.Duration(rl) * time.Second)
					// insert the message back into the queue
					messageQueue <- jb
					log.WithFields(log.Fields{"retry-after": rl}).Error("Rate limited")

				} else { // some unhandled error, add message back to queue so we can try again
					log.Error("Error posting to slack.", er)
					if jb.Retries < 3 {
						jb.Retries++
						log.WithFields(log.Fields{"retries": jb.Retries}).Error("Adding message back to Queue")
						// Update limiter with linear backoff
						// TODO: update to a better backoff mechanism
						//nolint:staticcheck // this is an endless loop
						limiter = time.Tick(time.Duration(jb.Retries) * 10 * time.Second)
						messageQueue <- jb
					} else {
						log.WithFields(log.Fields{"job fid": jb.FID}).Error("Message failed to send 3 times, dropping from queue")
					}
				}
			} else {
				// save message to database
				// save MessageTS to the database, allowing for future updating
				log.WithFields(log.Fields{"Job ID": jb.FID, "messageTs": messageTs}).Info("Inserting into DB")
				err := db.InsertSlackMessage(jb.FID, messageTs)
				if err != nil {
					log.Error(err)
				}
				// reset rate limiter
				//nolint:staticcheck // this is an endless loop
				limiter = time.Tick(rateLimit)

			}
		}
	}()
}

func slackAPI(app config.SlackApp) *slack.Client {
	var options []slack.Option
	if log.IsLevelEnabled(log.TraceLevel) {
		options = append(options, slack.OptionDebug(true))
	}
	return slack.New(app.Token, options...)
}

func PostToSlack(message slack.Message, appID config.SlackAppID, c config.Config) (messageTs string, err error) {
	log.WithFields(log.Fields{
		"message": message,
		"appID":   appID,
	}).Debug("Posting to slack")
	slackApp, ok := c.SlackApps[appID]
	if !ok {
		log.Error("Slack app not found with id %s", appID)
		return "", fmt.Errorf("No slack app found for ID %s. Check your config", appID)
	}

	channel := slackApp.Channel
	api := slackAPI(slackApp)

	_, ts, err := api.PostMessage(
		channel,
		slack.MsgOptionBlocks(message.Blocks.BlockSet...),
		slack.MsgOptionText(message.Text, false),
		slack.MsgOptionAttachments(message.Attachments...),
		slack.MsgOptionAsUser(true), // Add this if you want that the bot would post message as a user, otherwise it will send response using the default slackbot
	)

	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Error posting to slack")
		return "", err
	}

	log.WithFields(log.Fields{"messageTs": ts}).Debug("Slack message successfully posted")

	return ts, nil
}

func SlackCallback(w http.ResponseWriter, r *http.Request, c config.Config) {

	// Verify that the request is coming from Slack
	log.Info("Slack callback received")
	if r.Method != "POST" {
		w.WriteHeader(http.StatusNotFound)
		_, err := w.Write([]byte("unsupported event"))
		if err != nil {
			log.Error(err)
		}
		return
	}

	//TODO: Validate Signature

	var payload slack.InteractionCallback

	err := json.Unmarshal([]byte(r.FormValue("payload")), &payload)

	if err != nil {
		log.Error("Could not parse action response JSON: ", err)
		return
	}

	if log.IsLevelEnabled(log.DebugLevel) {
		p, _ := json.Marshal(payload)
		log.WithFields(log.Fields{"payload": string(p)}).Debug("Slack callback received")
	}

	appID := config.SlackAppID(payload.APIAppID)

	actionID := payload.ActionCallback.BlockActions[0].ActionID
	userID := payload.User.ID
	originalMessage := payload.Message.Msg
	messageTS := originalMessage.Timestamp
	responseURL := payload.ResponseURL

	var verifiedAs string
	var er error

	// actionID will be: verify_fid, fp_fid, safe_fid
	// split to get the status to set and the fid of the finding to change
	status := strings.Split(actionID, "_")

	log.WithFields(log.Fields{
		"actionID":    actionID,
		"userID":      userID,
		"messageTS":   messageTS,
		"status":      status,
		"responseURL": responseURL,
	}).Debug("Slack callback details")

	switch status[0] {
	case "verify":
		_, er = db.UpdateFinding(status[1], db.VERIFIED_POSITIVE)
		verifiedAs = fmt.Sprintf(":fire: Verified by <@%s> as POSITIVE", userID)
	case "fp":
		_, er = db.UpdateFinding(status[1], db.FALSE_POSITIVE)
		verifiedAs = fmt.Sprintf(":checkmark: Verified by <@%s> as FALSE_POSITIVE", userID)
	case "safe":
		_, er = db.UpdateFinding(status[1], db.KNOWN_SAFE)
		verifiedAs = fmt.Sprintf(":checkmark: Verified by <@%s> as KNOWN_SAFE", userID)
	default:
		log.Error("Unknown action state")
	}

	// respond to slack that the message has been parsed
	// it doesn't matter that we respond with 200 even if something broke,
	// slack just expects a 200 response
	w.WriteHeader(http.StatusOK)
	_, e := w.Write([]byte("OK"))
	if e != nil {
		log.Error(e)
	}

	// if an error had occurred during the update of the status, report error and
	// don't update the slack message
	if er != nil {
		log.Error(er)
		return
	}

	// rebuild the original message, we want all the same info except the actions section that contained the buttons
	originalBlocks := originalMessage.Blocks.BlockSet
	newBlocks := []slack.Block{}

	// keep all blocks except the actions section
	for _, block := range originalBlocks {
		if block.BlockType() != slack.MBTAction {
			log.WithFields(log.Fields{"block": block}).Debug("appending Block")
			newBlocks = append(newBlocks, block)
		}
	}

	if log.IsLevelEnabled(log.TraceLevel) {
		nb, _ := json.Marshal(newBlocks)
		log.WithFields(log.Fields{"new blocks": string(nb)}).Trace("filtering new blocks")
	}

	verifiedAsBlock := createMarkdownBlock(verifiedAs)
	newBlocks = append(newBlocks, verifiedAsBlock)

	msg := slack.NewBlockMessage(newBlocks...)

	if log.IsLevelEnabled(log.TraceLevel) {
		m, _ := json.Marshal(msg)
		log.WithFields(log.Fields{"message": string(m)}).Trace("Creating message with blocks")
	}
	msg.Text = verifiedAs
	if log.IsLevelEnabled(log.DebugLevel) {
		m, _ := json.Marshal(msg)
		log.WithFields(log.Fields{"message": string(m)}).Debug("Adding text to message")
	}

	// update the message in slack
	er = UpdateSlack(msg, messageTS, responseURL, appID, c)
	if er != nil {
		log.Error(e)
	}
	// update all other findings with the same fid as this one
	UpdateSlackMessages(msg, messageTS, appID, c)

}

func UpdateSlack(message slack.Message, ts string, respURL string, appID config.SlackAppID, c config.Config) (err error) {

	slackApp := c.SlackApps[appID]
	channel := slackApp.Channel
	api := slackAPI(slackApp)

	// Building options
	optionText := slack.MsgOptionText(message.Text, false)
	optionsBlocks := slack.MsgOptionBlocks(message.Blocks.BlockSet...)

	options := []slack.MsgOption{optionText, optionsBlocks}
	if respURL != "" {
		options = append(options, slack.MsgOptionReplaceOriginal(respURL))
	}

	if log.IsLevelEnabled(log.DebugLevel) {
		m, _ := json.Marshal(message)
		o, _ := json.Marshal(options)
		log.WithFields(log.Fields{"message": string(m), "options": string(o), "ts": ts}).Debug("Updating slack message")
	}
	_, _, _, err = api.UpdateMessage(
		channel,
		ts,
		options...,
	)

	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Error posting to slack")
		return err
	}

	return nil
}

func UpdateSlackMessages(message slack.Message, messageTS string, appID config.SlackAppID, c config.Config) {
	log.WithFields(log.Fields{"message ts": messageTS}).Info("Update findings for message")
	// get the message fid from the database
	fid, err := db.GetSlackMessageFid(messageTS)
	log.WithFields(log.Fields{"fid": fid}).Info("Slack message")
	if err != nil {
		log.Error(err)
		return
	}
	// get all the messages that have been sent for that fid
	fids, err := db.GetSlackMessagesFromFid(fid)
	if err != nil {
		log.Error(err)
		return
	}

	// update all the slack messages
	for _, v := range fids {
		log.WithFields(log.Fields{"message id": v}).Debug("Update slack message")
		err := UpdateSlack(message, v, "", appID, c)
		if err != nil {
			log.WithFields(log.Fields{"message id": v}).Error(err)
		}
	}
}

func createMarkdownBlock(text string) *slack.SectionBlock {
	t := slack.NewTextBlockObject("mrkdwn", text, false, false)
	return slack.NewSectionBlock(t, nil, nil)
}

func createStyledButton(text string, value string, actionid string, style string) *slack.ButtonBlockElement {
	t := slack.NewTextBlockObject("plain_text", text, false, false)
	be := slack.NewButtonBlockElement(actionid, value, t)
	return be.WithStyle(slack.Style(style))
}
