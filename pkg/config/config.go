package config

type ServerConfig struct {
	DatabaseConfig *DataBaseConfig `yaml:"databaseConfig" mapstructure:"databaseConfig"`
	Kubernetes     *Kubernetes     `yaml:"kubernetes" mapstructure:"kubernetes"`
	FileConfig     *FileConfig     `yaml:"fileConfig" mapstructure:"fileConfig"`
	LogFilePath    string          `yaml:"logFilePath" mapstructure:"logFilePath"`
	LogLevel       string          `yaml:"logLevel" mapstructure:"logLevel"`
	Port           uint32          `yaml:"port" mapstructure:"port"`
	BindIP         string          `yaml:"bindIP" mapstructure:"bindIP"`
}

type Kubernetes struct {
	//if null global else the config namespace
	WatchNamespace string `yaml:"watchNamespace" mapstructure:"watchNamespace"`
}

type FileConfig struct {
	FilePath string `yaml:"filePath" mapstructure:"filePath"`
}

type DataBaseConfig struct {
	Host     string `yaml:"host" mapstructure:"host"`
	Port     int    `yaml:"port" mapstructure:"port"`
	Type     string `yaml:"type" mapstructure:"type"`
	Name     string `yaml:"name" mapstructure:"name"`
	UserName string `yaml:"userName" mapstructure:"userName"`
	Password string `yaml:"password" mapstructure:"password"`
	OtherDSN string `yaml:"otherDSN" mapstructure:"otherDSN"`
}
