# Voteban bot

Simple bot to ban group members based on poll results

# Build
```sh
go build
```

# Run
```sh
# After build, run
./unixmerit-voteban

# or just
go run main.go
```

# Configuration
When run, bot expects 2 variables to be set in the current environment:
 - VOTEBAN_LOG_LEVEL -- one of debug/info/warn/error (case insensetive) or any integer according to log/slog package definition of log level
 - VOTEBAN_TG_TOKEN -- telegram bot token obtained from BotFather

It also reads .env file in current directory, if present

# Usage
Reply to message in group with /voteban command to start vote (that will last 1 hour) on whether to ban the author of the original message.
Don't forget to give the bot proper rights in the group (it will remind anyway)
