package db

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

var db *sql.DB

const (
	NEW_FINDING       = 0
	FALSE_POSITIVE    = 1
	KNOWN_SAFE        = 2
	VERIFIED_POSITIVE = 3
	REPEAT_FINDING    = 4
)

var FindingValues = []string{"NEW_FINDING", "FALSE,POSITIVE", "KNOWN_SAFE", "VERIFIED_POSITIVE", "REPEAT_FINDING"}

func initDB() (err error) {

	createTblStatement := ` CREATE TABLE IF NOT EXISTS scans
    (
        uid serial NOT NULL,
		commit character varying(40) NOT NULL,
		fid character varying(64) NOT NULL,
		repo character varying(100) NOT NULL,
		filepath character varying(255) NOT NULL,
		status int,
		updated int
    )
	WITH (OIDS=FALSE); `

	stmt, err := db.Prepare(createTblStatement)
	if err != nil {
		return err
	}

	_, err = stmt.Exec()

	createTblStatement = ` CREATE TABLE IF NOT EXISTS commits
    (
        uid serial NOT NULL,
        sha character varying(40) NOT NULL,
		repo character varying(100) NOT NULL,
		files int,
		findings int,
		date int
    )
	WITH (OIDS=FALSE); `

	stmt, err = db.Prepare(createTblStatement)
	if err != nil {
		return err
	}

	_, err = stmt.Exec()

	createTblStatement = ` CREATE TABLE IF NOT EXISTS slackMessages
    (
        uid serial NOT NULL,
        fid character varying(64) NOT NULL,
		msgid character varying(255) NOT NULL,
		sentat int
    )
	WITH (OIDS=FALSE); `

	stmt, err = db.Prepare(createTblStatement)
	if err != nil {
		return err
	}

	_, err = stmt.Exec()

	defer stmt.Close()

	return
}

// InsertFinding checks to see if a finding has already been created,
// if not, it adds a new entry to the database.
// If the entry already exists, the
// if it has changed, update the current entry and trigger a new notification
func InsertFinding(sha, repo, filepath, comment string) (status int, err error) {
	if db == nil {
		return -1, fmt.Errorf("database not initialized")
	}

	fid := fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%s.%s.%s", repo, filepath, comment))))

	eStatus := -1

	//check if finding already exists in DB
	e := db.QueryRow("SELECT status FROM scans WHERE fid LIKE $1", fid).Scan(&eStatus)
	if e != nil && e != sql.ErrNoRows {
		return 0, e
	}
	if eStatus != -1 { // finding already exists in the
		if eStatus == NEW_FINDING {
			// this is a repeat of an existing finding update to reflect
			return UpdateFinding(fid, REPEAT_FINDING)
		}
		return eStatus, nil
	}

	//it doesn't exist, insert it
	now := int(time.Now().Unix())

	stmt, err := db.Prepare("INSERT INTO scans(commit,fid,repo,filepath,status,updated) VALUES ($1,$2,$3,$4,$5,$6)")

	if err != nil {
		return -1, err
	}
	_, err = stmt.Exec(sha, fid, repo, filepath, NEW_FINDING, now)
	defer stmt.Close()

	if err != nil {
		return -1, err
	}
	//return that we inserted the value with the NEW_FINDING status
	status = NEW_FINDING

	return
}

// SelectFinding returns the status and updated date of a finding
func SelectFinding(fid string) (state int, updated int, er error) {

	e := db.QueryRow("SELECT status,updated FROM scans WHERE fid LIKE $1", fid).Scan(&state, &updated)
	if e != nil && e != sql.ErrNoRows {
		return 0, -1, e
	}
	if e == sql.ErrNoRows {
		return -1, -1, nil
	}
	return
}

// UpdateFinding sets the new status of an existing finding and updates the date at which it was set
func UpdateFinding(fid string, status int) (state int, er error) {

	stmt, err := db.Prepare("UPDATE scans SET status=$1, updated=$2 WHERE fid LIKE $3")
	if err != nil {
		return -1, err
	}
	now := int(time.Now().Unix())
	_, err = stmt.Exec(status, now, fid)
	if err != nil {
		return -1, err
	}
	defer stmt.Close()

	// updated with new status
	return status, nil
}

// InsertCommitScan inserts information about the scan run for a commit
func InsertCommitScan(commit, repo string, totalFiles, findings int) error {

	stmt, err := db.Prepare("INSERT INTO commits(sha, repo, files, findings, date) VALUES ($1,$2,$3,$4,$5)")
	if err != nil {
		return err
	}
	now := int(time.Now().Unix())
	_, err = stmt.Exec(commit, repo, totalFiles, findings, now)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// updated with new status
	return nil
}

// InsertSlackMessage inserts information about the scan run for a commit
func InsertSlackMessage(fid, msgid string) error {

	// It is pointless to insert a message if msgid is empty, since it's the
	// "primary key" for slackMessages.
	if msgid == "" {
		return fmt.Errorf("msgid is empty, not inserting")
	}

	stmt, err := db.Prepare("INSERT INTO slackMessages(fid,msgid,sentat) VALUES ($1,$2,$3)")
	if err != nil {
		return err
	}
	now := int(time.Now().Unix())
	_, err = stmt.Exec(fid, msgid, now)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// updated with new status
	return nil
}

func GetSlackMessageFid(msgid string) (string, error) {
	fid := ""

	//check if finding already exists in DB
	e := db.QueryRow("SELECT fid FROM slackMessages WHERE msgid LIKE $1", msgid).Scan(&fid)
	if e != nil && e != sql.ErrNoRows {
		return "", e
	}
	return fid, nil
}

func GetSlackMessagesFromFid(fid string) ([]string, error) {

	//check if finding already exists in DB
	rows, err := db.Query("SELECT msgid FROM slackMessages WHERE fid LIKE $1", fid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tfid string
	var fids []string

	for rows.Next() {
		err := rows.Scan(&tfid)
		if err != nil {
			log.Fatal(err)
		}
		fids = append(fids, tfid)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return fids, nil
}

//Connect establishes a connection with the back-end database
func Connect(dburl string) (err error) {
	db, err = sql.Open("postgres", dburl)
	if err != nil {
		log.Error(err)
		return err
	}
	//defer db.Close()

	err = initDB()
	if err != nil {
		log.Error(err)
	}

	return
}
