package main

import (
	"context"
	"fmt"
	"log"

	"github.com/mailtrap/mailtrap-go"
)

const (
	apiToken  = "your-api-token"
	accountID = 1000001
)

func main() {
	client, err := mailtrap.NewClient(apiToken, mailtrap.WithAccountID(accountID))
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	project, _, err := client.Projects.Create(ctx, "Demo project")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("created project %d (%s)\n", project.ID, project.Name)

	projects, _, err := client.Projects.List(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("account has %d project(s)\n", len(projects))

	project, _, err = client.Projects.Get(ctx, project.ID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("fetched project %d (%s)\n", project.ID, project.Name)

	_, _, err = client.Projects.Update(ctx, project.ID, "Renamed project")
	if err != nil {
		log.Fatal(err)
	}

	_, err = client.Projects.Delete(ctx, project.ID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("deleted project")
}
