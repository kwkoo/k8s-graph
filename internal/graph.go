package internal

import (
	"bytes"
	"encoding/json"
	"log"
)

type Node struct {
	Uid  string `json:"id"`
	Kind string `json:"kind"`
	Name string `json:"name"`
}

type Link struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

type Graph struct {
	nodeMap     map[string]*Node    // map of uid to node
	linkMap     map[string]struct{} // key is in the form source:target
	linkSources map[string]struct{} // key is the source uid
	linkTargets map[string]struct{} // key is the target uid
	Nodes       []*Node             `json:"nodes"`
	Links       []Link              `json:"links"`
}

func InitGraph() *Graph {
	graph := Graph{
		nodeMap:     make(map[string]*Node),
		linkMap:     map[string]struct{}{},
		linkSources: map[string]struct{}{},
		linkTargets: map[string]struct{}{},
		Nodes:       []*Node{},
		Links:       []Link{},
	}

	return &graph
}

func (g *Graph) addNode(uid, kind, name string) {
	n := Node{
		Uid:  uid,
		Kind: kind,
		Name: name,
	}
	g.nodeMap[uid] = &n
	g.Nodes = append(g.Nodes, &n)
}

func (g *Graph) nodeExists(uid string) bool {
	_, ok := g.nodeMap[uid]
	return ok
}

func (g *Graph) addLink(source, target string) {
	if g.linkExists(source, target) {
		return
	}
	l := Link{
		Source: source,
		Target: target,
	}
	g.Links = append(g.Links, l)
	g.linkMap[source+":"+target] = struct{}{}
	g.linkSources[source] = struct{}{}
	g.linkTargets[target] = struct{}{}
}

func (g *Graph) cleanLinks() {
	cleaned := []Link{}

	for _, link := range g.Links {
		if !g.nodeExists(link.Source) || !g.nodeExists(link.Target) {
			delete(g.linkMap, link.Source+":"+link.Target)
			delete(g.linkSources, link.Source)
			delete(g.linkTargets, link.Target)
			continue
		}
		cleaned = append(cleaned, link)
	}
	g.Links = cleaned
}

// Cleans out ConfigMaps and Secrets that are not linked to anything else in
// order to avoid cluttering the graph
func (g *Graph) cleanNodes() {
	cleaned := []*Node{}

	for _, node := range g.Nodes {
		if node.Kind == "cm" || node.Kind == "secret" {
			_, isSource := g.linkSources[node.Uid]
			if !isSource {
				_, isTarget := g.linkTargets[node.Uid]
				if !isTarget {
					// this node is a cm or secret and is not linked to
					// anything else
					delete(g.nodeMap, node.Uid)
					continue
				}
			}
		}
		cleaned = append(cleaned, node)
	}
	g.Nodes = cleaned
}

func (g *Graph) linkExists(source, target string) bool {
	_, ok := g.linkMap[source+":"+target]
	return ok
}

func (g *Graph) findResource(kind, name string) string {
	for _, node := range g.Nodes {
		if node.Kind == kind && node.Name == name {
			return node.Uid
		}
	}

	return ""
}

func (g Graph) String() string {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	if err := enc.Encode(g); err != nil {
		log.Fatal(err)
	}
	return b.String()
}
