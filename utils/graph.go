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
	Children PrometheusNodeList `json:"children"`
	Status   bool               `json:"status"`
	Host     string             `json:"host"`
}

// NewPrometheusNode create a new node with children nodes
func NewPrometheusNode(hostWithPort string) (*PrometheusNode, error) {
	if hostWithPort == "" {
		return nil, errors.New("host must not be empty")
	}
	return &PrometheusNode{
		Children: nil,
		Status:   false,
		Host:     hostWithPort,
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
		if children.Host == host {
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

// SearchAndUpdateStatus will find the matched nodes and update status
func (pn *PrometheusNode) SearchAndUpdateStatus(host string, recursive bool, status bool) bool {
	for _, children := range pn.Children {
		if children.Host == host {
			children.Status = status
			return true
		}
		if recursive == true {
			children.SearchAndUpdateStatus(host, recursive, status)
		}
	}
	return false
}

// InsertOrUpdate will insert new node if not exist or update exist
// if search exist in its tree
func (pn *PrometheusNode) InsertOrUpdate(newNode *PrometheusNode, search bool) {
	// search only in first layer
	// if not exist , then append on its children array
	if search && pn.Search(newNode.Host, false) == false {
		pn.Children = append(pn.Children, newNode)
		return
	}
	for _, child := range pn.Children {
		if child.Host == newNode.Host {
			child.Children = newNode.Children
			return
		}
		child.InsertOrUpdate(newNode, false)
	}
}

// DeleteNodeByHost will delete from graph by host name
func (pn *PrometheusNode) DeleteNodeByHost(host string) bool {
	for index, children := range pn.Children {
		if children.Host == host {
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
		if pn.Status == true {
			status = "[ok]"
		} else {
			status = "[error]"
		}
	} else {
	}
	tree := "\n" + prefixWithDepth + pn.Host + status
	for _, child := range pn.Children {
		tree = tree + fmt.Sprint(child.PrintNodesTree(prefix, depth+1, withStatus))
	}
	return tree
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
		if node.Host == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Invalid request"))
			return
		}
		pn.InsertOrUpdate(&node, true)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(pn)
	}
}
