package cmd

import (
	"errors"
	"log"

	"github.com/urfave/cli"
)

var parserFuncs = map[string]func(string) (string, error){
	"formida-parser": formidaParser,
	"nginx-parser":   nginxParser,
}

type Flag struct {
	Name      string
	ShortName string
	Usage     string
}

//WireUp validates flags and set up dependencies
func WireUp(c *cli.Context) error {
	var p Parser
	//First set up the parser
	if c.String("parser-func") != "" {
		if prsr, ok := parserFuncs[c.String("parser-func")]; ok {
			p = NewParser("", "", prsr)
		} else {
			return errors.New("Parser " + c.String("parser") + " does not exist")
		}
	} else if c.String("parser") == "" && c.String("regex") != "" && c.String("date-format") != "" {
		p = NewParser(c.String("regex"), c.String("date-format"), nil)
	} else {
		return errors.New("Neither regexp/dateformat or parser-func provided")
	}
	if c.String("log-file") == "" {
		return errors.New("log-file not provided")
	}
	SetupParser(p)
	switch c.Command.Name {
	case "bigquery":
		for _, v := range BQFlags {
			if c.String(v.Name) == "" {
				return errors.New("No " + v.Name + " provided")
			}
			s, err := NewBqService(c.String(BQFlags["project"].Name), c.String(BQFlags["dataset"].Name), c.String(BQFlags["table"].Name), c.String(BQFlags["serviceaccount"].Name))
			if err != nil {
				log.Fatalf("[ERROR] %+v", err)
				return err
			}
			SetupService(s)
		}
	case "influx":
		for _, v := range InfluxFlags {
			if c.String(v.Name) == "" {
				return errors.New("No " + v.Name + " provided")
			}
		}
		s, err := NewInfluxService(c.String(InfluxFlags["databse"].Name), c.String(InfluxFlags["serie"].Name), c.String(InfluxFlags["connection"].Name))
		if err != nil {
			log.Fatalf("[ERROR] %+v", err)
			return err
		}
		SetupService(s)
	}

	t, err := NewHPTailer(p, c.String("log-file"))
	if err != nil {
		log.Fatalf("[ERROR] %+v", err)
		return err
	}

	SetupTail(t)
	Start()
	return nil
}
