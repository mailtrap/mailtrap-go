package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/mailtrap/mailtrap-go"
)

func main() {
	apiToken := os.Getenv("MAILTRAP_API_TOKEN")
	projectID, _ := strconv.ParseInt(os.Getenv("MAILTRAP_PROJECT_ID"), 10, 64)

	client, err := mailtrap.NewClient(apiToken)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	sandbox, _, err := client.Sandboxes.Create(ctx, projectID, "QA sandbox")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("created sandbox %d (%s)\n", sandbox.ID, sandbox.Name)

	sandboxes, _, err := client.Sandboxes.List(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("account has %d sandbox(es)\n", len(sandboxes))

	sandbox, _, err = client.Sandboxes.Get(ctx, sandbox.ID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("sandbox status: %s\n", sandbox.Status)

	_, _, err = client.Sandboxes.Update(ctx, sandbox.ID, &mailtrap.SandboxUpdateRequest{Name: "Renamed sandbox"})
	if err != nil {
		log.Fatal(err)
	}

	_, _, err = client.Sandboxes.Clean(ctx, sandbox.ID)
	if err != nil {
		log.Fatal(err)
	}

	_, _, err = client.Sandboxes.ResetCredentials(ctx, sandbox.ID)
	if err != nil {
		log.Fatal(err)
	}

	_, _, err = client.Sandboxes.Delete(ctx, sandbox.ID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("deleted sandbox")
}
