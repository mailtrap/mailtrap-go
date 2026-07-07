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
	client, err := mailtrap.NewClient(os.Getenv("MAILTRAP_API_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	resources, _, err := client.Permissions.Resources(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("top-level resources: %d\n", len(resources))

	// Permissions are attached to an account access, so grab one to update.
	accesses, _, err := client.AccountAccesses.List(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	if len(accesses) == 0 || len(resources) == 0 {
		return
	}

	// Grant viewer access on a resource; pass Destroy: true to remove a permission.
	_, err = client.Permissions.BulkUpdate(ctx, accesses[0].ID, []*mailtrap.PermissionUpdate{
		{
			ResourceID:   strconv.FormatInt(resources[0].ID, 10),
			ResourceType: resources[0].Type,
			AccessLevel:  mailtrap.PermissionLevelViewer,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}
