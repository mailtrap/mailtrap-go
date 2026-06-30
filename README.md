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

Create a client with your API token:

```go
client, err := mailtrap.NewClient("your-api-token")
if err != nil {
	log.Fatal(err)
}
```

## Supported functionality & Examples

Email Sandbox (Testing):

- Project management — [`examples/projects`](examples/projects)

## Errors

Non-2xx responses decode into typed errors that work with `errors.As`:

```go
_, _, err := client.Projects.List(ctx)

var ve *mailtrap.ValidationError
if errors.As(err, &ve) {
	fmt.Println(ve.Fields) // field -> messages
}

var rle *mailtrap.RateLimitError
if errors.As(err, &rle) {
	time.Sleep(rle.RetryAfter)
}
```

## Contributing

Bug reports and pull requests are welcome on GitHub at https://github.com/mailtrap/mailtrap-go. This project is intended to be a safe, welcoming space for collaboration, and contributors are expected to adhere to the [Code of Conduct](CODE_OF_CONDUCT.md).

## License

This library is released under the [MIT License](LICENSE).
