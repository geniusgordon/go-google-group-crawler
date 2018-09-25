# Go Google Group Crawler

A google group crawler written in go

This crawler downloads messages in `RFC 822` format that can be imported
into your email clients (`Mutt`, `Thunderbird`, etc.)

![mutt screenshot](https://raw.githubusercontent.com/geniusgordon/go-google-group-crawler/master/mutt.png)

> Example of messages downloaded from [vim_use](https://groups.google.com/forum/#!forum/vim_use)

## Usage

```
$ go run crawler.go -h
Usage of crawler:
  -g string
        Group name
  -t int
        Threads count (default 1)
```

## Todos

- [ ] Download new messages
- [ ] Private group

## Credits

This project is inspired by [icy/google-group-crawler](https://github.com/icy/google-group-crawler), a bash version of the crawler

## License

Go Google Group Crawler is open source software [licensed as MIT](https://github.com/geniusgordon/go-google-group-crawler/blob/master/LICENSE).
