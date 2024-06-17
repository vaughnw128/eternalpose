# eternalpose

Minimal manga alerting application as a first project in Go.

I read One Piece weekly, and my friends quite like Jujutsu Kaisen. 
This will let us know when the next chapter is out!

## Configuration

Manga can be specified in `mangas.json` by specifying a manga title, 
a regex term to search for on TCBScans, and the current chapter.

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

```bash
$ go build cmd/eternalpose/main.go
$ ./main
```