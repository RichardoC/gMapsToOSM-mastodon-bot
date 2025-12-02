package main

import (
	"context"
	zlog "log"
	"math/rand"
	"time"

	"github.com/RichardoC/gMapsToOSM-mastodon-bot/pkg/gmaps"
	customMastodon "github.com/RichardoC/gMapsToOSM-mastodon-bot/pkg/mastodon"
	"github.com/RichardoC/gMapsToOSM-mastodon-bot/pkg/ratelimit"
	"github.com/RichardoC/gMapsToOSM-mastodon-bot/pkg/reply"
	"github.com/mattn/go-mastodon"
	"github.com/thought-machine/go-flags"
	"go.uber.org/automaxprocs/maxprocs"
	"go.uber.org/zap"
)

type Options struct {
	Server       string        `long:"server" description:"Mastodon server to connect to" default:"https://mastodon.bot"`
	ClientID     string        `long:"client-id" description:"Mastodon application client ID"`
	ClientSecret string        `long:"client-secret" description:"Mastodon application client secret"`
	AccessToken  string        `long:"access-token" description:"Mastodon application access token"`
	Verbose      bool          `long:"verbosity" short:"v" description:"Uses zap Development default verbose mode rather than production"`
	MaxRedirects int           `long:"max-redirects" description:"Maximum number of HTTP redirects to follow" default:"5"`
	PollInterval time.Duration `long:"poll-interval" description:"How often to poll for new notifications (minimum 60s)" default:"60s"`
}

// Bot represents the main bot instance
type Bot struct {
	client         *mastodon.Client
	replyChecker   *customMastodon.ReplyChecker
	replyGenerator *reply.Generator
	logger         *zap.SugaredLogger
	botAccountID   mastodon.ID
}

// NewBot creates a new bot instance
func NewBot(config *mastodon.Config, httpClient *ratelimit.RateLimitedClient, logger *zap.SugaredLogger) (*Bot, error) {
	client := mastodon.NewClient(config)

	// Verify credentials and get bot account ID
	ctx := context.Background()
	account, err := client.GetAccountCurrentUser(ctx)
	if err != nil {
		return nil, err
	}

	logger.Infow("Bot account verified", "username", account.Username, "id", account.ID)

	// Set up components
	extractor := gmaps.NewExtractor(httpClient, logger)
	replyGen := reply.NewGenerator(extractor, logger)
	replyCheck := customMastodon.NewReplyChecker(client, logger)

	return &Bot{
		client:         client,
		replyChecker:   replyCheck,
		replyGenerator: replyGen,
		logger:         logger,
		botAccountID:   account.ID,
	}, nil
}

// processNotifications fetches and processes all pending mention notifications
func (b *Bot) processNotifications(ctx context.Context) error {
	var pg mastodon.Pagination
	pg.Limit = 20 // Process up to 20 notifications at a time

	for {
		notifs, err := b.client.GetNotifications(ctx, &pg)
		if err != nil {
			return err
		}

		if len(notifs) == 0 {
			b.logger.Debug("No more notifications to process")
			break
		}

		b.logger.Infow("Fetched notifications", "count", len(notifs))

		// Process each notification
		for _, notif := range notifs {
			if notif.Type != "mention" {
				b.logger.Debugw("Skipping non-mention notification", "type", notif.Type, "id", notif.ID)
				continue
			}

			if err := b.processMention(ctx, notif); err != nil {
				b.logger.Errorw("Failed to process mention", "notificationID", notif.ID, "error", err)
				// Continue processing other notifications even if one fails
				continue
			}

			// Dismiss the notification after successful processing
			if err := b.client.DismissNotification(ctx, notif.ID); err != nil {
				b.logger.Warnw("Failed to dismiss notification", "notificationID", notif.ID, "error", err)
				// Not fatal, continue
			} else {
				b.logger.Debugw("Dismissed notification", "notificationID", notif.ID)
			}
		}

		// Check if there are more pages
		if pg.MaxID == "" {
			break
		}
	}

	return nil
}

