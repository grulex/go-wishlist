# Telegram Wishlist MiniApp Backend

Backend for Telegram Wishlist MiniApp

See more: https://github.com/grulex/telegram-wishlist-miniapp

## Quick start locally
Specify environment variables:
```bash
cp .env.example .env
```
Set you bot token and miniapp url in .env file:
```dotenv
TELEGRAM_BOT_TOKEN=YOU:TOKEN
TELEGRAM_MINI_APP_URL=https://t.me/NameYourBot/miniappName
```
Then you can start backend quickly with docker-compose:
```bash
docker-compose build && docker-compose up
```
Project will be available on http://localhost:8080

## Start image in production
1. Set up postgres database on your host
2. Create base schema in database (run this sql queries: [/sql/0_init.sql](https://github.com/grulex/go-wishlist/blob/main/sql/0_init.sql))
3. Build and start image:
```bash
docker build -t telegram-wishlist-backend:latest .
```
When running the image, specify environment variables for Postgres:
```bash
docker run --rm -p 8080:8080 \
    -e TELEGRAM_BOT_TOKEN='YOUR_BOT_TOKEN' \
    -e TELEGRAM_MINI_APP_URL='https://t.me/NameYourBot/miniappName' \
    -e PG_HOST='YOUR_PG_HOST' \
    -e PG_PORT='YOUR_PG_PORT' \
    -e PG_DATABASE='YOUR_PG_DATABASE' \
    -e PG_USER='YOUR_PG_USER' \
    -e PG_PASSWORD='YOUR_PG_PASSWORD' \
    telegram-wishlist-backend:latest
```
