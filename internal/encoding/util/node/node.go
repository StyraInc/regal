// Package node provides utilities for creating and caching YAML nodes.
// Contrary to the JSON utilities, these are less about convenience and
// more about trying to do something about the rather poor performance of
// the YAML library used.
package node

import (
	"go.yaml.in/yaml/v3"

	"github.com/open-policy-agent/opa/v1/util"

	"github.com/open-policy-agent/regal/pkg/roast/util/concurrent"
)

var (
	keys   = [...]*yaml.Node{String("level"), String("error"), String("warn"), String("ignore")}
	levels = map[string][]*yaml.Node{
		"error":  {keys[0], keys[1]},
		"warn":   {keys[0], keys[2]},
		"ignore": {keys[0], keys[3]},
	}
	levelMappings = map[string]*yaml.Node{
		"error":  Mapping(levels["error"]...),
		"warn":   Mapping(levels["warn"]...),
		"ignore": Mapping(levels["ignore"]...),
	}
	stringCache = concurrent.MapOf(make(map[string]*yaml.Node))
)

func Level(level string) []*yaml.Node {
	if node, ok := levels[level]; ok {
		return node
	}

	return []*yaml.Node{keys[0], String(level)}
}

func LevelMapping(level string) *yaml.Node {
	if node, ok := levelMappings[level]; ok {
		return node
	}

	return &yaml.Node{Kind: yaml.MappingNode, Content: []*yaml.Node{keys[0], String(level)}}
}

// String creates and returns cached YAML scalar node for the given string.
// The cache grows infinitely, which works in the contect of parsing Regal
// configuration files. Don't use anywhere else (or make it more generic).
func String(name string) *yaml.Node {
	if node, ok := stringCache.Get(name); ok {
		return node
	}

	node := &yaml.Node{Kind: yaml.ScalarNode, Value: name}
	stringCache.Set(name, node)

	return node
}

func ArrayOfString(values []string) *yaml.Node {
	return &yaml.Node{Kind: yaml.SequenceNode, Content: util.Map(values, String)}
}

func Mapping(content ...*yaml.Node) *yaml.Node {
	return &yaml.Node{Kind: yaml.MappingNode, Content: content}
}
