// Package handlers - hook
// Contains the logic to deal with incoming git hooks
package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/salesforce/lobster-pot/config"
	"github.com/salesforce/lobster-pot/db"
	gh "github.com/salesforce/lobster-pot/github"
	"github.com/salesforce/lobster-pot/scanner"
	"github.com/slack-go/slack"

	"github.com/google/go-github/v39/github"
	log "github.com/sirupsen/logrus"
)

//DIFF_MINUTES how many minutes between re-reporting the same issue
const DIFF_MINUTES = 900 // 15 minutes * 60

// GithubWebhookHandler function handles the incoming git hook and passes off
// handling of notifications, logging etc to the relevant functions
func GithubWebhookHandler(w http.ResponseWriter, r *http.Request, c config.Config) {

	defer r.Body.Close()

	log.WithFields(
		log.Fields{
			"Github-Delivery": r.Header.Get("X-Github-Delivery"),
		}).Debug("Received webhook")

	// Extract data from received webhook
	signature, payload, perr := gh.ParseReceivedWebHook(r)
	if perr != nil {
		log.Error(perr)
		w.WriteHeader(http.StatusInternalServerError)
		_, e := w.Write([]byte("Error!"))
		if e != nil {
			log.Error("Could not write to github ", e)
		}
		return
	}

	event, eerr := github.ParseWebHook(github.WebHookType(r), payload)
	if eerr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, e := w.Write([]byte("Error!"))
		if e != nil {
			log.Error("Could not write to github ", e)
		}
		log.Error("could not parse webhook:", eerr)
		return
	}

	var respstatus int
	var respbody []byte

	// hand off to relevant handlers
	switch e := event.(type) {
	case *github.PushEvent:
		// If the payload is a push event, validate it against the proper app secret
		owner := config.GithubOrgName(*e.Repo.Owner.Login)
		log.Debug("Handling push event for ", owner)
		app, ok := c.GithubApps[owner]
		if !ok {
			log.Error("Could not find app for owner ", owner)
			w.WriteHeader(http.StatusInternalServerError)
			_, e := w.Write([]byte("Error!"))
			if e != nil {
				log.Error("Could not write to github ", e)
			}
			return
		}
		if verr := github.ValidateSignature(signature, payload, []byte(app.Secret)); verr != nil {
			w.WriteHeader(http.StatusUnauthorized)
			_, werr := w.Write([]byte("Signature mismatch!"))
			if werr != nil {
				log.Error("Could not write to github ", werr)
			}
			log.Error(verr)
			return
		}
		// If we're here, the payload is valid, so we can continue

		// can't wait for the scan to finish since large scans will timeout
		// so send 200 response
		respstatus = http.StatusOK
		respbody = []byte("received")

		// trigger handler for the event
		go pushEvent(*e, c)

	default:
		respstatus = http.StatusNotFound
		respbody = []byte("unsupported event")
		log.WithFields(log.Fields{"event": event}).Debug("Unsupported event")
	}

	w.WriteHeader(respstatus)
	if respbody != nil {
		_, err := w.Write(respbody)
		if err != nil {
			log.Error(err)
		}
	}
}

