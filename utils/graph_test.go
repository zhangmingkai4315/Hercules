package utils

import (
	"testing"
)

var (
	rootHost      = "source-prometheus:9090"
	childrenHosts = []string{
		"source-prometheus-1:9090",
		"source-prometheus-11:9090",
		"source-prometheus-2:9090",
	}
	fedsRoot = []string{
		"source-prometheus-1:9090",
		"source-prometheus-2:9090",
		"source-prometheus-3:9090",
	}
	fedsChild11 = []string{
		"source-prometheus-11:9090",
		"source-prometheus-12:9090",
		"source-prometheus-13:9090",
	}
	fedsChild21 = []string{
		"source-prometheus-21:9090",
		"source-prometheus-22:9090",
		"source-prometheus-23:9090",
	}
	fedsChild111 = []string{
		"source-prometheus-111:9090",
		"source-prometheus-112:9090",
		"source-prometheus-113:9090",
	}
	fedsBadWithEmptyList = []string{
		"source-prometheus-1:9090",
		"",
		"source-prometheus-3:9090",
	}
)

func TestNewPrometheusNode(t *testing.T) {
	host := rootHost
	node, err := NewPrometheusNode(host)
	if err != nil {
		t.Fatalf("Exprect create a new node struct , but got error %s", err)
	}
	if node.Host != host {
		t.Fatalf("Exprect node host is %s , but got  %s", host, node.Host)
	}
	host = ""
	node, err = NewPrometheusNode(host)
	if err == nil {
		t.Fatal("Expect err is not nil, but got nil")
	}
}

func TestNewPrometheusNodeList(t *testing.T) {
	pnl := NewPrometheusNodeList(fedsRoot)
	if len(pnl) != len(fedsRoot) {
		t.Fatalf("Exprect list contain %d node, but got  %d", len(fedsRoot), len(pnl))
	}
	pnl = NewPrometheusNodeList(fedsBadWithEmptyList)
	if len(pnl) != 2 {
		t.Fatalf("Exprect list contain %d node, but got  %d", 2, len(pnl))
	}
}

func TestPrometheusNodeListSearch(t *testing.T) {
	nodeRoot, _ := NewPrometheusNode(rootHost)
	nodeRootList := NewPrometheusNodeList(fedsRoot)
	nodeRootList[0].Children = NewPrometheusNodeList(fedsChild11)
	nodeRootList[0].Children[0].Children = NewPrometheusNodeList(fedsChild111)
	nodeRootList[1].Children = NewPrometheusNodeList(fedsChild21)
	nodeRoot.Children = nodeRootList
	if searchResult := nodeRoot.Search(fedsChild11[0], true); searchResult == false {
		t.Fatalf("Expect search result is true, but got false")
	}

	if searchResult := nodeRoot.Search(fedsChild111[0], false); searchResult == true {
		t.Fatalf("Expect search result is false, but got true")
	}
	if searchResult := nodeRoot.Search(fedsChild11[0], true); searchResult == false {
		t.Fatalf("Expect search result is true, but got false")
	}
	if searchResult := nodeRoot.Search("not-exist.domain", true); searchResult == true {
		t.Fatalf("Expect search result is false, but got true")
	}
}

func TestPrometheusNodeInsertOrUpdate(t *testing.T) {
	nodeRoot, _ := NewPrometheusNode(rootHost)
	nodeRootList := NewPrometheusNodeList(fedsRoot)
	nodeRoot.Children = nodeRootList

	nodeChild1, _ := NewPrometheusNode(childrenHosts[0])
	nodeChild1List := NewPrometheusNodeList(fedsChild11)
	nodeChild1.Children = nodeChild1List

	nodeRootList[0].Children = NewPrometheusNodeList(fedsChild11)
	nodeRootList[0].Children[0].Children = NewPrometheusNodeList(fedsChild111)

	nodeChild2, _ := NewPrometheusNode(childrenHosts[2])
	nodeChild2List := NewPrometheusNodeList(fedsChild21)
	nodeChild2.Children = nodeChild2List

	nodeRoot.InsertOrUpdate(nodeChild2, true)

	if exist := nodeRoot.Search(fedsChild21[0], true); exist == false {
		t.Fatalf("Expect nodeRoot contain new element, but got nothing")
	}

	rootChildrenNumber := len(nodeRoot.Children)
	newChildWithNoFeds, _ := NewPrometheusNode("new-test.domain")
	nodeRoot.InsertOrUpdate(newChildWithNoFeds, true)
	if len(nodeRoot.Children) != (rootChildrenNumber + 1) {
		t.Fatal("Expect noderoot add a new element, but got no change")
	}
}

