package config

type Config struct {
	AutoCfg *AutoConfig
}

type AutoConfig struct {
	Port         int              `mapstructure:"port"`
	TraceApi     string           `mapstructure:"trace_api"`
	ScanInterval int64            `mapstructure:"scan_interval"`
	BlackList    *BlackListConfig `mapstructure:"black_list"`
}

type BlackListConfig struct {
	GoList   map[string][]string `mapstructure:"go"`
	JavaList map[string][]string `mapstructure:"java"`
}

func (cfg *BlackListConfig) GetGoList() []string {
	list := make([]string, 0)
	for _, value := range cfg.GoList {
		list = append(list, value...)
	}
	return list
}

func (cfg *BlackListConfig) GetJavaList() []string {
	list := make([]string, 0)
	for _, value := range cfg.JavaList {
		list = append(list, value...)
	}
	return list
}