func pushEvent(event github.PushEvent, cfg config.Config) (int, []byte, error) {
	log.Debug("********* Start Handling push event *********")
	//https://developer.github.com/v3/activity/events/types/#pushevent
	ref := *event.Ref
	commits := event.Commits
	before := *event.Before
	after := *event.After

	// get repository name and owner
	// could use full_name := heroku/reponame , but the api code expects owner and repo as strings
	// this should already be know, but using the webhook data to avoid
	// hardcoding these values
	repo := *event.Repo.Name
	owner := *event.Repo.Owner.Login
	// who made the push
	pusher := *event.Pusher.Name

	log.WithFields(log.Fields{
		"event":  "pushEvent",
		"ref":    ref,
		"before": before,
		"after":  after,
		"owner":  owner,
		"repo":   repo,
		"pusher": pusher,
	}).Info()

	app, ok := cfg.GithubApps[config.GithubOrgName(owner)]

	log.WithFields(log.Fields{
		"owner":     owner,
		"appID":     app.ID,
		"installID": app.InstallID,
	}).Debug("Github app found")

	if !ok {
		log.WithFields(log.Fields{
			"owner": owner,
		}).Error("Could not find GitHub App for owner")
		return http.StatusInternalServerError, []byte("Internal Server Error"), fmt.Errorf("Could not find GitHub App for owner")
	}

	ghrepo := gh.GithubRepo{
		Repo:  repo,
		Owner: owner,
		Ctx:   context.Background(),
		App:   app,
	}

	// for each commit in push, get files and check if files changed are
	// ones that we are monitoring for changes

	ghclient, err := gh.NewGithubAuthenticatedClient(app)
	if err != nil {
		log.Error("error getting Github authenticated client ", err)
		return http.StatusInternalServerError, []byte("Internal Server Error"), err
	}

	ghrepo.Client = ghclient
	log.Trace(commits)
	for _, c := range commits {
		processCommit(c, ghrepo, cfg)

	}
	log.Debug("********* End Handling push event *********")

	return 200, nil, nil

}

func processCommit(commit *github.HeadCommit, ghrepo gh.GithubRepo, c config.Config) {

	sha := *commit.ID
	repo := ghrepo.Repo
	owner := ghrepo.Owner
	log.WithFields(log.Fields{
		"event":  "processCommit",
		"repo":   repo,
		"owner":  owner,
		"commit": sha,
	}).Info()
	// create temp location for all the files to be downloaded to
	// TODO: make this a configurable location with sane defaults
	tmpFolder, err := ioutil.TempDir("", "lobster-pot")
	if err != nil {
		log.Error(err)
		return
	}

	// count of the total files in the commit. This is different from len(commit.Files) since
	// a commit can contain deleted files, deleted files don't count as a scanned file
	totalFiles := 0

	for _, f := range commit.Removed {
		log.WithFields(log.Fields{
			"event":  "fileRemoved",
			"commit": sha,
			"file":   f,
		}).Info()
	}

	fileToScan := append(commit.Added, commit.Modified...)

	for _, f := range fileToScan {
		totalFiles++

		//TODO: make skipping configurable
		// skip vendor
		if strings.HasPrefix(f, "vendor/") {
			log.WithFields(log.Fields{
				"event":      "skipping",
				"commit":     sha,
				"file":       f,
				"totalFiles": totalFiles,
			}).Debug("Skipping file")
			continue
		}
		// skip node_modules
		if strings.HasPrefix(f, "node_modules/") {
			log.WithFields(log.Fields{
				"event":      "skipping",
				"commit":     sha,
				"file":       f,
				"totalFiles": totalFiles,
			}).Debug("Skipping file")
			continue
		}

		log.WithFields(log.Fields{
			"event":  "DownloadingFile",
			"commit": sha,
			"file":   f,
		}).Info()

		// a retry loop to attempt to retrieve a file 3 times,
		// this tries to account for GitHub sometimes returning a 502 on a new file
		for retry := 0; retry < 3; retry++ {
			content, err := gh.DownloadContent(ghrepo, f, sha)
			if err != nil {
				// If error downloading file, sleep before retrying
				log.Error("Error downloading file ", err)
				time.Sleep(3 * time.Second)
				continue
			}
			err = writeFileOnDisk(tmpFolder, f, content)
			if err != nil {
				log.Error("Error writing file to disk ", err)
				continue
			}
			break
		}

	}

	// scan all the downloaded files
	findingCount, _ := scan(tmpFolder, ghrepo, sha, c)

	// save commit info for metrics
	// commit, repo, number of files scanned, findings
	e := db.InsertCommitScan(sha, fmt.Sprintf("%s/%s", ghrepo.Owner, ghrepo.Repo), totalFiles, findingCount)
	if e != nil {
		log.Error(e)
	}

	// remove all files - make sure to capture the error if files couldn't be removed
	defer func() {
		if err := os.RemoveAll(tmpFolder); err != nil {
			log.Error(err)
		}
	}()
}

