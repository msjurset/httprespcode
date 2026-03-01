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
hrc 404
```

```
404: Not Found — Not Found: The requested resource could not be found but may be
available in the future. Subsequent requests by the client are permissible.
```

### Verbose Mode

```bash
hrc -v 404
```

```
404: Not Found — Not Found: The requested resource could not be found but may be
available in the future. Subsequent requests by the client are permissible.

───

Common causes: Typo in the URL, deleted resource, incorrect route configuration,
undeployed endpoint, or missing file on disk.

Real-world usage: The most recognized HTTP error. Returned when a URL doesn't map
to any resource. In REST APIs, returned for GET/PUT/DELETE on a resource ID that
doesn't exist. Some security-conscious APIs return 404 instead of 403 to hide the
existence of resources from unauthorized users.

Related codes: 410 Gone (resource existed but was permanently removed — use this
when you know it's intentional), 405 Method Not Allowed (URL exists but method is
wrong).

Troubleshooting: Check for URL typos, verify the resource exists, check route
configuration, and ensure the server/application is properly deployed.

RFC: RFC 7231, Section 6.5.4.
```

Output is color-coded in the terminal: status lines are bold and colored by class (2xx green, 3xx yellow, 4xx red, 5xx magenta), section labels are highlighted, and descriptions are dimmed.

### Interactive Mode

```bash
hrc
```

```
Enter HTTP status code (or 'q' to quit): 200
200: OK — OK: Standard response for successful HTTP requests...

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
