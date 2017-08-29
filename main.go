package main

import (
	"cloud.google.com/go/datastore"
	"cloud.google.com/go/pubsub"
	"context"
	"fmt"
	"github.com/asiragusa/wschat/application"
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/recover"
	"github.com/urfave/cli"
	"os"
)

func main() {
	app := cli.NewApp()

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "projectID",
			Value:  "test",
			Usage:  "Google Cloud Project ID",
			EnvVar: "PROJECT_ID",
		},
		cli.StringFlag{
			Name:   "jwtSecret",
			Value:  "default",
			Usage:  "JWT secret",
			EnvVar: "JWT_SECRET",
		},
		cli.StringFlag{
			Name:   "jwtIssuer",
			Value:  "http://localhost",
			Usage:  "Jwt issuer eg. http://myapp.com",
			EnvVar: "JWT_ISSUER",
		},
	}

	app.Action = cliMain
	app.Run(os.Args)
}

func getDatastoreClient(projectID string) (*datastore.Client, error) {
	ctx := context.Background()

	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func getPubsubClient(projectID string) (*pubsub.Client, error) {
	ctx := context.Background()

	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func cliMain(c *cli.Context) error {
	datastoreClient, err := getDatastoreClient(c.String("projectID"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	pubsubClient, err := getPubsubClient(c.String("projectID"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	appConfig := &application.AppConfig{
		JwtSecret:       c.String("jwtSecret"),
		JwtIssuer:       c.String("jwtIssuer"),
		DatastoreClient: datastoreClient,
		PubsubClient:    pubsubClient,
	}

	app, err := application.NewApplication(appConfig)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	router := app.GetRouter()
	router.Use(recover.New())
	router.Run(iris.Addr(":80"))

	return nil
}
