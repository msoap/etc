# export TB_TOKEN=***
# docker build -t latin-bot .
# docker run -d --rm --name latin-bot --env TB_TOKEN=$TB_TOKEN -v $PWD:/db latin-bot
# docker logs -f latin-bot

# build image
FROM golang:alpine as go_builder

ENV CGO_ENABLED=0
ADD gt-cli.go .
RUN go build -v -trimpath -ldflags="-w -s" -o /go/bin/gt-cli gt-cli.go

# final image
FROM msoap/shell2telegram

COPY --from=go_builder /go/bin/gt-cli /app/gt-cli

RUN ln -s /app/gt-cli /app/gt-cli-uk; \
    ln -s /app/gt-cli /app/gt-cli-pt; \
    ln -s /app/gt-cli /app/gt-cli-ja

CMD ["-description", "Latin-bot", \
     "-cache", "3600", \
     "-one-thread", \
     "-shell", "", \
     "-persistent-users", \
     "-users-db", "/db/latin-bot.db", \
     "/:plain_text:desc=translate text", "/app/gt-cli", \
     "/uk:desc=translate text to ua", "/app/gt-cli-uk", \
     "/pt:desc=translate text to pt", "/app/gt-cli-pt", \
     "/ja:desc=translate text to ja", "/app/gt-cli-ja" \
    ]
