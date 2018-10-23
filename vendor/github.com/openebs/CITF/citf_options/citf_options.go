package citfoptions

// CreateOptions specifies which fields of CITF should be included when created or reloadedd
type CreateOptions struct {
	ConfigPath         string
	EnvironmentInclude bool
	K8SInclude         bool
	DockerInclude      bool
	LoggerInclude      bool
}

// CreateOptionsIncludeAll returns CreateOptions where all fields are set to `true` and ConfigPath is set to configPath
func CreateOptionsIncludeAll(configPath string) *CreateOptions {
	var citfCreateOptions CreateOptions
	citfCreateOptions.ConfigPath = configPath
	citfCreateOptions.EnvironmentInclude = true
	citfCreateOptions.K8SInclude = true
	citfCreateOptions.DockerInclude = true
	citfCreateOptions.LoggerInclude = true
	return &citfCreateOptions
}

// CreateOptionsIncludeAllButEnvironment returns CreateOptions where all fields except `Environment` are set to `true` and ConfigPath is set to configPath
func CreateOptionsIncludeAllButEnvironment(configPath string) *CreateOptions {
	citfCreateOptions := CreateOptionsIncludeAll(configPath)

	citfCreateOptions.EnvironmentInclude = false
	return citfCreateOptions
}

// CreateOptionsIncludeAllButK8s returns CreateOptions where all fields except `K8S` are set to `true` and ConfigPath is set to configPath
func CreateOptionsIncludeAllButK8s(configPath string) *CreateOptions {
	citfCreateOptions := CreateOptionsIncludeAll(configPath)

	citfCreateOptions.K8SInclude = false
	return citfCreateOptions
}

// CreateOptionsIncludeAllButDocker returns CreateOptions where all fields except `Docker` are set to `true` and ConfigPath is set to configPath
func CreateOptionsIncludeAllButDocker(configPath string) *CreateOptions {
	citfCreateOptions := CreateOptionsIncludeAll(configPath)

	citfCreateOptions.DockerInclude = false
	return citfCreateOptions
}

// CreateOptionsIncludeAllButLogger returns CreateOptions where all fields except `Logger` are set to `true` and ConfigPath is set to configPath
func CreateOptionsIncludeAllButLogger(configPath string) *CreateOptions {
	citfCreateOptions := CreateOptionsIncludeAll(configPath)

	citfCreateOptions.LoggerInclude = false
	return citfCreateOptions
}
