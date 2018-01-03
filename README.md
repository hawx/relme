# rel-me

```go
me := "https://example.com"

links, _ := relme.Find(me)
for _, link := range links {
  if relme.LinksTo(link, me) {
    fmt.Printf("me=%s\n", link)
  }
}
```

See http://microformats.org/wiki/rel-me
