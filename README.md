# Mailtrap Go client

[![CI](https://github.com/mailtrap/mailtrap-go/actions/workflows/ci.yml/badge.svg)](https://github.com/mailtrap/mailtrap-go/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/mailtrap/mailtrap-go.svg)](https://pkg.go.dev/github.com/mailtrap/mailtrap-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

The official Go client library for the [Mailtrap](https://mailtrap.io) email delivery platform — transactional and bulk sending, email sandbox (testing), and email marketing.

## Prerequisites

To get the most out of this Mailtrap.io Go SDK:

- [Create a Mailtrap account](https://mailtrap.io/signup)
- [Verify your domain](https://mailtrap.io/sending/domains)
- [Create an API token](https://mailtrap.io/api-tokens)

## Installation

```bash
go get github.com/mailtrap/mailtrap-go
```

Requires Go 1.23 or newer.

## Usage

```go
package main

import (
	"log"

	"github.com/mailtrap/mailtrap-go"
)

func main() {
	client, err := mailtrap.NewClient("your-api-token")
	if err != nil {
		log.Fatal(err)
	}
}
```

## Supported functionality & Examples

This SDK is under active development. Documentation and runnable examples are added to this section as each part of the API is implemented.

## Contributing

Bug reports and pull requests are welcome on GitHub at https://github.com/mailtrap/mailtrap-go. This project is intended to be a safe, welcoming space for collaboration, and contributors are expected to adhere to the [Code of Conduct](CODE_OF_CONDUCT.md).

## License

This library is released under the [MIT License](LICENSE).