func scan(tmpFolder string, ghrepo gh.GithubRepo, sha string, c config.Config) (int, error) {
	log.WithFields(log.Fields{
		"event":  "scan",
		"owner":  ghrepo.Owner,
		"repo":   ghrepo.Repo,
		"commit": sha,
	}).Info()

	// scan
	findings, err := scanner.ScanFolder(tmpFolder, c)
	if err != nil {
		log.Error(err)
		return 0, err
	}

	// track findings that have been reported for a single commit
	// incase multiple Grover rules trigger for a single file+comment
	// we don't want to report the same file multiple times in a single commit
	// reportedFindings := []scanner.Finding{}

	for _, f := range findings {

		fPath := strings.Replace(f.FilePath, tmpFolder, "", 1)

		fid := fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%s/%s.%s.%s", ghrepo.Owner, ghrepo.Repo, fPath, f.Secret))))
		status, updated, e := db.SelectFinding(fid)
		if e != nil { //something went wrong selecting the finding, guess we need to create a new one
			status = -1
			updated = 0
		}

		// check if finding has come up before and if it has
		// is it marked as a False-Positive or "Safe"
		// this is based on the repository, filename and the comment
		// if finding is a repeat and status is VERIFIED_POSITIVE - warn again
		// if finding is new -  warn
		// if finding is FALSE_POSITIVE or KNOWN_SAFE - don't warn

		if status == db.FALSE_POSITIVE || status == db.KNOWN_SAFE {
			log.WithFields(log.Fields{
				"event":    "scanKnownFinding",
				"commit":   sha,
				"filename": fPath,
				"status":   db.FindingValues[status],
			}).Info()
			continue
		}

		// if this is a repeat finding, check when last it was reported
		// if reported within the last 15min, don't report again
		// repeat findings should be reported, but if two or more Grover filters flag the same value
		// only report it once per scan
		if status != -1 {
			// the finding exists in the DB but this is the first time it is seen
			// as a repeat, so we need to ensure it gets marked as a repeat
			if status == db.NEW_FINDING {
				status = db.REPEAT_FINDING
			}
			// update the last seen time
			_, e := db.UpdateFinding(fid, status)
			if e != nil {
				log.Error(e)
			}

			// check if seen in the last 15 min and skip report if reported in last 15 min
			now := int(time.Now().Unix())

			if now-updated < DIFF_MINUTES {
				log.WithFields(log.Fields{
					"event":    "scanKnownFinding",
					"commit":   sha,
					"filename": fPath,
					"status":   db.FindingValues[status],
					"timeDiff": now - updated,
				}).Info("Not posting because too recent")
				continue
			}
		}

		// finding does not exist
		if status == -1 {
			// try insert the finding.
			_, er := db.InsertFinding(sha, fmt.Sprintf("%s/%s", ghrepo.Owner, ghrepo.Repo), fPath, f.Secret)

			if er != nil {
				log.Error(er)
			}
		}

		commitURL := fmt.Sprintf("https://github.com/%s/%s/commit/%s", ghrepo.Owner, ghrepo.Repo, sha)
		// add the line number to the link to make finding the value easier
		fPathURL := fmt.Sprintf("https://github.com/%s/%s/blob/%s%s#L%s", ghrepo.Owner, ghrepo.Repo, sha, fPath, f.LineNumber)

		// Build the message
		// header
		headerSection := createMarkdownBlock("Possible secret detected! :rotating_light:")
		divSection := slack.NewDividerBlock()

		// file path
		fileSection := createMarkdownBlock(fmt.Sprintf("*Repo:* %s/%s\n*Commit:* <%s|%s>\n*Description:* %s\n*FilePath:* <%s|%s#L%s>\n*Scanner:* %s", ghrepo.Owner, ghrepo.Repo, commitURL, sha, f.RuleDescription, fPathURL, fPath, f.LineNumber, f.Scanner))

		// status section (optional)
		// if it is a repeat finding make a note of it
		var statusSection *slack.SectionBlock
		if status == db.VERIFIED_POSITIVE {
			statusSection = createMarkdownBlock(":warning: This finding has been marked as VERIFIED_POSITIVE in the past")
		}
		// if it is a repeat finding make a note of it
		if status == db.REPEAT_FINDING {
			statusSection = createMarkdownBlock(":warning: This is a REPEAT finding and has not been manually verified")
		}

		// test section (optional)
		var testSection *slack.SectionBlock
		// if it is a *spec.rb or *test.go file, label it as such
		// test files tend to contain fake values - still needs to be checked
		if strings.HasSuffix(f.FilePath, "test.go") || strings.HasSuffix(f.FilePath, "spec.rb") || strings.HasSuffix(f.FilePath, "helper.rb") {
			testSection = createMarkdownBlock(":speech_balloon: Likely a spec or test file.")
		}

		// Sample section (optional)
		// check if secret is an example key
		var sampleSection *slack.SectionBlock

		if isSample, sampleType := checkExample(f.Secret); isSample {
			sampleSection = createMarkdownBlock(fmt.Sprintf(":checkmark: Looks like a sample key of type %s", sampleType))
		}

		// Action block with buttons

		vButton := createStyledButton("Verified", "bVerified", fmt.Sprintf("verify_%s", fid), "danger")
		fpButton := createStyledButton("False Positive", "bFP", fmt.Sprintf("fp_%s", fid), "primary")
		ksButton := createStyledButton("Known Safe", "bSafe", fmt.Sprintf("safe_%s", fid), "primary")

		actionBlock := slack.NewActionBlock("", vButton, fpButton, ksButton)

		// Build the top part
		msg := slack.NewBlockMessage(
			headerSection,
			divSection,
			fileSection,
		)

		// add optional sections
		if statusSection != nil {
			msg.Blocks.BlockSet = append(msg.Blocks.BlockSet, statusSection)
		}
		if testSection != nil {
			msg.Blocks.BlockSet = append(msg.Blocks.BlockSet, testSection)
		}
		if sampleSection != nil {
			msg.Blocks.BlockSet = append(msg.Blocks.BlockSet, sampleSection)
		}

		// Build the bottom part
		msg.Blocks.BlockSet = append(msg.Blocks.BlockSet, divSection, actionBlock)

		// queue the message to be sent off via Slack
		// the queued message will also get added to the database once sent
		QueueMessage(fid, msg, ghrepo.App.SlackAppID)

		// reportedFindings = append(reportedFindings, f)

	}
	if len(findings) == 0 {
		log.WithFields(log.Fields{"event": "scanResult", "commit": sha, "result": "Clean_scan"}).Info("Scan result")
	} else {
		log.WithFields(log.Fields{"event": "scanResult", "commit": sha, "result": "Found_secrets", "secrets_found": len(findings)}).Info("Scan result")
	}
	return len(findings), nil
}

type sampleKeys struct {
	Keys []keyPair `json:"keys"`
}
type keyPair struct {
	ID    string `json:"id"`
	Value string `json:"value"`
}

// checkExample compares the string with a list known safe keys
// and checks if it contains the "EXAMPLE" keyword
func checkExample(comment string) (bool, string) {

	var sampleKeys []keyPair
	// load sample keys
	sampleKeys = loadKeys("rules/sample_keys.json")
	// load sample pem keys
	sampleKeys = append(sampleKeys, loadKeys("rules/sample_pem_keys.json")...)

	// check list of keys
	for _, v := range sampleKeys {
		if strings.Contains(comment, v.Value) {
			return true, v.ID
		}
	}

	return false, ""
}

func loadKeys(keysPath string) []keyPair {
	jsonFile, err := os.Open(keysPath)
	if err != nil {
		log.Error(err)
		return []keyPair{}
	}

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	// we initialize our Users array
	var keys sampleKeys

	e := json.Unmarshal(byteValue, &keys)
	if e != nil {
		log.Error(e)
		return []keyPair{}
	}
	return keys.Keys
}
