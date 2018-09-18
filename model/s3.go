package model

import (
	"strings"
	"time"
)

type S3ObjectType int

const (
	Bucket S3ObjectType = iota //0
	Dir
	PreDir
	Object
)

type S3Object struct {
	ObjType S3ObjectType
	Name    string
	Date    *time.Time
	Size    *int64
}

func NewS3Object(objType S3ObjectType, name string, date *time.Time, size *int64) *S3Object {
	return &S3Object{
		ObjType: objType,
		Name:    name,
		Date:    date,
		Size:    size,
	}
}

type Node struct {
	Key      string
	Parent   *Node
	children map[string]*Node
	Objects  []*S3Object
	Position int
}

func NewNode(key string, parent *Node, objects []*S3Object) *Node {
	node := &Node{
		Key:      key,
		Parent:   parent,
		Objects:  objects,
		children: map[string]*Node{},
	}
	if len(objects) > 1 {
		node.Position = 1
	}
	return node
}

func (n *Node) IsRoot() bool {
	if n.Parent == nil {
		return true
	}
	return false
}

func (n *Node) IsBucketRoot() bool {
	if n.IsRoot() {
		return false
	}
	if n.Parent.IsRoot() {
		return true
	}
	return false
}

type S3ListType int

const (
	BucketList S3ListType = iota //0
	BucketRootList
	ObjectList
)

func (n *Node) GetType() S3ListType {
	if n.IsRoot() {
		return BucketList
	}
	if n.IsBucketRoot() {
		return BucketRootList
	}
	return ObjectList
}

func (n *Node) IsExistChildren(key string) bool {
	_, ok := n.children[key]
	return ok
}

func (n *Node) GetChild(key string) *Node {
	return n.children[key]
}

func (n *Node) AddChild(key string, node *Node) {
	n.children[key] = node
}

func Filename(path string) string {
	sp := strings.Split(path, "/")
	return sp[len(sp)-1]
}
