package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/sourcegraph/go-papertrail/papertrail"
)

var (
	systemID = flag.String("system", "", "system ID (filter)")
	groupID  = flag.String("group", "", "group ID (filter)")
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

	opt := &papertrail.SearchOptions{
		Query: strings.Join(flag.Args(), " "),
	}
	searchResp, _, err := c.Search(opt)
	if err != nil {
		log.Fatal(err)
	}

	if len(searchResp.Events) == 0 {
		log.Println("No matching log events found.")
		return
	}
	for _, e := range searchResp.Events {
		var prog string
		if e.Program != nil {
			prog = *e.Program
		}
		fmt.Printf("%s %s %s %s: %s\n", e.ReceivedAt, e.SourceName, e.Facility, prog, e.Message)
	}
}
