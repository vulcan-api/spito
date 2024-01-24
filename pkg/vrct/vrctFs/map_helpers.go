package vrctFs

import (
	"encoding/json"
	"fmt"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/yaml.v3"
	"os"
	"reflect"
)

func GetBsonMap(pathToFile string) (map[string]interface{}, error) {
	file, err := os.ReadFile(pathToFile)
	if err != nil {
		return nil, err
	}

	var bsonMap map[string]interface{}
	err = bson.Unmarshal(file, &bsonMap)
	if err != nil {
		return bsonMap, err
	}

	return bsonMap, nil
}

func SaveBsonMap(toSave map[string]interface{}, pathToFile string) error {
	content, err := bson.Marshal(toSave)
	if err != nil {
		return err
	}

	err = os.WriteFile(pathToFile, content, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func GetMapFromBytes(content []byte, configType FileType) (map[string]interface{}, error) {
	var err error
	var resultMap map[string]interface{}
	switch configType {
	case TextFile:
		break
	case JsonConfig:
		if content == nil {
			content = []byte("{}")
		}
		err = json.Unmarshal(content, &resultMap)
		break
	case YamlConfig:
		err = yaml.Unmarshal(content, &resultMap)
		break
	case TomlConfig:
		err = toml.Unmarshal(content, &resultMap)
		break
	default:
		return resultMap, fmt.Errorf("unsupported config type (FileType argument), passed '%d'", configType)
	}

	if err != nil {
		return resultMap, fmt.Errorf("could not obtain map from given array of bytes: %s", err)
	}

	if len(resultMap) == 0 {
		resultMap = make(map[string]interface{})
	}

	return resultMap, err
}
func getBoolMap(value interface{}, option interface{}) (map[string]interface{}, map[string]interface{}, error) {
	valueKind := reflect.ValueOf(value).Kind()
	optionKind := reflect.ValueOf(option).Kind()

	if valueKind != reflect.Map {
		return nil, nil, fmt.Errorf("trying to map interface that is not map")
	}

	mappedValue := value.(map[string]interface{})

	var mappedOption map[string]interface{}
	if optionKind == reflect.Map {
		mappedOption = option.(map[string]interface{})
	} else if option == reflect.Bool {
		for key := range mappedValue {
			mappedOption[key] = option.(bool)
		}
	} else {
		return nil, nil, fmt.Errorf("types conflict")
	}

	return mappedValue, mappedOption, nil
}
