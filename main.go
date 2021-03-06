package main

import (
	"fmt"
	"github.com/Tebro/prj/db"
	"gopkg.in/urfave/cli.v1"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {

	app := cli.NewApp()
	app.Version = "0.5.5"
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
					Name:    "list",
					Aliases: []string{"l", "ls"},
					Usage:   "Lists all configuration options",
					Action:  listConfig,
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
					Name:  "categories, c",
					Usage: "Optional organising levels, each category gets created in between the base dir and actual project dir",
				},
				cli.BoolFlag{
					Name:  "git, g",
					Usage: "Create a git repository",
				},
				cli.StringFlag{
					Name:  "name, n",
					Usage: "Explicitly set name of project for database (when names might otherwise clash)",
				},
			},
			Action: createNew,
		},
		{
			Name:      "add",
			Usage:     "Add existing directory to prj. If path is left out the current directory will be used.",
			ArgsUsage: "[name] <[path]>",
			Action:    addExisting,
		},
		{
			Name:      "delete",
			Aliases:   []string{"remove", "rm"},
			Usage:     "Remove a project from prj",
			ArgsUsage: "[name]",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "nocache, f",
					Usage: "Also remove the directory",
				},
			},
			Action: removeProject,
		},
		{
			Name:      "goto",
			Aliases:   []string{"g"},
			Usage:     "Prints command to go to project directory, meant to be eval'ed",
			ArgsUsage: "[name]",
			Action:    printGoToCommand,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "editor, e",
					Usage: "Add $EDITOR startup command to the output",
				},
			},
			BashComplete: func(c *cli.Context) {
				projects := db.GetProjects()
				for _, p := range projects {
					fmt.Println(p.Name)
				}
			},
		},
		{
			Name:    "list",
			Aliases: []string{"l", "ls"},
			Usage:   "Prints your projects with their respective paths",
			Action:  listProjects,
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}
	db.PrepareForShutdown()
}

func log(c *cli.Context, format string, args ...interface{}) {
	fmt.Fprintf(c.App.Writer, format+"\n", args...)
}

func exitErrorWrapper(format string, args ...interface{}) *cli.ExitError {
	return cli.NewExitError(fmt.Sprintf(format, args...), 1)
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

func createBaseDirIfNotExists(c *cli.Context) error {
	path := db.GetConfigBaseDir()

	if len(c.GlobalString("basedir")) > 0 {
		path = c.GlobalString("basedir")
	}

	isDir, err := pathIsDir(path)
	if err != nil {
		return err
	}
	if !isDir {
		err = os.MkdirAll(path, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

func createNew(c *cli.Context) error {
	if c.NArg() <= 0 {
		return exitErrorWrapper("name is required")
	}

	if err := createBaseDirIfNotExists(c); err != nil {
		return exitErrorWrapper("could not find or create base dir : %s", err.Error())
	}

	finalPath := getFinalPath(c)

	exists, err := pathExists(finalPath)
	if err != nil {
		return err
	}

	if exists {
		return exitErrorWrapper("path %s exits", finalPath)
	}

	projectName := getProjectName(c)

	err = db.AddProject(projectName, finalPath)
	if err != nil {
		return err
	}

	err = os.MkdirAll(finalPath, 0755)
	if err != nil {
		return exitErrorWrapper("could not create project directory: %s", err.Error())
	}

	if shouldCreateGit(c) {
		os.Chdir(finalPath)
		cmd := exec.Command("git", "init")
		err := cmd.Run()
		if err != nil {
			log(c, "Failed to create git repository. Error %s", err.Error())
		} else {
			log(c, "Created Git repository")
		}
	}

	log(c, "Created project")

	return nil
}

func listConfig(c *cli.Context) error {
	log(c, db.GetConfigList())
	return nil
}

func setConfig(c *cli.Context) error {
	if c.NArg() != 2 {
		return exitErrorWrapper("invalid number of arguments, expected 2")
	}

	db.SetConfigOption(c.Args()[0], c.Args()[1])

	return nil
}

func printGoToCommand(c *cli.Context) error {
	if c.NArg() != 1 {
		return exitErrorWrapper("invalid number of arguments, expected 1")
	}
	path, err := db.GetProjectDir(c.Args()[0])
	if err != nil {
		return err
	}
	log(c, "cd %s;", path)
	if c.Bool("editor") {
		editor := os.Getenv("EDITOR")
		format := "%s . %s;"
		inBackground := ""
		if db.GetConfigEditorInBackground() {
			inBackground = "&"
		}

		log(c, format, editor, inBackground)
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

func pathIsDir(path string) (bool, error) {
	stat, err := os.Stat(path)
	if err == nil {
		return stat.IsDir(), nil
	}
	return false, err
}

func addExisting(c *cli.Context) error {
	if c.NArg() < 1 || c.NArg() > 2 {
		return exitErrorWrapper("invalid number of arguments, expected 1 or 2")
	}

	name := c.Args()[0]

	var path string
	if c.NArg() > 1 {
		path = c.Args()[1]
	} else {
		workDir, err := os.Getwd()
		if err != nil {
			return exitErrorWrapper("could not retrieve current working directory: %s", err.Error())
		}
		path = workDir
	}

	exists, err := pathExists(path)
	if err != nil {
		return exitErrorWrapper("could not determine if path exists: %s", err.Error())
	}
	if !exists {
		return exitErrorWrapper("path '%s' does not exist, use 'prj new' to create a new project", path)
	}

	isDir, err := pathIsDir(path)
	if err != nil {
		return exitErrorWrapper("could not determine if path is a directory: %s", err)
	}
	if !isDir {
		return exitErrorWrapper("path '%s' is not a directory", path)
	}

	err = db.AddProject(name, path)
	if err != nil {
		return exitErrorWrapper("could not add project: %s", err)
	}

	return nil
}

func removeProject(c *cli.Context) error {
	if c.NArg() != 1 {
		return exitErrorWrapper("invalid number of arguments, expected 1")
	}

	name := c.Args()[0]
	path, err := db.GetProjectDir(name)
	if err != nil {
		return exitErrorWrapper("could not delete project: %s", err.Error())
	}

	if c.Bool("nocache") {
		log(c, "Removing directory: %s", path)
		os.RemoveAll(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to remove project directory: %s", err.Error())
		}
	} else {
		log(c, "Leaving directory in place")
	}

	db.DeleteProject(name)
	log(c, "Project: '%s' deleted", name)

	return nil
}
