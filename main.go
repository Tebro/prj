package main

import (
	"gopkg.in/urfave/cli.v1"
	"fmt"
	"os"
	"path/filepath"
	"github.com/Tebro/prj/db"
	"os/exec"
	"strings"
)

func main() {

	app := cli.NewApp()
	app.Version = "0.1.0"
	app.Name = "prj"
	app.Description = "A project management tool"
	app.EnableBashCompletion = true

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "basedir, b",
			Usage: "The base directory to use (overrides global configuration)",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:  "config",
			Usage: "change configuration options",
			Subcommands: []cli.Command{
				{
					Name:   "list",
					Aliases: []string{"l", "ls"},
					Usage:  "Lists all configuration options",
					Action: listConfig,
				},
				{
					Name:      "set",
					Usage:     "Set a global configuration option in the database",
					ArgsUsage: "[key] [value]",
					Action:    setConfig,
				},
			},
		},
		{
			Name:      "new",
			Aliases:   []string{"n"},
			Usage:     "Create a new project",
			ArgsUsage: "[name]",
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name: "categories, c",
					Usage: "Optional organising levels, each category gets created in between the base dir and actual project dir",
				},
				cli.BoolFlag{
					Name:  "git, g",
					Usage: "Create a git repository",
				},
				cli.StringFlag{
					Name:  "name",
					Usage: "Explicitly set name of project for database (when names might otherwise clash)",
				},
			},
			Action: createNew,
		},
		{
			Name:      "goto",
			Aliases:   []string{"g"},
			Usage:     "Prints command to go to project directory, meant to be eval'ed",
			ArgsUsage: "[name]",
			Action:    printGoToCommand,
		},
		{
			Name:    "list",
			Aliases: []string{"l", "ls"},
			Usage:   "Prints your projects with their respective paths",
			Action:  listProjects,
		},
	}
	app.Run(os.Args)
	db.PrepareForShutdown()
}


func log(c *cli.Context, format string, args ...interface{}) {
	fmt.Fprintf(c.App.Writer, format + "\n", args...)
}

func getBaseDir(c *cli.Context) string {
	if len(c.GlobalString("basedir")) > 0 {
		return c.GlobalString("basedir")
	}
	return db.GetConfigBaseDir()
}

func getFinalPath(c *cli.Context) string {
	base := getBaseDir(c)
	cats := c.StringSlice("categories")
	catPath := strings.Join(cats, "/")
	name := c.Args()[0]

	return filepath.Join(base, catPath, name)
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func getProjectName(c *cli.Context) string {
	if len(c.String("name")) > 0 {
		return c.String("name")
	}
	return c.Args()[0]

}

func shouldCreateGit(c *cli.Context) bool {
	return c.Bool("git") || db.GetConfigAlwaysGit()
}

func createNew(c *cli.Context) error {
	if c.NArg() <= 0 {
		log(c,"name is required")
		return fmt.Errorf("name is required")
	}

	finalPath := getFinalPath(c)

	exists, err := pathExists(finalPath)
	if err != nil {
		log(c, err.Error())
		return err
	}

	if exists {
		err = fmt.Errorf("path %s exits", finalPath)
		log(c, err.Error())
		return err
	}

	projectName := getProjectName(c)

	err = db.AddProject(projectName, finalPath)
	if err != nil {
		log(c, err.Error())
		return err
	}

	err = os.MkdirAll(finalPath, 0755)
	if err != nil {
		log(c,"Could not create project directory")
		return err
	}

	if shouldCreateGit(c) {
		os.Chdir(finalPath)
		cmd := exec.Command("git", "init")
		err := cmd.Run()
		if err != nil {
			log(c,"Failed to create git repository. Error %s", err.Error())
		}
		log(c,"Creating Git repository")
	}

	log(c,"Created project")

	return nil
}

func listConfig(c *cli.Context) error {
	log(c, db.GetConfigList())
	return nil
}

func setConfig(c *cli.Context) error {
	if c.NArg() != 2 {
		log(c,"Invalid number of arguments, expected 2")
		return fmt.Errorf("invalid number of arguments")
	}

	db.SetConfigOption(c.Args()[0], c.Args()[1])

	return nil
}

func printGoToCommand(c *cli.Context) error {
	if c.NArg() != 1 {
		log(c,"Invalid number of arguments, expected 1")
		return fmt.Errorf("invalid number of arguments")
	}
	path, err := db.GetProjectDir(c.Args()[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	} else {
		log(c,"cd %s", path)
	}

	return nil
}

func listProjects(c *cli.Context) error {
	msg := fmt.Sprintf(
`Projects
--------
%s`, db.ListProjects())

	log(c, msg)
	return nil
}