func TestPrometheusNodeSearchAndUpdateStatus(t *testing.T) {
	nodeRoot, _ := NewPrometheusNode(rootHost)
	nodeRootList := NewPrometheusNodeList(fedsRoot)
	nodeRoot.Children = nodeRootList

	nodeRoot.SearchAndUpdateStatus(childrenHosts[0], true, true)

	if nodeRoot.Children[0].Status != true {
		t.Fatalf("Expect change status to true, but got %v", nodeRoot.Children[0].Status)
	}

	nodeChild2, _ := NewPrometheusNode(childrenHosts[2])
	nodeChild2List := NewPrometheusNodeList(fedsChild21)
	nodeChild2.Children = nodeChild2List
	nodeRoot.InsertOrUpdate(nodeChild2, true)

	nodeRoot.SearchAndUpdateStatus(childrenHosts[2], true, true)

	// status := nodeRoot.Children[1].Children

}

func TestPrometheusNodePrintTree(t *testing.T) {
	nodeRoot, _ := NewPrometheusNode(rootHost)
	nodeRootList := NewPrometheusNodeList(fedsRoot)
	nodeRoot.Children = nodeRootList
	nodeChild2, _ := NewPrometheusNode(childrenHosts[2])
	nodeChild2List := NewPrometheusNodeList(fedsChild21)
	nodeChild2.Children = nodeChild2List
	nodeRoot.InsertOrUpdate(nodeChild2, true)
	expect := `
source-prometheus:9090
————source-prometheus-1:9090
————source-prometheus-2:9090
————————source-prometheus-21:9090
————————source-prometheus-22:9090
————————source-prometheus-23:9090
————source-prometheus-3:9090`

	if expect != nodeRoot.PrintNodesTree("————", 0, false) {
		t.Fatalf("Expect %s, but got %s", expect, nodeRoot.PrintNodesTree("————", 0, false))
	}
	expect = `
source-prometheus:9090[error]
————source-prometheus-1:9090[error]
————source-prometheus-2:9090[error]
————————source-prometheus-21:9090[error]
————————source-prometheus-22:9090[error]
————————source-prometheus-23:9090[error]
————source-prometheus-3:9090[error]`
	if expect != nodeRoot.PrintNodesTree("————", 0, true) {
		t.Fatalf("Expect %s, but got %s", expect, nodeRoot.PrintNodesTree("————", 0, true))
	}
	nodeRoot.SearchAndUpdateStatus(childrenHosts[2], true, true)
	expect = `
source-prometheus:9090[error]
————source-prometheus-1:9090[error]
————source-prometheus-2:9090[ok]
————————source-prometheus-21:9090[error]
————————source-prometheus-22:9090[error]
————————source-prometheus-23:9090[error]
————source-prometheus-3:9090[error]`
	if expect != nodeRoot.PrintNodesTree("————", 0, true) {
		t.Fatalf("Expect %s, but got %s", expect, nodeRoot.PrintNodesTree("————", 0, true))
	}
}

func TestPrometheusNodeDeleteByHost(t *testing.T) {
	nodeRoot, _ := NewPrometheusNode(rootHost)
	nodeRootList := NewPrometheusNodeList(fedsRoot)
	nodeRoot.Children = nodeRootList
	nodeChild2, _ := NewPrometheusNode(childrenHosts[2])
	nodeChild2List := NewPrometheusNodeList(fedsChild21)
	nodeChild2.Children = nodeChild2List
	nodeRoot.InsertOrUpdate(nodeChild2, true)
	deleteItem := "source-prometheus-21:9090"
	if true != nodeRoot.Search(deleteItem, true) {
		t.Fatalf("Expect %s found in list, but not in list", deleteItem)
	}
	deleteStatus := nodeRoot.DeleteNodeByHost(deleteItem)
	if deleteStatus == false {
		t.Fatal("Expect delete and return true, but got false")
	}
	if false != nodeRoot.Search(deleteItem, true) {
		t.Fatalf("Expect %s not found, but still in list", deleteItem)
	}
}
