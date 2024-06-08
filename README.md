# Openrouter-GPT-Telegram-Bot
A Telegram bot that integrates with Openrouter APIs to provide answers.

## Installing
Clone the repository and navigate to the project directory:

`
git clone https://github.com/defriend/openrouter-gpt-telegram-bot.git
`

`
cd openrouter-gpt-telegram-bot
` 

Rename example.env to .env and setup API keys and Telegram API token.
API Keys works from [Openrouter API](https://openrouter.ai/keys) and [OpenAI API](https://platform.openai.com/api-keys).

Telegram API Token you can find in [@BotFather](https://t.me/BotFather).

**If you have go installed:**

`go run main.go
`
### Using Docker Compose
Run the following command to build and run the Docker image:

`docker compose up`

Thanks to:
- https://github.com/n3d1117/chatgpt-telegram-bot
- https://openrouter.ai Api Docs
