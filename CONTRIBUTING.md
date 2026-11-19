# Adding game covers

Covers live in [`internal/covers/catalog.json`](internal/covers/catalog.json):

```json
{
  "PPSA26344": "https://example.com/ghost-of-yotei-1024.png"
}
```

Rules (checked by `go test ./...`):
- key is the PS5 title ID, `PPSA` + 5 digits (the companion logs it on every game change)
- value is a public `https` image URL, square, ideally 1024x1024
- only submit images you are allowed to link to; do not commit image files

Users can also override or extend the catalog locally with `covers.json`.
