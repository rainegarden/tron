package runconfig

type ProjectType string

const (
	ProjectTypeNone    ProjectType = ""
	ProjectTypeDjango  ProjectType = "django"
	ProjectTypeFlask   ProjectType = "flask"
	ProjectTypeFastAPI ProjectType = "fastapi"
	ProjectTypePython  ProjectType = "python"
)

type RunCommandMsg struct {
	Config *RunConfig
}

type ConfigSelectedMsg struct {
	Index int
	Name  string
}

type ShowDropdownMsg struct{}

type HideDropdownMsg struct{}

type EditConfigMsg struct {
	Index int
}

type DefaultConfig struct {
	Name    string
	Command string
	Args    []string
}

var DefaultConfigsByType = map[ProjectType][]DefaultConfig{
	ProjectTypeDjango: {
		{Name: "Run Server", Command: "python", Args: []string{"manage.py", "runserver"}},
		{Name: "Run Tests", Command: "python", Args: []string{"manage.py", "test"}},
		{Name: "Migrate", Command: "python", Args: []string{"manage.py", "migrate"}},
	},
	ProjectTypeFlask: {
		{Name: "Run Server", Command: "flask", Args: []string{"run"}},
		{Name: "Run App", Command: "python", Args: []string{"app.py"}},
	},
	ProjectTypeFastAPI: {
		{Name: "Run Server", Command: "uvicorn", Args: []string{"main:app", "--reload"}},
		{Name: "Run Tests", Command: "pytest", Args: []string{}},
	},
	ProjectTypePython: {
		{Name: "Run File", Command: "python", Args: []string{}},
	},
}
