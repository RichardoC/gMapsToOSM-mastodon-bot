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

func dealWithMention(log *zap.SugaredLogger, client *mastodon.Client) {
	// first check whether the coords are already in the url

	// then try opening the url, and use the url we got redirected to, and see if it has the coords
	// example https://maps.app.goo.gl/PXcjbEmiTyRmxVNdA -> https://www.google.com/maps/place/Mussenden+Temple/data=!4m7!3m6!1s0x48602287e2f8db07:0xa0c2c065afc70175!8m2!3d55.1677806!4d-6.8108972!16zL20vMDZndzVw!19sChIJB9v44ociYEgRdQHHr2XAwqA?coh=277533&entry=tts&g_ep=EgoyMDI1MTExNy4wIPu8ASoASAFQAw%3D%3D&skid=6864092b-d6e1-4c04-8864-4ec8b7c0d841 [useful bit suspected to be 3d55.1677806!4d-6.8108972] -> ~ 55.1677806,-6.8108972
	// https://maps.app.goo.gl/rso1KKh2qK5tMqEm8 -> https://www.google.com/maps/search/54.375880,+-5.551608?entry=tts&g_ep=EgoyMDI1MTExNy4wIPu8ASoASAFQAw%3D%3D&skid=b6f8004e-1d60-49c0-b344-2186a6902830 -> 54.375880, -5.551608

	// if not, fail

}

// Preform user actions wihtout having to re-authenticate again
func doUserActions(log *zap.SugaredLogger, client *mastodon.Client) {

	// Let's do some actions on behalf of the user!
	var pg mastodon.Pagination

	for {
		notifs, err := client.GetNotifications(context.Background(), &pg)
		if err != nil {
			log.Fatal("Failed to get notifications", "error", err)
		}

		for _, n := range notifs {
			log.Info("Got notification", "notification", n, "type", n.Type)
			log.Info("Is the type a mention?", "mention", n.Type == "mention")
		}

		for _, n := range notifs {
			if n.Type != "mention" {
				go func() {
					err := client.DismissNotification(context.TODO(), n.ID)
					if err != nil {
						log.Errorw("failed to dismiss notification", "notificationID", n.ID, "err", err)
					}
					return
				}()
			} else {
				go dealWithMention(log, client)
			}

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

	mastodonConnection := mastodon.NewClient(config)

	_, err = mastodonConnection.VerifyAppCredentials(context.Background())
	if err != nil {
		log.Fatal("Failed to verify credentials with the server", err, err)
	}

	// check credentials work
	// do the actual thing, try not to do too many things at once, and have backoffs if we get 429s
	// polling loop
	doUserActions(log, mastodonConnection)
}