// processMention handles a single mention notification
func (b *Bot) processMention(ctx context.Context, notif *mastodon.Notification) error {
	status := notif.Status
	if status == nil {
		b.logger.Warnw("Mention notification has no status", "notificationID", notif.ID)
		return nil
	}

	b.logger.Infow("Processing mention", "statusID", status.ID, "from", notif.Account.Username)

	// Check if we've already replied to this status
	alreadyReplied, err := b.replyChecker.HasAlreadyReplied(ctx, status.ID, b.botAccountID)
	if err != nil {
		return err
	}

	if alreadyReplied {
		b.logger.Infow("Already replied to this status, skipping", "statusID", status.ID)
		return nil
	}

	// Generate the reply
	replyText, err := b.replyGenerator.GenerateReply(ctx, status.Content)
	if err != nil {
		return err
	}

	// Post the reply
	toot := &mastodon.Toot{
		Status:      replyText,
		InReplyToID: status.ID,
		Visibility:  status.Visibility, // Match the visibility of the original post
	}

	postedStatus, err := b.client.PostStatus(ctx, toot)
	if err != nil {
		return err
	}

	b.logger.Infow("Posted reply", "statusID", postedStatus.ID, "inReplyTo", status.ID, "text", replyText)

	return nil
}

// Run starts the bot's main polling loop with jitter and exponential backoff
func (b *Bot) Run(ctx context.Context, basePollInterval time.Duration) {
	b.logger.Infow("Starting bot polling loop", "baseInterval", basePollInterval)

	currentInterval := basePollInterval
	consecutiveErrors := 0
	maxBackoff := basePollInterval * 8 // Max 8x the base interval

	// Process notifications immediately on startup
	if err := b.processNotifications(ctx); err != nil {
		b.logger.Errorw("Error processing notifications", "error", err)
		consecutiveErrors++
	}

	// Then continue polling
	for {
		// Add jitter: ±10% of current interval
		jitter := time.Duration(rand.Int63n(int64(currentInterval) / 5)) - currentInterval/10
		nextPoll := currentInterval + jitter

		b.logger.Debugw("Scheduling next poll", "interval", nextPoll, "jitter", jitter)

		select {
		case <-ctx.Done():
			b.logger.Info("Bot shutting down")
			return
		case <-time.After(nextPoll):
			b.logger.Debug("Polling for notifications")
			if err := b.processNotifications(ctx); err != nil {
				b.logger.Errorw("Error processing notifications", "error", err)
				consecutiveErrors++

				// Exponential backoff on errors
				currentInterval *= 2
				if currentInterval > maxBackoff {
					currentInterval = maxBackoff
				}
				b.logger.Warnw("Backing off due to errors", "newInterval", currentInterval, "consecutiveErrors", consecutiveErrors)
			} else {
				// Success - reset to base interval
				if consecutiveErrors > 0 {
					b.logger.Infow("Polling successful, resetting interval", "interval", basePollInterval)
					consecutiveErrors = 0
					currentInterval = basePollInterval
				}
			}
		}
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

	// Validate poll interval
	if opts.PollInterval < 60*time.Second {
		log.Warnw("Poll interval too low, setting to minimum 60s", "requested", opts.PollInterval)
		opts.PollInterval = 60 * time.Second
	}

	// Create rate-limited HTTP client (1 request per second)
	httpClient := ratelimit.NewRateLimitedClient(opts.MaxRedirects, 1.0)

	config := &mastodon.Config{
		Server:       opts.Server,
		ClientID:     opts.ClientID,
		ClientSecret: opts.ClientSecret,
		AccessToken:  opts.AccessToken,
	}

	// Create and start the bot
	bot, err := NewBot(config, httpClient, log)
	if err != nil {
		log.Fatalw("Failed to create bot", "error", err)
	}

	// Run the bot with a cancellable context
	ctx := context.Background()
	bot.Run(ctx, opts.PollInterval)
}
