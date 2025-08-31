# Uaru Votes Bot

Simple bot to ban group members based on poll results

## Build
```sh
go build
```

## Run
```sh
# After build, run
./votes

# or just
go run main.go
```

## Configuration
When run, bot expects these variables to be set in the current environment:
 - `VOTEBAN_LOG_LEVEL` -- one of debug/info/warn/error (case insensitive) or any integer according to log/slog package definition of log level
 - `VOTEBAN_TG_TOKEN` -- telegram bot token obtained from BotFather
 - `VOTEBAN_POLL_DURATION_SECONDS` -- poll duration in seconds (optional, defaults to 3600 seconds = 1 hour, min 30s, max 24h)
 - `ADMINS_ONLY` -- if set to "true", only administrators can use bot commands (useful for testing)
 - `TARGET_USER_ID` -- user ID for message filtering (optional)
 - `DELETION_PROBABILITY` -- probability of message deletion for target user (0.0 to 1.0, optional)

It also reads .env file in current directory, if present

## Commands
- `/voteban`, `/vote`, `/ban` - Start vote to ban user (Yes/No)
- `/voteunban`, `/unban` - Start vote to unban user (Yes/No)
- `/votegif`, `/gif` - Start vote to restrict gifs/stickers (Restrict/Allow)
- `/votemedia`, `/media` - Start vote to restrict media (Restrict/Allow)

