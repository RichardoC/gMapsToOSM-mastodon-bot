package mastodon

import (
	"context"

	"github.com/mattn/go-mastodon"
	"go.uber.org/zap"
)

// ReplyChecker checks if the bot has already replied to a status
type ReplyChecker struct {
	client *mastodon.Client
	logger *zap.SugaredLogger
}

// NewReplyChecker creates a new reply checker
func NewReplyChecker(client *mastodon.Client, logger *zap.SugaredLogger) *ReplyChecker {
	return &ReplyChecker{
		client: client,
		logger: logger,
	}
}

// HasAlreadyReplied checks if the bot's account has already replied to the given status
func (rc *ReplyChecker) HasAlreadyReplied(ctx context.Context, statusID mastodon.ID, botAccountID mastodon.ID) (bool, error) {
	// Get the status context (ancestors and descendants)
	statusCtx, err := rc.client.GetStatusContext(ctx, statusID)
	if err != nil {
		return false, err
	}

	// Check if any of the descendants (replies) are from the bot
	for _, status := range statusCtx.Descendants {
		if status.Account.ID == botAccountID {
			rc.logger.Debugw("Bot has already replied to this status", "statusID", statusID, "replyID", status.ID)
			return true, nil
		}
	}

	return false, nil
}
