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

### Sandbox vs Production (easy switching)

Mailtrap lets you test safely in the Email Sandbox and then switch to Production sending with a single flag. In sandbox mode `Send` captures the email in a sandbox instead of delivering it to real recipients.

Keep the sandbox ID configured and toggle `WithSandbox` from configuration — the ID is ignored outside sandbox mode, so the exact same code switches environments without touching any call site:

```go
isSandbox := os.Getenv("MAILTRAP_USE_SANDBOX") == "true"

client, err := mailtrap.NewClient("your-api-token",
	mailtrap.WithSandbox(isSandbox),
	mailtrap.WithSandboxID(3000001), // ignored unless sandbox mode is on
)
if err != nil {
	log.Fatal(err)
}

resp, _, err := client.Send(context.Background(), &mailtrap.SendRequest{
	From:    mailtrap.Address{Email: "sender@example.com", Name: "Example"},
	To:      []mailtrap.Address{{Email: "recipient@example.com"}},
	Subject: "Hello from mailtrap-go",
	Text:    "Captured by the sandbox when MAILTRAP_USE_SANDBOX=true.",
})
```

## Supported functionality & Examples

Email API (sending):

- Transactional, bulk & batch sending — text/HTML, templates, attachments, categories, custom variables & headers — [`examples/sending`](examples/sending)

Email API (management):

- Sending domain management — DNS setup instructions & company info — [`examples/sending-domains`](examples/sending-domains)
- Suppression management — the do-not-send list of bounces, complaints & unsubscribes — [`examples/suppressions`](examples/suppressions)
- Sending stats — aggregated, or grouped by domain, category, ESP & date — [`examples/stats`](examples/stats)
- Email logs — filter & cursor-paginate sent messages, inspect delivery events — [`examples/email-logs`](examples/email-logs)
- Webhook management — create, list, update & delete event webhooks — [`examples/webhooks`](examples/webhooks)
- Email template management — create, list, update & delete templates — [`examples/email-templates`](examples/email-templates)

Email Sandbox (Testing):

- Sandbox sending — single & batch — [`examples/sandbox-sending`](examples/sandbox-sending)
- Project management — [`examples/projects`](examples/projects)
- Sandbox (inbox) management & actions — [`examples/sandboxes`](examples/sandboxes)
- Message management & inspection — spam report, HTML analysis, headers, raw bodies — [`examples/messages`](examples/messages)
- Attachment operations — [`examples/attachments`](examples/attachments)

Account & organization management:

- Accounts — list the accounts a token can access — [`examples/accounts`](examples/accounts)
- Account accesses — list & remove user/invite/token access — [`examples/account-accesses`](examples/account-accesses)
- Permissions — list resources & bulk-update access permissions — [`examples/permissions`](examples/permissions)

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
