package confreader

import (
	"fmt"
	"io/ioutil"
	"path"
	"runtime"

	yaml "gopkg.in/yaml.v2"
)

// ReadConfig - parse yaml config
func ReadConfig(fName string, conf interface{}) error {
	_, caller, _, ok := runtime.Caller(0)
	if !ok {
		panic("No caller information")
	}
	fPath := path.Dir(caller) + "/../etc/" + fName + "conf.yaml"

	data, err := ioutil.ReadFile(fPath)
	if err != nil {
		return fmt.Errorf("can't read yaml file %q: %s", fName, err)
	}

	if err := yaml.Unmarshal(data, conf); err != nil {
		return fmt.Errorf("can't write YAML data into file %q: %s", fName, err)
	}

	return nil
}
