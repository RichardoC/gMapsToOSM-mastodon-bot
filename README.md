# gMapsToOSM-mastodon-bot
Code for https://mastodon.bot/@gMapsToOSM




## Running locally

### Set up a local mastodon instance
<https://github.com/mastodon/mastodon/blob/main/docs/DEVELOPMENT.md#docker>

curl 'http://localhost:3000/api/v1/statuses' \
  -X POST \
  --data-raw '{"status":"boop","spoiler_text":"","in_reply_to_id":"115573172928989452","media_ids":[],"sensitive":false,"visibility":"public","poll":null,"language":"en","quoted_status_id":null,"quote_approval_policy":"public"}'