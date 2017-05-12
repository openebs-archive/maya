package command

import (
	"io/ioutil"
	"log"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	Version float64 `json:"version"`
	Kind    string  `json:"kind"`
	Spec    struct {
		Provider string `json:"provider"`
		Bin      []struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"bin"`
	} `json:"spec"`
	Metadeta []struct {
		Role string `json:"role"`
	} `json:"metadeta"`
	Args []struct {
		Name string `json:"name"`
		Addr string `json:"addr"`
	} `json:"args"`
}

func getConfig(path string) Config {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal("Config File Missing. ", err)
	}

	var config Config
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		log.Fatal("Config Parse Error: ", err)
	}
	return config
}
