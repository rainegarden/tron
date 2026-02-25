package runconfig

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

func DetectProjectType(rootPath string) ProjectType {
	if hasDjango(rootPath) {
		return ProjectTypeDjango
	}
	if hasFlask(rootPath) {
		return ProjectTypeFlask
	}
	if hasFastAPI(rootPath) {
		return ProjectTypeFastAPI
	}
	if hasPythonFiles(rootPath) {
		return ProjectTypePython
	}
	return ProjectTypeNone
}

func hasDjango(rootPath string) bool {
	managePath := filepath.Join(rootPath, "manage.py")
	if _, err := os.Stat(managePath); err == nil {
		return true
	}

	if hasDependency(rootPath, "django") {
		return true
	}

	return false
}

func hasFlask(rootPath string) bool {
	appPath := filepath.Join(rootPath, "app.py")
	if content, err := os.ReadFile(appPath); err == nil {
		if strings.Contains(string(content), "flask") || strings.Contains(string(content), "Flask") {
			return true
		}
	}

	if hasDependency(rootPath, "flask") {
		return true
	}

	return false
}

func hasFastAPI(rootPath string) bool {
	mainPath := filepath.Join(rootPath, "main.py")
	if content, err := os.ReadFile(mainPath); err == nil {
		if strings.Contains(string(content), "fastapi") || strings.Contains(string(content), "FastAPI") {
			return true
		}
	}

	if hasDependency(rootPath, "fastapi") {
		return true
	}

	return false
}

func hasPythonFiles(rootPath string) bool {
	entries, err := os.ReadDir(rootPath)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".py") {
			return true
		}
	}

	return false
}

func hasDependency(rootPath, depName string) bool {
	reqPath := filepath.Join(rootPath, "requirements.txt")
	if content, err := os.ReadFile(reqPath); err == nil {
		if containsDependency(string(content), depName) {
			return true
		}
	}

	pyprojectPath := filepath.Join(rootPath, "pyproject.toml")
	if content, err := os.ReadFile(pyprojectPath); err == nil {
		if containsDependencyPyproject(string(content), depName) {
			return true
		}
	}

	return false
}

func containsDependency(content, depName string) bool {
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		line = strings.ToLower(line)
		if strings.HasPrefix(line, strings.ToLower(depName)) {
			return true
		}
	}
	return false
}

func containsDependencyPyproject(content, depName string) bool {
	depNameLower := strings.ToLower(depName)
	contentLower := strings.ToLower(content)

	dependenciesSection := ""
	inDeps := false
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[project.dependencies]") || strings.HasPrefix(trimmed, "[tool.poetry.dependencies]") {
			inDeps = true
			continue
		}
		if inDeps && strings.HasPrefix(trimmed, "[") {
			inDeps = false
		}
		if inDeps {
			dependenciesSection += line + "\n"
		}
	}

	if dependenciesSection != "" {
		return containsDependency(dependenciesSection, depName)
	}

	return strings.Contains(contentLower, depNameLower)
}
