# mangoplus

[![Go Reference](https://pkg.go.dev/badge/github.com/raf555/mangoplus?status.svg)](https://pkg.go.dev/github.com/raf555/mangoplus?tab=doc)

Unofficial MangaPlus API Client.

## Installation

```
go get github.com/raf555/mangoplus@latest
```

## Usage

The simplest usage is as follows.

```go
import "github.com/raf555/mangoplus"

client, err := mangoplus.NewClient()
if err != nil {
    // handle error
}

_, err = client.Register(context.Background())
if err != nil {
    // handle error
}

manga, err := client.Title.GetTitleDetailV3(context.Background(), 100185)
if err != nil {
    // handle error
}
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## Disclaimer

> [!WARNING]  
this package is an unofficial API wrapper for the MangaPlus android application and is not affiliated with, endorsed by, or sponsored by Shueisha or MangaPlus. "MangaPlus" and all related content are trademarks of their respective owners. The API is undocumented and may change or break at any time. Use of this package may be subject to MangaPlus's Terms of Service; users are responsible for ensuring their use complies with applicable terms and laws.
