package utils

import (
	"errors"
	"fmt"
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
func (pnl PrometheusNodeList) Search(host string, recursive bool) bool {
	for _, children := range pnl {
		if children.Host == host {
			return true
		}
		if recursive == true {
			find := children.Children.Search(host, recursive)
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
	if search && pn.Children.Search(newNode.Host, false) == false {
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

func (pn *PrometheusNode) PrintNodesTree(prefix string, depth int) string {
	prefixWithDepth := strings.Repeat(prefix, depth)
	tree := "\n" + prefixWithDepth + pn.Host
	for _, child := range pn.Children {
		tree = tree + fmt.Sprint(child.PrintNodesTree(prefix, depth+1))
	}
	return tree
}
