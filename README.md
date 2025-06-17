# HAR file management library

A Golang library for parsing and managing HAR (HTTP Archive) files. Provides easy-to-use structs and functions for loading, inspecting, and manipulating HAR files, making HTTP traffic analysis and debugging simpler.

- [X] Check for case sensitivity in HeaderOrder keys
- [ ] Add Remote Address (proxy used)
- [ ] Fix bodySize

## How to compile

### MacOS

```bash
cd example
GOOS=darwin GOARCH=arm64 go build -o example-harkit main.go
```
