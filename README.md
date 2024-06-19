# eternalpose

Minimal manga alerting application as a first project in Go. I figured this would be a good first project as it covers
web scraping, error handling, text parsing, scheduling, and how to deploy go code to containers.

I read One Piece weekly, and my friends quite like Jujutsu Kaisen. 
This will let us know when the next chapter is out!

## Configuration

Manga can be specified in `mangas.json` by specifying a manga title, 
a regex term to search for on TCBScans, and the current chapter. This JSON is also used as simple storage of 
the current chapter so as not to resend the same chapter many times.

```json
[
    {
        "title": "One Piece",
        "regex": "One Piece Chapter (?P<Chapter>\\d{4})$",
        "users": [
            "<@173232081575346178>"
        ],
        "currentChapter": 1117
    },
    {
        "title": "Jujutsu Kaisen",
        "regex": "Jujutsu Kaisen Chapter (?P<Chapter>\\d{3})$",
        "users": [
            "<c@228997488563060736>",
            "<@262637906865291264>"
        ],
        "currentChapter": 262
    }
]
```

## Running

Set up environment variables in a file called .env

```shell
export WEBHOOK_URL=https://discord.com/api/webhooks/123/456
```

Source those env files with:

```shell
$ source .env
```

Finally, it can be run:

```shell
$ go build cmd/eternalpose/main.go
$ ./main
```

The program will run eternally, and run the scheduled manga scraping job once per hour, every day of the week.

## Dockerizing

The docker image can be built pretty easily:

```shell
$ docker build . --tag eternalpose
```

Then run the dockerfile with the environment variable:

```shell
$ docker run -dit -e WEBHOOK_URL="https://discord.com/api/webhooks/123/456" --name eternalpose-service eternalpose
```

## Deploying to Kubernetes

First, the discord webhook needs to be initialized as a secret:

```shell
$ kubectl create secret generic discord-webhook -n eternalpose --from-literal="discord-webhook=https://discord.com/api/webhooks/123/456"
```

Once the secret has been created, the application can be deployed:

```shell
$ kubectl apply -f eternalpose.yml
```