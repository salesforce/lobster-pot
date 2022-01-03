package gh

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"

	"github.com/salesforce/lobster-pot/config"
	log "github.com/sirupsen/logrus"

	ghinstallation "github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v39/github"
)

// Repo is an authenticated Github client associated with a specific repo
type GithubRepo struct {
	Client *github.Client
	Ctx    context.Context
	Repo   string
	Owner  string
	App    config.GithubApp
}

// NewGithubAuthenticatedClient initializes an authenticated github client
func NewGithubAuthenticatedClient(app config.GithubApp) (*github.Client, error) {

	// Shared transport to reuse TCP connections.
	tr := http.DefaultTransport

	// Wrap the shared transport for use with the app ID authenticating with installation ID
	itr, err := ghinstallation.New(tr, app.ID, app.InstallID, app.PrivateKey)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return github.NewClient(&http.Client{Transport: itr}), nil
}

func DownloadContent(authRepo GithubRepo, path, ref string) ([]byte, error) {
	gh := authRepo.Client
	ow, re := authRepo.Owner, authRepo.Repo

	sha := github.RepositoryContentGetOptions{Ref: ref}
	io, _, err := gh.Repositories.DownloadContents(authRepo.Ctx, ow, re, path, &sha)
	if err != nil {
		log.WithFields(log.Fields{
			"event":  "downloadContent",
			"repo":   re,
			"owner":  ow,
			"commit": sha,
			"error":  err,
		}).Error("Could not download file content")
		return nil, err
	}

	bytes, err := ioutil.ReadAll(io)
	if err != nil {
		log.WithFields(log.Fields{
			"event":  "readContent",
			"repo":   re,
			"owner":  ow,
			"commit": sha,
			"error":  err,
		}).Error("Could not read io content")
		return nil, err
	}
	return bytes, err
}

// ParseReceivedWebhook parses the webhook payload received from Github and
// returns the signature, and the payload. The payload can then be read multiple times.
func ParseReceivedWebHook(r *http.Request) (signature string, payload []byte, err error) {
	signature = r.Header.Get(github.SHA256SignatureHeader)
	if signature == "" {
		signature = r.Header.Get(github.SHA1SignatureHeader)
	}

	if signature == "" {
		msg := "No signature found"
		log.Error(msg)
		return "", nil, errors.New(msg)

	}

	contentType, _, cterr := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if cterr != nil {
		log.Error(cterr)
		return "", nil, cterr
	}

	if contentType != "application/json" {
		msg := fmt.Sprintf("Invalid Content-Type %s", contentType)
		log.Error(msg)
		return "", nil, errors.New(msg)
	}

	payload, perr := ioutil.ReadAll(r.Body)
	if perr != nil {
		log.Error(perr)
		return "", nil, perr
	}

	return signature, payload, nil
}
