package vrctFs

import (
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"reflect"
	"sort"
)

func (p *FilePrototype) mergeLayers() (PrototypeLayer, error) {
	switch p.FileType {
	case TextFile:
		return p.mergeTextLayers()
	default:
		return p.mergeConfigLayers()
	}
}

func (p *FilePrototype) mergeTextLayers() (PrototypeLayer, error) {
	finalLayer := PrototypeLayer{
		IsOptional: false,
	}

	sort.Slice(p.Layers, func(i, j int) bool {
		return !p.Layers[i].IsOptional && p.Layers[j].IsOptional
	})

	for i, layer := range p.Layers {
		// TODO: check if contents differ
		if finalLayer.ContentPath != "" && layer.ContentPath != "" && !layer.IsOptional {
			return finalLayer, fmt.Errorf(
				"non optional layer nr. %d is in conflict with previous layers", i,
			)
		}
		if (layer.IsOptional && finalLayer.ContentPath == "") || !layer.IsOptional {
			finalLayer.ContentPath = layer.ContentPath
		}
	}

	return finalLayer, nil
}

func (p *FilePrototype) mergeConfigLayers() (PrototypeLayer, error) {
	// create new layer first
	finalLayer, err := p.CreateLayer(nil, nil, true)
	if err != nil {
		return finalLayer, err
	}

	finalLayerContent, err := GetBsonMap(finalLayer.ContentPath)
	if err != nil {
		return finalLayer, err
	}

	finalLayerOptions, err := GetBsonMap(finalLayer.OptionsPath)
	if err != nil {
		return finalLayer, err
	}

	sort.Slice(p.Layers, func(i, j int) bool {
		return !p.Layers[i].IsOptional && p.Layers[j].IsOptional
	})

	for _, layer := range p.Layers {
		currentLayerContent, err := GetBsonMap(layer.ContentPath)
		if err != nil {
			return finalLayer, err
		}

		currentLayerOptions, err := GetBsonMap(layer.OptionsPath)
		if err != nil {
			return finalLayer, err
		}

		finalLayerContent, finalLayerOptions, err = mergeConfigs(finalLayerContent, finalLayerOptions, currentLayerContent, currentLayerOptions, layer.IsOptional)
		if err != nil {
			return finalLayer, err
		}
	}

	marshalledContent, err := bson.Marshal(finalLayerContent)
	if err != nil {
		return finalLayer, err
	}
	err = finalLayer.SetContent(marshalledContent)

	err = SaveBsonMap(finalLayerOptions, finalLayer.OptionsPath)

	return finalLayer, err
}

func mergeConfigs(merger map[string]any, mergerOptions map[string]any, toMerge map[string]any, toMergeOptions map[string]any, isOptional bool) (map[string]any, map[string]any, error) {
	for key, toMergeVal := range toMerge {
		mergerValue, ok := merger[key]

		mergerOption, mergerOptionOk := mergerOptions[key]
		toMergeOption, toMergeOptionOk := toMergeOptions[key]

		if !ok {
			merger[key] = toMergeVal
			if toMergeOptionOk {
				mergerOptions[key] = toMergeOption
			} else {
				mergerOptions[key] = isOptional
			}

			mergerValueType := reflect.ValueOf(merger[key]).Kind()
			mergerOptionType := reflect.ValueOf(mergerOptions[key]).Kind()

			if mergerValueType == reflect.Map && mergerOptionType == reflect.Array || mergerValueType == reflect.Bool && mergerOptionType != reflect.Bool {
				return merger, mergerOptions, fmt.Errorf("merger key and option key are of different types: '%s' and '%s'", mergerValueType, mergerOptionType)
			}
			continue
		}

		if reflect.ValueOf(toMergeVal).Kind() == reflect.Map {
			mergerMapValue, mergerOptionsMapValue, err := getBoolMap(mergerValue, mergerOption)
			if err != nil {
				return merger, mergerOptions, err
			}

			toMergeMapValue, toMergeOptionsMapValue, err := getBoolMap(mergerValue, mergerOption)
			if err != nil {
				return merger, mergerOptions, err
			}

			merger[key], mergerOptions[key], err = mergeConfigs(toMergeMapValue, mergerOptionsMapValue, mergerMapValue, toMergeOptionsMapValue, isOptional)
			if err != nil {
				return merger, mergerOptions, err
			}
			continue
		}

		if reflect.ValueOf(mergerValue).Kind() == reflect.Slice {
			return merger, mergerOptions, fmt.Errorf("key '%s' is unmergable as arrays merging is unsupported yet", key)
		}

		isMergerKeyOptional := true
		mergerOptionKind := reflect.ValueOf(mergerOption).Kind()
		if mergerOptionOk {
			if mergerOptionKind != reflect.Bool {
				return merger, mergerOptions, fmt.Errorf("passed key '%s' in merger options is not of type bool ('%s' passed)", key, mergerOptionKind)
			}
			isMergerKeyOptional = mergerOption.(bool)
		}

		isToMergeKeyOptional := isOptional
		toMergeOptionKind := reflect.ValueOf(toMergeOption).Kind()
		if toMergeOptionOk {
			if toMergeOptionKind != reflect.Bool {
				return merger, mergerOptions, fmt.Errorf("passed key '%s' in to merge options is not of type bool ('%s' passed)", key, toMergeOptionKind)
			}
			isToMergeKeyOptional = toMergeOption.(bool)
		}

		if isMergerKeyOptional && isToMergeKeyOptional {
			merger[key] = toMergeVal
			mergerOptions[key] = true
		} else if isMergerKeyOptional {
			merger[key] = toMergeVal
			mergerOptions[key] = false
		} else if isToMergeKeyOptional {
			mergerOptions[key] = false
		} else if mergerValue != toMergeVal {
			return merger, toMergeOptions, fmt.Errorf("passed key '%s' is unmergable (both merger and to merge are required)", key)
		}
	}

	return merger, mergerOptions, nil
}
