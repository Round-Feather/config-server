package utils

import (
	"bytes"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

type Configuration struct {
	Path string
}

func (c Configuration) GetProperties(properties map[string]string) {
	data, _ := os.ReadFile(c.Path)
	reader := bytes.NewReader(data)
	decoder := yaml.NewDecoder(reader)

	var y map[string]interface{}

	for decoder.Decode(&y) == nil {
		readObject([]string{}, y, properties)
	}
}

func readObject(keyParts []string, object interface{}, propertiesMap map[string]string) {
	mapObject, isMap := object.(map[string]interface{})
	if isMap {
		for k, v := range mapObject {
			readObject(append(keyParts, fmt.Sprintf("%v", k)), v, propertiesMap)
		}
	}
	genericMapObject, isGenericMap := object.(map[interface{}]interface{})
	if isMap {
		for k, v := range genericMapObject {
			readObject(append(keyParts, fmt.Sprintf("%v", k)), v, propertiesMap)
		}
	}

	listObject, isSlice := object.([]interface{})
	if isSlice {
		stringSlice := []string{}
		for _, v := range listObject {
			stringSlice = append(stringSlice, fmt.Sprintf("%v", v))
		}
		key := strings.Join(keyParts, ".")
		val := strings.Join(stringSlice, ",")

		propertiesMap[key] = val
	}

	if !isMap && !isGenericMap && !isSlice {
		key := strings.Join(keyParts, ".")
		val := fmt.Sprintf("%v", object)

		propertiesMap[key] = val
	}
}
