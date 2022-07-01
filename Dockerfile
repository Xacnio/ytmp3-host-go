FROM golang:1.17-alpine

RUN apk add gcc libc-dev python3 py3-pip ffmpeg

RUN pip3 install yt-dlp

WORKDIR /root/ytmp3-host-go

COPY go.mod go.mod

COPY go.sum go.sum

RUN go mod download

COPY . .

RUN go build .

EXPOSE 5001

CMD ["./ytmp3-host-go"]