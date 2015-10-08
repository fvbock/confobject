package main

import (
	"log"

	yaml "gopkg.in/yaml.v2"
)

func main() {
	v := [][]string{
		[]string{"StringSetting", "foo"},
		[]string{"IntSetting", "23"},
		[]string{"FloatSetting", "42.0"},
		[]string{"BoolSetting", "true"},
		[]string{"SliceConfig.StringSliceSetting", "foo", "bar", "baz"},
		[]string{"SliceConfig.IntSliceSetting", "1", "2", "3"},
		[]string{"SliceConfig.FloatSliceSetting", "1.2", "2.3", "3.4"},
	}
	out, err := yaml.Marshal(v)
	if err != nil {
		log.Println("YAML marshall ERROR:", err)
		return
	}
	log.Println(string(out))
}
