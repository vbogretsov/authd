package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/akamensky/argparse"
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

// Version indicates application version, should be set during the build via -X.
var Version = ""

type dbupArgT struct {
	dburl *string
}

type dbdownArgT struct {
	dburl *string
}

const (
	dburlHelp = "database url, example: postgres://user:password@host/dbname"
)

var (
	app = argparse.NewParser(
		"authdctl",
		"control programm for authd service")

	dbupCmd = app.NewCommand(
		"dbup",
		"migrate database to the latest version")
	dbupArg = dbupArgT{}

	dbdownCmd = app.NewCommand(
		"dbdown",
		"downgrade database to 1 version")
	dbdownArg = dbdownArgT{}

	versionCmd = app.NewCommand(
		"version",
		"print version and exit")
)

func init() {
	dbupArg.dburl = dbupCmd.String("d", "dburl", &argparse.Options{
		Required: true,
		Help:     dburlHelp,
	})

	dbdownArg.dburl = dbdownCmd.String("d", "dburl", &argparse.Options{
		Required: true,
		Help:     dburlHelp,
	})
}

func parseDBURL(dburl string) (string, string, error) {
	n := strings.Index(dburl, "://")
	if n == -1 {
		return "", "", errors.New("invalid database URL format")
	}

	return dburl[:n], dburl, nil
}

func createMigrator(dburl string) (*gormigrate.Gormigrate, error) {
	if migrations[len(migrations)-1].ID != Version {
		panic("migrations version does not match application version")
	}

	dialect, url, err := parseDBURL(dburl)
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(dialect, url)
	if err != nil {
		return nil, err
	}
	db.LogMode(true)

	return gormigrate.New(db, gormigrate.DefaultOptions, migrations), nil
}

func dbup(arg dbupArgT) error {
	migrator, err := createMigrator(*arg.dburl)
	if err != nil {
		return err
	}
	return migrator.Migrate()
}

func dbdown(arg dbdownArgT) error {
	migrator, err := createMigrator(*arg.dburl)
	if err != nil {
		return err
	}
	return migrator.RollbackLast()
}

func printVersion() error {
	fmt.Fprintf(os.Stdout, "%s\n", Version)
	return nil
}

func run() error {
	if err := app.Parse(os.Args); err != nil {
		return err
	}

	switch {
	case dbupCmd.Happened():
		return dbup(dbupArg)
	case dbdownCmd.Happened():
		return dbdown(dbdownArg)
	case versionCmd.Happened():
		return printVersion()
	default:
		return errors.New("unknown command")
	}
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
