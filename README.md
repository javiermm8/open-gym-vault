# OpenGymVault
A simple but scalable and extremely customizable gym tracker.

> [!WARNING]
> OpenGymVault is in active development, things might break. They shouldn't but they might, if they do please create an issue and I'll try to fix them as soon as possible. Do not rely on this project to store any kind of valuable information.  

## What does it do?
At its core its just a REST API to modify a database. You save your gym sessions so you can retrieve them later and get stats about them to enhance your gym experience. There's also a client_s field(which stands for client stuff, I know, very sophisticated) which stores whatever json blob you provide. This can be used by clients to expand open-gym-vault's functionality. For example, someone might want to make a gameified client, so they could store ranks, pets or whatever they want in that field and retrieve it later.

> [!NOTE]
> OpenGymVault is on a very early stage of development, the stats engine is closer to a placeholder than an actual feature(I wanted to calculate personal records so I added it in quickly) and client_s isn't used to its full potential. The idea for the future is to modularize it more so forking and adapting OpenGymVault to your necessities is trivial or altogether unnecessary. 

## Why does this exist?
I didn't want to pay $90/year for one of the millions of saas apps on the App Store. They're essentially just gym-themed excel spreadsheets, so I'm making my own. It's also the perfect opportunity for me to take a shot at building something actually production ready, it's simple and it's something I would use almost every day.

## Features
- Full OpenAPI spec at /openapi.json and interactive redocly docs at /openapi
- Persistence through a Postgres database
- Redis based rate limiter, supports multiple replicas
- Authorization with Bearer token or cookie supported

## What did I build it with?
- Languages: Go for basically everything, SQL for migrations and queries(that sqlc translates into go code) and YAML for the openapi spec(that oapi-codegen translates into go code)
- Tools: sqlc(takes my handwritten SQL queries and generates a Go code layer so I never have to touch SQL inside Go), golang-migrate(applies the migrations), oapi-codegen(does the same but with the openapi.yml specification. It forces my code to follow it, if it doesn't, the program doesn't compile, usually), docker compose and makefile.

## How to set it and run it?
With docker compose:
1. Create a directory and cd into it
2. Create a .env file with your database password(everything else is optional and has sensible defaults in the compose file below):
```
  POSTGRES_PASSWORD=YOUR_DB_PASSWORD_GOES_HERE
``` 

> [!NOTE]
> The password is only applied when the database is initialized for the first time. Run `ALTER ROLE opengymvault WITH PASSWORD 'new-password';` inside the postgres container and update your .env if you want to change it later.

3. Create a docker-compose.yml file like this one:
```
name: opengymvault

services:
  migrate:
    image: ghcr.io/javiermm8/opengymvault:latest
    entrypoint: ["/migrate"]
    command: ["up"]
    restart: "no"
    environment:
      DATABASE_URL: postgres://opengymvault:${POSTGRES_PASSWORD:?create a .env file with POSTGRES_PASSWORD first}@postgres:5432/${POSTGRES_DB:-opengymvault}?sslmode=disable
    depends_on:
      postgres:
        condition: service_healthy

  app:
    image: ghcr.io/javiermm8/opengymvault:latest
    restart: unless-stopped
    environment:
      DATABASE_URL: postgres://opengymvault:${POSTGRES_PASSWORD:?create a .env file with POSTGRES_PASSWORD first}@postgres:5432/${POSTGRES_DB:-opengymvault}?sslmode=disable
      REDIS_ADDR: redis:6379
      HTTP_ADDR: ":8080"
      RATE_LIMIT_AUTH: ${RATE_LIMIT_AUTH:-5}
      RATE_LIMIT_DEFAULT: ${RATE_LIMIT_DEFAULT:-100}
      # Comma-separated IPs/CIDRs of reverse proxies you control.
      # Leave empty unless running behind nginx/Caddy/etc.
      TRUSTED_PROXIES: ${TRUSTED_PROXIES:-}
    ports:
      - "${HTTP_PORT:-8080}:8080"
    depends_on:
      migrate:
        condition: service_completed_successfully
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy

  postgres:
    image: postgres:18-alpine
    restart: unless-stopped
    environment:
      POSTGRES_USER: opengymvault
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:?create a .env file with POSTGRES_PASSWORD first}
      POSTGRES_DB: ${POSTGRES_DB:-opengymvault}
    volumes:
      - pgdata:/var/lib/postgresql
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U opengymvault"]
      interval: 5s
      timeout: 5s
      retries: 5

  redis:
    image: redis:8-alpine
    restart: unless-stopped
    volumes:
      - redis-data:/data
    healthcheck:
      test: ["CMD-SHELL", "redis-cli ping | grep PONG"]
      interval: 5s
      timeout: 5s
      retries: 5

volumes:
  pgdata:
  redis-data:

```
4. Then, run  `docker compose pull && docker compose up -d`

## Screenshots
_(Obviously dummy credentials)_

/openapi:
![](/docs/assets/openapi.png)

/auth/register example:
![](/docs/assets/register.png)

/new_session example:
![](/docs/assets/new_session.png)

The rate limiter surviving an attack:
![](/docs/assets/rate_limiter.png)

The session cookie saved in a browser:
![](/docs/assets/cookie.png)
