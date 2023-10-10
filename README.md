# Telegram Wishlist MiniApp Backend

Backend for Telegram Wishlist MiniApp

See more: https://github.com/grulex/telegram-wishlist-miniapp

## Quick start
You can start backend quickly with docker:
```bash
docker build -t telegram-wishlist-backend .
```
```bash
docker run --rm -p 8080:8080 \
-e TELEGRAM_BOT_TOKEN='YOUR_BOT_TOKEN' \
-e TELEGRAM_MINI_APP_URL='http://t.me/NameYourBot/miniappName' \
telegram-wishlist-backend:latest
```
Project will be available on http://localhost:8080

All data will be stored in memory, if you stop container, all data will be lost.

If you want store data permanently, you can use postgres database (for example, with [this image](https://hub.docker.com/_/postgres)).

When running the image, specify environment variables for Postgres:
```bash
docker run --rm -p 8080:8080 \
-e TELEGRAM_BOT_TOKEN='YOUR_BOT_TOKEN' \
-e TELEGRAM_MINI_APP_URL='http://t.me/NameYourBot/miniappName' \
-e PG_HOST='YOUR_PG_HOST' \
-e PG_PORT='YOUR_PG_PORT' \
-e PG_DATABASE='YOUR_PG_DATABASE' \
-e PG_USER='YOUR_PG_USER' \
-e PG_PASSWORD='YOUR_PG_PASSWORD' \
telegram-wishlist-backend:latest
```
