package main

import (
	"log"
	"os"

	"github.com/oskar-r/yals/cmd"

	"github.com/urfave/cli"
)

var version = "undefined"

var standardFlags = []cli.Flag{
	cli.StringFlag{
		Name:  "parser-func, f",
		Usage: "Named parser function (can not be combined with the regex and dateformat flag)",
	},
	cli.StringFlag{
		Name:  "regex, r",
		Usage: "Regular expresion with named capture groups. To indicate type in catching group use (?<name99t>) where t can be i (int), f(float64) or t(time)",
	},
	cli.StringFlag{
		Name:  "date-format, df",
		Usage: "Date format of provided timestamp (e.g. 02/Jan/2006:15:04:05 -0700)",
	},
	cli.StringFlag{
		Name:  "log-file, l",
		Usage: "`Log file to watch`",
	},
}

func main() {
	app := cli.NewApp()
	app.Name = "yals"
	app.Version = version
	app.HelpName = "Yet another log stasher"
	app.Usage = "This is a command line utility to tail a log file, parse it and ship it to a stash"

	app.Commands = []cli.Command{
		cli.Command{
			Name:   "bigquery",
			Usage:  "stash your logs in bigquery",
			Flags:  standardFlags,
			Action: cmd.WireUp,
		},
		cli.Command{
			Name:   "influx",
			Usage:  "stash your logs in influx db",
			Flags:  standardFlags,
			Action: cmd.WireUp,
		},
	}
	for _, v := range cmd.BQFlags {
		app.Commands[0].Flags = append(app.Commands[0].Flags,
			cli.StringFlag{
				Name:  v.Name + ", " + v.ShortName,
				Usage: v.Usage,
			})
	}
	for _, v := range cmd.InfluxFlags {
		app.Commands[1].Flags = append(app.Commands[1].Flags,
			cli.StringFlag{
				Name:  v.Name + ", " + v.ShortName,
				Usage: v.Usage,
			})
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
