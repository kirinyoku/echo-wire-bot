# News Feed Telegram Bot
Bot for Telegram that gets and posts news to a channel.
## Features
- Fetching articles from RSS feeds
- Article summaries powered by GPT-3.5
- Admin commands for managing sources
## Configuration
### Environment variables
EW_TELEGRAM_BOT_TOKEN — token for Telegram Bot API
EW_TELEGRAM_CHANNEL_ID — ID of the channel to post to, can be obtained via @JsonDumpBot
EW_DATABASE_DSN — PostgreSQL connection string
EW_FETCH_INTERVAL — the interval of checking for new articles, default 10m
EW_NOTIFICATION_INTERVAL — the interval of delivering new articles to Telegram channel, default 1m
EW_FILTER_KEYWORDS — comma separated list of words to skip articles containing these words
EW_OPENAI_KEY — token for OpenAI API
EW_OPENAI_PROMPT — prompt for GPT-3.5 Turbo to generate summary
### HCL
News Feed Bot can be configured with HCL config file. The service is looking for config file in following locations:
- ``./config.hcl``
- ``./config.local.hcl``

The names of parameters are the same except that there is no prefix and names are in lower case instead of upper case.