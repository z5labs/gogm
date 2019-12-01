package main

import (
	"errors"
	"github.com/mindstand/gogm/cmd/gogmcli/gen"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

func main() {
	app := &cli.App{
		Name:                 "gogmcli",
		HelpName:             "gogmcli",
		Version:              "0.2.0",
		Usage: "used for neo4j operations from gogm schema",
		Description:          "cli for generating and executing migrations with gogm",
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			{
				Name: "generate",
				Aliases: []string{
					"g",
					"gen",
				},
				ArgsUsage: "directory to search and write to",
				Usage: "to generate link and unlink functions for nodes",
				Action: func(c *cli.Context) error {
					directory := c.Args().Get(0)

					if directory == "" {
						return errors.New("must specify directory")
					}

					log.Printf("generating link and unlink from directory [%s]", directory)

					return gen.Generate(directory)
				},
			},
		},
		Authors: []*cli.Author{
			{
				Name:  "Eric Solender",
				Email: "eric@mindstand.com",
			},
			{
				Name:  "Nikita Wootten",
				Email: "nikita@mindstand.com",
			},
		},
		Copyright:              "© MindStand Technologies 2019",
		UseShortOptionHandling: true,
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
