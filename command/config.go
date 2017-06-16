package command

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"

	yaml "gopkg.in/yaml.v2"
)

//Config holds the configuration of maya
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

//getConfig to load the config from a yaml file
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

type validatable interface {
	validate() (bool, []error)
}

// validate will vaidate the config file and
// returns boolean value and err based on the errors
func (c *Config) validate() (bool, []error) {
	v := newValidator()

	if c.Version != 0.3 {
		v.addError(errors.New("Config Version cannot be empty"))
	}

	if c.Spec.Provider != "nomad" && c.Spec.Provider != "kubernetes" {
		v.addError(errors.New("Invalid orchestrator provider"))

	}

	if c.Spec.Bin[0].Version == "" {

		v.addError(errors.New("Orchestrator Version cannot be empty"))

	}

	if c.Metadeta[0].Role == "" && c.Metadeta[1].Role == "" {

		v.addError(errors.New("Please specify the Role of Nodes"))
	}

	if c.Args[0].Name == "" && c.Args[1].Name == "" {
		v.addError(errors.New("Hostname cannot be empty"))
	}

	if c.Args[0].Addr == "" && c.Args[1].Addr == "" {
		v.addError(errors.New("Invalid IP address"))
	}

	return v.valid()
}

type validator struct {
	errs []error
}

func newValidator() *validator {
	return &validator{
		errs: []error{},
	}
}

func (v *validator) addError(err ...error) {
	v.errs = append(v.errs, err...)
}

func (v *validator) valid() (bool, []error) {
	if len(v.errs) > 0 {
		return false, v.errs
	}
	return true, nil
}

// PrintValidationErrors loops through the errors
func PrintValidationErrors(errors []error) {
	for _, err := range errors {
		//	PrintColor(out, Red, "- %v\n", err)
		fmt.Println("-", err)
	}
}

// PrintColor prints text in color
//func PrintColor(out io.Writer, clr *color.Color, msg string, a ...interface{}) {
// Remove any newline, results in only one \n
//	line := fmt.Sprintf("%s", clr.SprintfFunc()(msg, a...))
//	fmt.Fprint(out, line)
//}

//func validateConfig(out io.Writer, config *Config) error {
//	var c *Config
//	ok, errs := c.validate()
//	if !ok {
//		PrintValidationErrors(out, errs)
//		return fmt.Errorf("file validation error prevents installation from proceeding")
//	}
//	return nil
//}

/* func getEnvOrDefault(name string, defaultValue string) string {
	v := os.Getenv(name)
	if v == "" {
		v = defaultValue
	}
	return v
}*/

/*
func getEnvOrDefault(env string) string {
	if env == "" {
		host, _ := os.Hostname()
		addrs, _ := net.LookupIP(host)
		for _, addr := range addrs {
			if ipv4 := addr.To4(); ipv4 != nil {
				env = ipv4.String()
				if env == "127.0.0.1" {
					continue
				}
				break
			}
		}
	}
	return "http://" + env + ":5656"
}*/
