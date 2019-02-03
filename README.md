# rel-me [![](https://godoc.org/hawx.me/code/relme?status.svg)](https://godoc.org/hawx.me/code/relme)

See http://microformats.org/wiki/rel-me

#### Go

```sh
$ go get hawx.me/code/relme
```

```go
import "hawx.me/code/relme"

me := "https://example.com/me"

links, _ := relme.Find(me)
for _, link := range links {
  if ok, _ := relme.LinksTo(link, me); ok {
    fmt.Printf("me=%s\n", link)
  }
}
```


#### Command line

```sh
$ go get hawx.me/code/relme/cmd/relme
$ relme https://example.com/me
...
$ relme -verified https://example.com/me
...
```
