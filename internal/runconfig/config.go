package runconfig

import (
	"os"
	"path/filepath"
)

type RunConfig struct {
	Name        string
	Command     string
	Args        []string
	WorkingDir  string
	Environment map[string]string
}

type ConfigManager struct {
	Configs       []*RunConfig
	SelectedIndex int
	ProjectRoot   string
	ProjectType   ProjectType
}

func NewConfigManager(rootPath string) *ConfigManager {
	cm := &ConfigManager{
		Configs:       make([]*RunConfig, 0),
		SelectedIndex: 0,
		ProjectRoot:   rootPath,
	}

	cm.ProjectType = DetectProjectType(rootPath)
	cm.LoadConfigs(rootPath)

	return cm
}

func (cm *ConfigManager) LoadConfigs(rootPath string) []*RunConfig {
	configs := cm.loadFromConfigFile(rootPath)
	if len(configs) > 0 {
		cm.Configs = configs
		return configs
	}

	configs = cm.generateDefaults()
	cm.Configs = configs
	return configs
}

func (cm *ConfigManager) loadFromConfigFile(rootPath string) []*RunConfig {
	configPath := filepath.Join(rootPath, ".tron", "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil
	}

	return nil
}

func (cm *ConfigManager) generateDefaults() []*RunConfig {
	defaults, ok := DefaultConfigsByType[cm.ProjectType]
	if !ok {
		return []*RunConfig{}
	}

	configs := make([]*RunConfig, 0, len(defaults))
	for _, d := range defaults {
		config := &RunConfig{
			Name:        d.Name,
			Command:     d.Command,
			Args:        d.Args,
			WorkingDir:  cm.ProjectRoot,
			Environment: make(map[string]string),
		}
		configs = append(configs, config)
	}

	return configs
}

func (cm *ConfigManager) GetDefault() *RunConfig {
	if len(cm.Configs) == 0 {
		return nil
	}
	return cm.Configs[cm.SelectedIndex]
}

func (cm *ConfigManager) Add(name, command string, args ...string) *RunConfig {
	config := &RunConfig{
		Name:        name,
		Command:     command,
		Args:        args,
		WorkingDir:  cm.ProjectRoot,
		Environment: make(map[string]string),
	}
	cm.Configs = append(cm.Configs, config)
	return config
}

func (cm *ConfigManager) Select(index int) {
	if index >= 0 && index < len(cm.Configs) {
		cm.SelectedIndex = index
	}
}

func (cm *ConfigManager) GetSelected() *RunConfig {
	if cm.SelectedIndex >= 0 && cm.SelectedIndex < len(cm.Configs) {
		return cm.Configs[cm.SelectedIndex]
	}
	return nil
}

func (cm *ConfigManager) Remove(index int) {
	if index < 0 || index >= len(cm.Configs) {
		return
	}
	cm.Configs = append(cm.Configs[:index], cm.Configs[index+1:]...)
	if cm.SelectedIndex >= len(cm.Configs) {
		cm.SelectedIndex = len(cm.Configs) - 1
	}
	if cm.SelectedIndex < 0 {
		cm.SelectedIndex = 0
	}
}

func (cm *ConfigManager) Update(index int, name, command string, args ...string) {
	if index < 0 || index >= len(cm.Configs) {
		return
	}
	cm.Configs[index].Name = name
	cm.Configs[index].Command = command
	cm.Configs[index].Args = args
}
