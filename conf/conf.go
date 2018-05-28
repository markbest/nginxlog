package conf

import (
	"errors"
	"github.com/BurntSushi/toml"
	"io/ioutil"
)

var (
	Conf              config
	defaultConfigFile = "./conf/conf.toml"
)

type config struct {
	//App config
	App app `toml:"app"`

	//Elastic config
	Elastic elastic `toml:"elastic"`

	//Log config
	Log log `toml:"log"`

	//Redis config
	Redis redis `toml:"redis"`
}

type app struct {
	Port  string `toml:"port"`
	Pprof string `toml:"pprof"`
	Debug bool   `toml:"debug"`
}

type elastic struct {
	ElasticUrl         string `toml:"elastic_url"`
	ElasticIndex       string `toml:"elastic_index"`
	ElasticType        string `toml:"elastic_type"`
	ElasticLogPath     string `toml:"elastic_log_path"`
	ElasticLogMaxFiles int    `toml:"elastic_log_max_files"`
}

type log struct {
	TargetPath       string `toml:"target_path"`
	TargetFilePrefix string `toml:"target_file_prefix"`
}

type redis struct {
	Host     string `toml:"redis_host"`
	Password string `toml:"redis_password"`
	Port     string `toml:"redis_port"`
	Db       int    `toml:"redis_db"`
}

func InitConfig() (err error) {
	configBytes, err := ioutil.ReadFile(defaultConfigFile)
	if err != nil {
		return errors.New("config load err:" + err.Error())
	}
	_, err = toml.Decode(string(configBytes), &Conf)
	if err != nil {
		return errors.New("config decode err:" + err.Error())
	}
	return nil
}
