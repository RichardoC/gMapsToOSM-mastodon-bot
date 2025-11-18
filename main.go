package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/mattn/go-mastodon"
	"github.com/thought-machine/go-flags"
	"go.uber.org/automaxprocs/maxprocs"
	"go.uber.org/zap"
)

type Options struct {
	Server   string `long:"server" description:"Mastodon server to connect to" default:"https://mastodon.bot"`
	Verbose          bool   `long:"verbosity" short:"v" description:"Uses zap Development default verbose mode rather than production"`
}

func ConfigureClient(server string) {
	appConfig := &mastodon.AppConfig{
		Server:       server,
		ClientName:   "publicApp",
		Scopes:       "read write follow",
		Website:      "https://github.com/mattn/go-mastodon",
		RedirectURIs: "urn:ietf:wg:oauth:2.0:oob",
	}

	app, err := mastodon.RegisterApp(context.Background(), appConfig)
	if err != nil {
		log.Fatal(err)
	}

	// Have the user manually get the token and send it back to us
	u, err := url.Parse(app.AuthURI)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Open your browser to \n%s\n and copy/paste the given authroization code\n", u)
	var userAuthorizationCode string
	fmt.Print("Paste the code here:")
	fmt.Scanln(&userAuthorizationCode)

	config := &mastodon.Config{
		Server:       "https://mastodon.social",
		ClientID:     app.ClientID,
		ClientSecret: app.ClientSecret,
	}

	// Create the client
	c := mastodon.NewClient(config)

	// Exchange the User authentication code with an access token, that can be used to interact with the api on behalf of the user
	err = c.GetUserAccessToken(context.Background(), userAuthorizationCode, app.RedirectURI)
	if err != nil {
		log.Fatal(err)
	}

	// Lets Export the secrets so we can use them later to preform actions on behalf of the user
	// Without having to request authroization all the time.
	// Exporting this as Environment variables, but it can be a configuration file, or database, anywhere you'd like to keep this credentials
	os.Setenv("MASTODON_CLIENT_ID", c.Config.ClientID)
	os.Setenv("MASTODON_CLIENT_SECRET", c.Config.ClientSecret)
	os.Setenv("MASTODON_ACCESS_TOKEN", c.Config.AccessToken)
}

// Preform user actions wihtout having to re-authenticate again
func doUserActions() {
	// Load Environment variables, config file, secrets from db
	clientID := os.Getenv("MASTODON_CLIENT_ID")
	clientSecret := os.Getenv("MASTODON_CLIENT_SECRET")
	accessToken := os.Getenv("MASTODON_ACCESS_TOKEN")

	config := &mastodon.Config{
		Server:       "https://mastodon.social",
		ClientID:     clientID,
		ClientSecret: clientSecret,
		AccessToken:  accessToken,
	}

	// instanciate the new client
	c := mastodon.NewClient(config)

	// Let's do some actions on behalf of the user!
	acct, err := c.GetAccountCurrentUser(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Account is %v\n", acct)

	finalText := "this is the content of my new post!"
	visibility := "public"

	// Post a toot
	toot := mastodon.Toot{
		Status:     finalText,
		Visibility: visibility,
	}
	post, err := c.PostStatus(context.Background(), &toot)

	if err != nil {
		log.Fatalf("%#v\n", err)
	}

	fmt.Printf("My new post is %v\n", post)

}

func main() {
	// Set and parse command line options
	var opts Options
	parser := flags.NewParser(&opts, flags.Default)
	_, err := parser.Parse()
	if err != nil {
		log.Fatalf("can't parse flags: %v", err)
	}

	var z *zap.Logger

	// Configure the logger
	if opts.Verbose {
		z = zap.Must(zap.NewDevelopment())
	} else {
		z = zap.Must(zap.NewProduction())
	}
	defer z.Sync()
	logger := z.Sugar()

	logger.Infow("Got config", "config", opts)

	// Set maxprocs and have it use our nice logger
	maxprocs.Set(maxprocs.Logger(logger.Infof))
	undo := zap.RedirectStdLog(z)
	defer undo()

	ConfigureClient(opts.Server)
	doUserActions()
}
