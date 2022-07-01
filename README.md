# ytmp3-host-go

YT mp3 hosting script in Golang

MP3s are sending to Telegram Cloud Storage by bot and they aren't stored in your server. When you want to reach some mp3 from http server, the server forwarding you to telegram http server with reverse proxy.

- Storage on Telegram Cloud Storage
- Simple data management with Redis
- Save a MP3 from Telegram by Bot
- Dockerized & replicas are supported

# Screenshots

- [1](assets/1.png)
- [2](assets/2.png)
- [3](assets/3.png)

## Config

Create .env file and edit ([Sample](.env.sample))

## Run

```
> go run main.go
```

### Docker

```
> docker-compose --env-file ./.env.dev up -d
```