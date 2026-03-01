# hrc - HTTP Response Code

A command-line tool for looking up HTTP response status codes with detailed descriptions.

## Features

- Look up any HTTP status code with its standard name and detailed description
- Covers all standard codes: 1xx informational, 2xx success, 3xx redirection, 4xx client error, 5xx server error
- Includes RFC references and usage context
- Verbose mode (`-v`) with extended details: common causes, real-world usage, related codes, and troubleshooting tips
- Interactive mode for looking up multiple codes in a session

## Installation

### Prerequisites

- Go 1.16 or later

### Build

```bash
make build
```

### Deploy

```bash
make deploy
```

This builds the binary, installs the man page, installs zsh completions, and copies `hrc` to `~/.local/bin/`.

## Usage

### Single Lookup

```bash
# Look up a status code
hrc 200
hrc 404
hrc 503
```

### Verbose Mode

```bash
# Show extended details for a status code
hrc -v 404
hrc -v 503
```

### Interactive Mode

```bash
# Enter interactive mode
hrc

Enter HTTP status code (or 'q' to quit): 200
200: OK
OK: Standard response for successful HTTP requests...

Enter HTTP status code (or 'q' to quit): q
```

## Supported Status Codes

| Range | Category |
|-------|----------|
| 100-103 | Informational |
| 200-226 | Success |
| 300-308 | Redirection |
| 400-451 | Client Error |
| 500-511 | Server Error |

## License

MIT License
