package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/mailtrap/mailtrap-go"
)

func main() {
	apiToken := os.Getenv("MAILTRAP_API_TOKEN")
	accountID, _ := strconv.ParseInt(os.Getenv("MAILTRAP_ACCOUNT_ID"), 10, 64)

	client, err := mailtrap.NewClient(apiToken)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	tokens, _, err := client.APITokens.List(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("account has %d API token(s)\n", len(tokens))

	token, _, err := client.APITokens.Create(ctx, &mailtrap.CreateAPITokenRequest{
		Name: "CI token",
		// Omit ExpiresAt for the server default expiration, or pass
		// mailtrap.NeverExpires() for a token that never expires.
		ExpiresAt: mailtrap.ExpiresAt(time.Now().AddDate(0, 0, 30).Format(time.RFC3339)),
		Resources: []*mailtrap.APITokenPermission{
			{ResourceType: mailtrap.ResourceTypeAccount, ResourceID: accountID, AccessLevel: mailtrap.AccessLevelViewer},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	// The full token value is only returned by Create and Reset — store it securely.
	fmt.Printf("created token %d (expires %s): %s\n", token.ID, token.ExpiresAt, token.Token)

	if _, _, err = client.APITokens.Get(ctx, token.ID); err != nil {
		log.Fatal(err)
	}

	// Reset expires the token and issues a replacement with the same permissions.
	// Pass nil instead of a request to apply the server default expiration.
	token, _, err = client.APITokens.Reset(ctx, token.ID, &mailtrap.ResetAPITokenRequest{
		ExpiresAt: mailtrap.NeverExpires(),
	})
	if err != nil {
		log.Fatal(err)
	}

	if _, err = client.APITokens.Delete(ctx, token.ID); err != nil {
		log.Fatal(err)
	}
	fmt.Println("deleted token")
}
