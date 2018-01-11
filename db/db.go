package db

import (
	"os"
	"fmt"
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"sort"
)

var configPath = filepath.Join(os.Getenv("HOME"), ".prj")
var dbPath = filepath.Join(configPath, "db.json")
var database Database


type Config struct {
	BaseDir   string
	AlwaysGit bool
}

func (c Config) String() string {
	return fmt.Sprintf(
`Configuration options
Name: value
-----------
BaseDir: %s
AlwaysGit: %t
`, c.BaseDir, c.AlwaysGit)
}

type Project struct {
	Name string
	Path string
}

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

func loadDatabase() (Database) {
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
			BaseDir:   fmt.Sprintf("%s/%s", os.Getenv("HOME"), "Projects"),
			AlwaysGit: false,
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

func PrepareForShutdown() {
	saveDatabase(database)
}

func GetConfigList() string {
	return database.Config.String()
}

func SetConfigOption(key string, value string) {
	switch key {
	case "BaseDir":
		database.Config.BaseDir = value
		break
	case "AlwaysGit":
		converted := value == "true"
		database.Config.AlwaysGit = converted
		break
	}
}

func GetConfigBaseDir() string {
	return database.Config.BaseDir
}

func GetConfigAlwaysGit() bool {
	return database.Config.AlwaysGit
}

func AddProject(name string, path string) error {
	if _, ok := database.Projects[name]; ok {
		return fmt.Errorf("project exists")
	}

	database.Projects[name] = Project{Name:name, Path: path}

	return nil
}

func ListProjects() string {
	retval := ""

	var projects []Project
	for _, v := range database.Projects {
		projects = append(projects, v)
	}

	sort.Slice(projects, func(a int, b int) bool {
		return projects[a].Path < projects[b].Path
	})

	for _, v := range projects {
		retval = fmt.Sprintf("%s%s: %s\n", retval, v.Name, v.Path)
	}

	return retval
}

func GetProjectDir(name string) (string, error) {
	if _, ok := database.Projects[name]; !ok {
		return "", fmt.Errorf("project does not exists")
	}
	return database.Projects[name].Path, nil
}

func DeleteProject(name string) {
	delete(database.Projects, name)
}