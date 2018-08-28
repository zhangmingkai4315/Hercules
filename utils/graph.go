package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// PrometheusNode will include current node infomations
type PrometheusNode struct {
	Children         PrometheusNodeList `json:"children"`
	AgentStatus      bool               `json:"agent_status"`
	PrometheusStatus bool               `json:"prometheus_status"`
	PrometheusHost   string             `json:"prometheus_host"`
	AgentHost        string             `json:"agent_host"`
}

// NewPrometheusNode create a new node with children nodes
func NewPrometheusNode(prometheusHost string) (*PrometheusNode, error) {
	if prometheusHost == "" {
		return nil, errors.New("host must not be empty")
	}
	return &PrometheusNode{
		Children:         nil,
		AgentStatus:      false,
		PrometheusStatus: false,
		PrometheusHost:   prometheusHost,
		AgentHost:        "",
	}, nil
}

// PrometheusNodeList list of many prometheus node
type PrometheusNodeList []*PrometheusNode

// NewPrometheusNodeList create nodes list for management
func NewPrometheusNodeList(feds []string) PrometheusNodeList {
	var pnl PrometheusNodeList
	for _, fed := range feds {
		if pn, err := NewPrometheusNode(fed); err == nil {
			pnl = append(pnl, pn)
		}
	}
	return pnl
}

// Search will find the matched nodes and  return true if exist
func (pn *PrometheusNode) Search(host string, recursive bool) bool {
	for _, children := range pn.Children {
		if children.PrometheusHost == host {
			return true
		}
		if recursive == true {
			find := children.Search(host, recursive)
			if find == true {
				return true
			}
		}
	}
	return false
}

// SearchAndUpdateAgentStatus will find the matched nodes and update agent status
func (pn *PrometheusNode) SearchAndUpdateAgentStatus(host string, recursive bool, status bool) bool {
	for _, children := range pn.Children {
		if children.PrometheusHost == host {
			children.AgentStatus = status
			return true
		}
		if recursive == true {
			children.SearchAndUpdateAgentStatus(host, true, status)
		}
	}
	return false
}

// SearchAndUpdatePrometheusStatus will find the matched nodes and update prometheus status
func (pn *PrometheusNode) SearchAndUpdatePrometheusStatus(host string, recursive bool, status bool) bool {
	for _, children := range pn.Children {
		if children.PrometheusHost == host {
			children.PrometheusStatus = status
			return true
		}
		if recursive == true {
			children.SearchAndUpdatePrometheusStatus(host, true, status)
		}
	}
	return false
}

// InsertOrUpdate will insert new node if not exist or update exist
// if search exist in its tree
func (pn *PrometheusNode) InsertOrUpdate(newNode *PrometheusNode, search bool) {
	// search only in first layer
	// if not exist , then append on its children array
	if search && pn.Search(newNode.PrometheusHost, false) == false {
		pn.Children = append(pn.Children, newNode)
		return
	}
	for _, child := range pn.Children {
		if child.PrometheusHost == newNode.PrometheusHost {
			child.Children = newNode.Children
			return
		}
		child.InsertOrUpdate(newNode, false)
	}
}

// DeleteNodeByHost will delete from graph by host name
func (pn *PrometheusNode) DeleteNodeByHost(host string) bool {
	for index, children := range pn.Children {
		if children.PrometheusHost == host {
			pn.Children = append(pn.Children[:index], pn.Children[index+1:]...)
			return true
		}
		deleteStatus := children.DeleteNodeByHost(host)
		if deleteStatus == true {
			return true
		}
	}
	return false
}

// PrintNodesTree print out the struct of nodes
func (pn *PrometheusNode) PrintNodesTree(prefix string, depth int, withStatus bool) string {
	prefixWithDepth := strings.Repeat(prefix, depth)
	status := ""
	if withStatus == true {
		if pn.AgentStatus == true {
			status = "[agent=ok]"
		} else {
			status = "[agent=error]"
		}
	} else {
	}
	tree := "\n" + prefixWithDepth + pn.PrometheusHost + status
	for _, child := range pn.Children {
		tree = tree + fmt.Sprint(child.PrintNodesTree(prefix, depth+1, withStatus))
	}
	return tree
}

// Ping will send request to each children and get back the response
func (pn *PrometheusNode) Ping() {
	// todo : update status only ping the first level of children
	// check current prometheus host status
	// MakeRequest(pn.PrometheusHost)
}

// GetGraph will return a http handler function
// which encode current node info to json response
func GetGraph(pn *PrometheusNode) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(pn)
	}
}

// UpdateGraph will insert or update a node from post data
func UpdateGraph(pn *PrometheusNode) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var node PrometheusNode
		if err := json.NewDecoder(r.Body).Decode(&node); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Invalid request"))
			return
		}
		if node.PrometheusHost == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Invalid request"))
			return
		}
		pn.InsertOrUpdate(&node, true)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(pn)
	}
}
