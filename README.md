# gMapsToOSM-mastodon-bot

Code for <https://c.im/@gMapsToOSM>

Mastodon bot which replies with OpenStreetMap/App links when tagged.

```console
go run . --help
Usage:
  gMapsToOSM-mastodon-bot [OPTIONS]

Application Options:
      --server=        Mastodon server to connect to (default: https://c.im)
      --client-id=     Mastodon application client ID
      --client-secret= Mastodon application client secret
      --access-token=  Mastodon application access token
  -v, --verbosity      Uses zap Development default verbose mode rather than production
      --max-redirects= Maximum number of HTTP redirects to follow (default: 5)
      --poll-interval= How often to poll for new notifications (minimum 60s) (default: 60s)

Help Options:
  -h, --help           Show this help message

2025/12/03 19:47:45 can't parse flags: Usage:
  gMapsToOSM-mastodon-bot [OPTIONS]

Application Options:
      --server=        Mastodon server to connect to (default: https://c.im)
      --client-id=     Mastodon application client ID
      --client-secret= Mastodon application client secret
      --access-token=  Mastodon application access token
  -v, --verbosity      Uses zap Development default verbose mode rather than production
      --max-redirects= Maximum number of HTTP redirects to follow (default: 5)
      --poll-interval= How often to poll for new notifications (minimum 60s) (default: 60s)

Help Options:
  -h, --help           Show this help message
```

Running on a raspberry pi under my desk, so no

## Running locally

### Set up a local mastodon instance

<https://github.com/mastodon/mastodon/blob/main/docs/DEVELOPMENT.md#docker>

### Users

Create two users, and make a status tagging one with a message containing a google maps link, then run gMapsToOSM-mastodon-bot with the relevant credentials

### Running the bot

```
go run . --client-id=FROM_INSTANCE --client-secret=FROM_INSTANCE  --access-token=FROM_INSTANCE --server=http://localhost:8080
```

### Required permissions/scopes

```text
read:notifications 
read:statuses 
profile 
write:notifications 
write:statuses
```