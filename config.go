package main

import (
	"flag"
	"log"
	"os"

	"github.com/peterbourgon/ff/v3"
)

type Config struct {
	UseTLS       bool
	Domain       string
	FTPPort      int
	WebPort      int
	StderrLogger log.Logger
	StdoutLogger log.Logger
}

func GenConfig() Config {
	log.Println("Read configurations.")
	fs := flag.NewFlagSet("easygoftp", flag.ContinueOnError)
	var (
		useTLS  = fs.Bool("useTLS", true, "if to use TLS for connections")
		ftpPort = fs.Int("ftpPort", 21, "default port of ftp server")
		webPort = fs.Int("webPort", 8080, "default port of web server")
		domain  = fs.String("domain", "", "domain name for ftp server, necessary if useTLS is checked")
	)

	if _, err := os.Stat(".env"); os.IsNotExist(err) {
		fs.String("config", "", "config file")
	} else {
		fs.String("config", ".env", "config file")
	}

	err := ff.Parse(fs, os.Args[1:],
		ff.WithEnvVars(),
		ff.WithConfigFileFlag("config"),
		ff.WithConfigFileParser(ff.EnvParser),
	)
	if err != nil {
		log.Fatalf("Unable to parse args. Error: %s", err)
	}

	return Config{
		UseTLS:       *useTLS,
		FTPPort:      *ftpPort,
		WebPort:      *webPort,
		Domain:       *domain,
		StderrLogger: *log.New(os.Stderr, "", log.LstdFlags),
		StdoutLogger: *log.New(os.Stderr, "", log.LstdFlags),
	}
}
