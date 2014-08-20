package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/sourcegraph/go-papertrail/papertrail"
)

var (
	systemID   = flag.String("system", "", "system ID (filter)")
	groupID    = flag.String("group", "", "group ID (filter)")
	follow     = flag.Bool("f", false, "follow logs (like 'tail -f')")
	minTimeAgo = flag.Duration("min-time-ago", 0, "show all logs after this time")
	delay      = flag.Duration("delay", 2*time.Second, "poll delay")
)

func main() {
	log.SetFlags(0)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: papertrail-go [options] [query...]\n")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Papertrail log viewer.")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Specify query terms as command-line arguments.")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "The options are:")
		flag.PrintDefaults()
		os.Exit(1)
	}
	flag.Parse()

	token, err := papertrail.ReadToken()
	if err == papertrail.ErrNoTokenFound {
		log.Fatal("No Papertrail API token found; exiting.\n\npapertrail-go requires a valid Papertrail API token (which you can obtain from https://papertrailapp.com/user/edit) to be set in the PAPERTRAIL_TOKEN environment variable or in ~/.papertrail.yml (in the format `token: MYTOKEN`).")
	} else if err != nil {
		log.Fatal(err)
	}

	c := papertrail.NewClient((&papertrail.TokenTransport{Token: token}).Client())

	opt := papertrail.SearchOptions{
		SystemID: *systemID,
		GroupID:  *groupID,
		Query:    strings.Join(flag.Args(), " "),
	}

	if *minTimeAgo != 0 {
		opt.MinTime = time.Now().In(time.UTC).Add(-1 * *minTimeAgo)
	}

	stopWhenEmpty := !*follow && (*minTimeAgo == 0)
	polling := false

	for {
		searchResp, _, err := c.Search(opt)
		if err != nil {
			log.Fatal(err)
		}

		if len(searchResp.Events) == 0 {
			if stopWhenEmpty {
				return
			} else {
				// No more messages are immediately available, so now we'll just
				// poll periodically.
				polling = true
			}
		}
		for _, e := range searchResp.Events {
			var prog string
			if e.Program != nil {
				prog = *e.Program
			}
			fmt.Printf("%s %s %s %s: %s\n", e.ReceivedAt, e.SourceName, e.Facility, prog, e.Message)
		}

		opt.MinID = searchResp.MaxID

		if polling {
			time.Sleep(*delay)
		}
	}
}
