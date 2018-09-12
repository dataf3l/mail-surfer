// mail.go: Show mail stats

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/rollbar/rollbar-go"

	"io/ioutil"
	"sort"
	"strconv"
	"strings"
	"time"
)

// GetMaildirList Configuration: which maildirs are available
func GetMaildirList() []string {
	var maildirs []string
	maildirs = make([]string, 0)

	maildirs = append(maildirs, "/home/teachers2/Maildir/cur")
	maildirs = append(maildirs, "/home/teachers2/Maildir/new")
	maildirs = append(maildirs, "/home/sa/Maildir/cur")
	maildirs = append(maildirs, "/home/sa/Maildir/new")
	if os.Getenv("DEVELOPMENT") == "1" {
		maildirs = make([]string, 0)
		maildirs = append(maildirs, "/Users/b/work/ntutree/tutree_jobs_v2_dev2/tutree_jobs_v2/mail_parser2/MaildirTeachers/Maildir/new")
	}

	return maildirs

}

// SendErrorNotification Notifies us of errors, and also calls Rollbar APIs.
func SendErrorNotification(subject string, body string, to string) {
	rollbar.Warning(body)
	if os.Getenv("DEVELOPMENT") != "1" {
		fmt.Println("Email:" + subject)
	}
}

// ShowMailDir Show a directory
func ShowMailDir(inbox string) string {
	// read dir
	// sumarize
	ax := ""
	files, err := ioutil.ReadDir(inbox)
	if err != nil {
		ax += "FO:" + err.Error()
		return ax
	}

	// process whole folder
	var badFormat uint64
	var folders uint64
	var processed uint64

	badFormat = 0
	folders = 0
	processed = 0
	destinationFolders := make(map[string]int64)
	for _, f := range files {
		processed++
		parts := strings.Split(f.Name(), ".")
		if len(parts) == 1 {
			ax += "FO:bad format:" + f.Name() + "\n"
			badFormat++
			continue
		}
		timestampS := parts[0]
		if timestampS == "" {
			// file is a dot file, ignore file.
			badFormat++
			continue
		}
		timestamp, err := strconv.ParseInt(timestampS, 10, 64)
		if err != nil {
			ax += "FO:not a number, or not a timestamp:" + timestampS + " on file:" + f.Name() + "\n"
			badFormat++
			continue
		}
		tm := time.Unix(timestamp, 0)
		//if time.Since(tm).Hours() > 24*EMAILAGE {
		//	log.Print(tm)
		//	log.Printf("%v %f \n", tm, time.Since(tm).Hours())
		//	tooOld++
		//	continue
		//}

		//log.Println(tm)
		isoDate := tm.Format("2006-01-02")

		destinationFolders[isoDate]++

	}

	names := make([]string, 0, len(destinationFolders))
	for name := range destinationFolders {
		names = append(names, name)
	}
	sort.Strings(names) //sort by key

	ax += "FO:File Organizer V1.0:\n"
	ax += "FO:------- FOLDERS:" + inbox + " ----------\n"
	for name := range names {
		folder := names[name]
		ax += fmt.Sprintf("FO:%s:%d\n", folder, destinationFolders[folder])
	}
	ax += "FO:------- STATS ----------\n"
	ax += fmt.Sprintf("FO:PROCESSED      :%d\n", processed)
	ax += fmt.Sprintf("FO:BAD FORMAT     :%d\n", badFormat)
	ax += fmt.Sprintf("FO:FOLDERS        :%d\n", folders)

	return "<pre>" + ax + "</pre>"
}

// StatsHandler Show stats.
func StatsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	maildirs := GetMaildirList()
	ax := ""
	for m := range maildirs {
		ax += ShowMailDir(maildirs[m])
	}
	fmt.Fprintf(w, "%s", ax)
}

// PingHandler handler for Server monitoring
/* There are several monitoring tools, like
   For instance, statuscake, and uptimerobot.
*/
func PingHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s", "ALL OK")
}

// RollBarSetup Used for server monitoring, If the program fails, it will report the errors here:
//
// https://rollbar.com/dataf4l/ATS_FORM/
//
// Since we have a small amount of rollbar budget (only 5000), then that means
// ideally the program should not send any Info or Warning, only real errors we can
// Do something about.
func RollBarSetup() {
	rollbar.SetToken(os.Getenv("ROLLBAR_TOKEN"))
	if os.Getenv("IS_STAGING") == "1" {
		rollbar.SetEnvironment("staging")
	} else {
		if os.Getenv("IS_DEVELOPMENT") == "1" {
			rollbar.SetEnvironment("development")
		} else {
			rollbar.SetEnvironment("production")
		}
	}
	rollbar.SetCodeVersion("v2") // optional Git hash/branch/tag (required for GitHub integration)
	name, err := os.Hostname()
	if err != nil {
		name = "unknown"
	}
	rollbar.SetServerHost(name)                                     // optional override; defaults to hostname
	rollbar.SetServerRoot("https://github.com/dataf3l/mail-surfer") // path of project (required for GitHub integration and non-project stacktrace collapsing)

}

// main()
// This function starts a new webserver, it uses the net/http package,

func main() {

	RollBarSetup()

	http.HandleFunc("/ping", PingHandler)

	//http.HandleFunc("/ats/candidates", candidate_list_handler)
	http.HandleFunc("/stats", StatsHandler)

	//fs := http.FileServer(http.Dir("public"))
	//http.Handle("/", fs)

	port := os.Getenv("PORT")
	if port == "" {
		port = "7745"
	}
	err := http.ListenAndServe("0.0.0.0:"+port, nil)
	if err != nil {
		rollbar.Critical(err)
		log.Println(err)
	}

	rollbar.Wait()

}
