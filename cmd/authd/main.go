package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/akamensky/argparse"
	"github.com/jinzhu/gorm"
	"github.com/labstack/gommon/log"
	"github.com/streadway/amqp"

	_ "github.com/jinzhu/gorm/dialects/mssql"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"

	api "github.com/vbogretsov/authd/api"
	apiv1 "github.com/vbogretsov/authd/api/v1"
	"github.com/vbogretsov/authd/auth"
)

// Version indicates application version, should be set during the build via -X.
var Version = ""

const (
	name  = "authd"
	usage = "bearer token authentication service"
)

const (
	dburlHelp    = "database url, example: postgres://user:password@host/dbname"
	mqurlHelp    = "AMQP broker url, example: amqp://user:password@host:port"
	secretHelp   = "secret key used to generate access token"
	pwminlenHelp = "minimum allowed password length"
	portHelp     = "service TCP port number"
)

const (
	logfmtKube = "kubernetes"
	logfmtJSON = "json"
)

var (
	args argT
	argp = argparse.NewParser(fmt.Sprintf("%s %s", name, Version), usage)
)

var (
	logLevels = map[string]log.Lvl{
		"error": log.ERROR,
		"warn":  log.WARN,
		"info":  log.INFO,
		"debug": log.DEBUG,
	}
)

type argT struct {
	dburl    *string
	mqurl    *string
	secret   *string
	pwminlen *int
	port     *int
	loglvl   *string
}

func init() {
	args.dburl = argp.String("d", "dburl", &argparse.Options{
		Required: true,
		Help:     dburlHelp,
	})
	args.mqurl = argp.String("m", "mqurl", &argparse.Options{
		Required: true,
		Help:     mqurlHelp,
	})
	args.secret = argp.String("s", "secret", &argparse.Options{
		Required: true,
		Help:     secretHelp,
	})
	args.pwminlen = argp.Int("n", "pwminlen", &argparse.Options{
		Default: 6,
		Help:    pwminlenHelp,
	})
	args.port = argp.Int("p", "port", &argparse.Options{
		Default: 9000,
		Help:    portHelp,
	})

	loglvls := []string{}
	for k := range logLevels {
		loglvls = append(loglvls, k)
	}

	args.loglvl = argp.Selector("l", "loglevel", loglvls, &argparse.Options{
		Default: "info",
	})
}

func connectdb(url string) (*gorm.DB, error) {
	n := strings.Index(url, "://")
	if n == -1 {
		return nil, errors.New("invalid database URL format")
	}

	return gorm.Open(url[:n], url)
}

func run() error {
	if err := argp.Parse(os.Args); err != nil {
		return err
	}

	db, err := connectdb(*args.dburl)
	if err != nil {
		return err
	}
	db.LogMode(true)

	mq, err := amqp.Dial(*args.mqurl)
	if err != nil {
		return err
	}

	sender, err := mail.NewSender(mq, "maild")
	if err != nil {
		return err
	}

	// TODO: pass parameters from args
	cf := auth.DefaultConfig

	ap, err := auth.New(cf, db, time.Now, sender)
	if err != nil {
		return err
	}

	e := api.New()
	apiv1.Include(ap, e)

	e.Logger.SetLevel(logLevels[*args.loglvl])
	return e.Start(fmt.Sprintf(":%d", *args.port))
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	}
}
