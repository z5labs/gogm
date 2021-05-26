// Copyright (c) 2021 MindStand Technologies, Inc
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package main

import (
	"errors"
	"github.com/mindstand/gogm/v2/cmd/gogmcli/gen"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

//main is the main function
func main() {
	var debug bool

	app := &cli.App{
		Name:                 "gogmcli",
		HelpName:             "gogmcli",
		Version:              "2.0.0",
		Usage:                "used for neo4j operations from gogm schema",
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
				Usage:     "to generate link and unlink functions for nodes",
				Action: func(c *cli.Context) error {
					directory := c.Args().Get(0)

					if directory == "" {
						return errors.New("must specify directory")
					}

					if debug {
						log.Printf("generating link and unlink from directory [%s]", directory)
					}

					return gen.Generate(directory, debug)
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
		Copyright:              "© MindStand Technologies, Inc 2021",
		UseShortOptionHandling: true,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "debug",
				Aliases:     []string{"d"},
				Usage:       "execute in debug mode",
				Value:       false,
				Destination: &debug,
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
