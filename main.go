package main

import (
	"context"
	// "fmt"
	zlog "log"

	"github.com/mattn/go-mastodon"
	"github.com/thought-machine/go-flags"
	"go.uber.org/automaxprocs/maxprocs"
	"go.uber.org/zap"
)

type Options struct {
	Server       string `long:"server" description:"Mastodon server to connect to" default:"https://mastodon.bot"`
	ClientID     string `long:"client-id" description:"Mastodon application client ID"`
	ClientSecret string `long:"client-secret" description:"Mastodon application client secret"`
	AccessToken  string `long:"access-token" description:"Mastodon application access token"`
	Verbose      bool   `long:"verbosity" short:"v" description:"Uses zap Development default verbose mode rather than production"`
}

// Preform user actions wihtout having to re-authenticate again
func doUserActions(log *zap.SugaredLogger, config *mastodon.Config) {

	// instanciate the new client
	c := mastodon.NewClient(config)

	// Let's do some actions on behalf of the user!
	var pg mastodon.Pagination
	notifs, err := c.GetNotifications(context.Background(), &pg)
	if err != nil {
		log.Fatal("Failed to get notifications", "error", err)
	}
	for _, n := range notifs {
		log.Info("Got notification", "notification", n, "type", n.Type)
		log.Info("Is the type a mention?", "mention", n.Type == "mention")
	}
	// filter only for mentions
	// do the replies, with a pool of goroutines to ensure we don't do too many?
	// dismiss the notification once we do the reply
	// continue paging through
	// fmt.Printf("notifications are %+v\n", notifs)

	// finalText := "this is the content of my new post!"
	// visibility := "public"

	// // Post a toot
	// toot := mastodon.Toot{
	// 	Status:     finalText,
	// 	Visibility: visibility,
	// }
	// post, err := c.PostStatus(context.Background(), &toot)

	// if err != nil {
	// 	log.Fatalf("%#v\n", err)
	// }

	// fmt.Printf("My new post is %v\n", post)

}

func main() {
	// Set and parse command line options
	var opts Options
	parser := flags.NewParser(&opts, flags.Default)
	_, err := parser.Parse()
	if err != nil {
		zlog.Fatalf("can't parse flags: %v", err)
	}

	var z *zap.Logger

	// Configure the logger
	if opts.Verbose {
		z = zap.Must(zap.NewDevelopment())
	} else {
		z = zap.Must(zap.NewProduction())
	}
	defer z.Sync()
	log := z.Sugar()

	// Set maxprocs and have it use our nice logger
	maxprocs.Set(maxprocs.Logger(log.Infof))
	undo := zap.RedirectStdLog(z)
	defer undo()

	config := &mastodon.Config{
		Server:       opts.Server,
		ClientID:     opts.ClientID,
		ClientSecret: opts.ClientSecret,
		AccessToken:  opts.AccessToken,
	}
	// check credentials work
	// do the actual thing, try not to do too many things at once, and have backoffs if we get 429s
	// polling loop
	doUserActions(log, config)
}
