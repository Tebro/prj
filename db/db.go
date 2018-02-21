package db

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
)

var configPath = filepath.Join(os.Getenv("HOME"), ".prj")
var dbPath = filepath.Join(configPath, "db.json")
var database Database

// Config contains various configuratble variables
type Config struct {
	BaseDir            string
	AlwaysGit          bool
	EditorInBackground bool
}

func (c Config) String() string {
	return fmt.Sprintf(
		`Configuration options
Name: value
-----------
BaseDir: %s
AlwaysGit: %t
EditorInBackground: %t
`, c.BaseDir, c.AlwaysGit, c.EditorInBackground)
}

// Project describes a Project, contains a name and a path
type Project struct {
	Name string
	Path string
}

// Database is the top level object that the software uses to persist data and configuration
type Database struct {
	Config   Config
	Projects map[string]Project
}

func serializeDatabase(db Database) ([]byte, error) {
	data, err := json.MarshalIndent(db, "", "    ")
	return data, err
}

func deserializeDatabase(data []byte) (Database, error) {
	var decoded Database
	err := json.Unmarshal(data, &decoded)
	return decoded, err
}

func saveDatabase(db Database) {
	data, err := serializeDatabase(db)
	if err != nil {
		panic("Unable to serialize database, something is badly wrong")
	}

	err = ioutil.WriteFile(dbPath, data, 0644)
	if err != nil {
		panic("unable to write database to save file, you are fucked")
	}
}

func loadDatabase() Database {
	data, err := ioutil.ReadFile(dbPath)
	if err != nil {
		panic("Could not load existing database")
	}

	db, err := deserializeDatabase(data)
	if err != nil {
		panic("Unable to read database, data might be corrupted")
	}

	return db
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

func databaseExists() bool {
	exists, err := pathExists(dbPath)
	if err != nil {
		panic("Could not determine if database exists")
	}
	return exists
}

func createDefaultDatabase() Database {
	return Database{
		Config: Config{
			BaseDir:            fmt.Sprintf("%s/%s", os.Getenv("HOME"), "Projects"),
			AlwaysGit:          false,
			EditorInBackground: false,
		},
		Projects: make(map[string]Project),
	}
}

func configPathExists() bool {
	exists, err := pathExists(configPath)
	if err != nil {
		panic("Could not determine if $HOME/.prj exists")
	}
	return exists
}

func createSavePath() {
	err := os.MkdirAll(configPath, 0755)
	if err != nil {
		panic("Could not create config directory ($HOME/.prj)")
	}
}

func init() {
	if !configPathExists() {
		createSavePath()
	}

	if databaseExists() {
		database = loadDatabase()
	} else {
		database = createDefaultDatabase()
	}
}

// PrepareForShutdown tells the db package to save the database to disk in preparation for program termination. Should be the last call before the program exits.
func PrepareForShutdown() {
	saveDatabase(database)
}

// GetConfigList returns the Config objects String representation from the database.
func GetConfigList() string {
	return database.Config.String()
}

// SetConfigOption is a wrapper for modifying the Config part of the database.
func SetConfigOption(key string, value string) {
	switch key {
	case "BaseDir":
		database.Config.BaseDir = value
		break
	case "AlwaysGit":
		converted := value == "true"
		database.Config.AlwaysGit = converted
		break
	case "EditorInBackground":
		database.Config.EditorInBackground = value == "true"
		break
	}
}

// GetConfigBaseDir returns the BaseDir option from the configuration
func GetConfigBaseDir() string {
	return database.Config.BaseDir
}

// GetConfigAlwaysGit returns the AlwaysGit option from the configuration
func GetConfigAlwaysGit() bool {
	return database.Config.AlwaysGit
}

// GetConfigEditorInBackground returns the EditorInBackground option from the configuration
func GetConfigEditorInBackground() bool {
	return database.Config.EditorInBackground
}

// AddProject adds a new Project to the Database
func AddProject(name string, path string) error {
	if _, ok := database.Projects[name]; ok {
		return fmt.Errorf("project exists")
	}

	database.Projects[name] = Project{Name: name, Path: path}

	return nil
}

// GetProjects returns a list of all projects in the Database
func GetProjects() []Project {
	var projects []Project
	for _, v := range database.Projects {
		projects = append(projects, v)
	}
	return projects
}

// ListProjects returns a string representation of all the projects in the Database
func ListProjects() string {
	retval := ""

	projects := GetProjects()

	sort.Slice(projects, func(a int, b int) bool {
		return projects[a].Path < projects[b].Path
	})

	for _, v := range projects {
		retval = fmt.Sprintf("%s%s: %s\n", retval, v.Name, v.Path)
	}

	return retval
}

// GetProjectDir returns the path of a project identified by name
func GetProjectDir(name string) (string, error) {
	if _, ok := database.Projects[name]; !ok {
		return "", fmt.Errorf("project does not exists")
	}
	return database.Projects[name].Path, nil
}

// DeleteProject deletes a project from the Database
func DeleteProject(name string) {
	delete(database.Projects, name)
}
