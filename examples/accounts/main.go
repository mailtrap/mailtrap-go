package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/mailtrap/mailtrap-go"
)

func main() {
	client, err := mailtrap.NewClient(os.Getenv("MAILTRAP_API_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	accounts, _, err := client.Accounts.List(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	for _, account := range accounts {
		fmt.Printf("account %d: %s (access levels %v)\n", account.ID, account.Name, account.AccessLevels)
	}
}
