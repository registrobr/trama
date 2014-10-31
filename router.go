package trama

import (
	"errors"
	"strings"
)

var (
	ErrRouteAlreadyExists = errors.New("Route already exists")
	ErrRouteNotFound      = errors.New("Route not found")
	ErrWildcardConflict   = errors.New("Wildcard node cannot have siblings")
)

type router struct {
	root *node
}

func newRouter() router {
	return router{root: newNode(token{})}
}

func (r *router) appendRoute(uri string, h *adapter) error {
	sequence := newTokenSequence(uri)
	nod, remainingSequence := r.lastNodeThatMatches(sequence)

	if len(remainingSequence) == 0 && nod.hasHandler() {
		return ErrRouteAlreadyExists
	}

	for _, tok := range remainingSequence {
		if !nod.canAddTokenAsChild(tok) {
			return ErrWildcardConflict
		}

		n := newNode(tok)
		nod.addChild(n)
		nod = n
	}

	nod.handler = h
	return nil
}

func (r *router) match(uri string) (*adapter, error) {
	sequence := newTokenSequence(uri)
	node, err := r.findNode(sequence)

	if err != nil {
		return nil, err
	}

	return node.handler, nil
}

func (r *router) lastNodeThatMatches(sequence []token) (*node, []token) {
	current := r.root
	var i int

	for i = 0; i < len(sequence); i++ {
		child := current.child(sequence[i])

		if child == nil {
			break
		}

		current = child
	}

	return current, sequence[i:]
}

func (r *router) findNode(sequence []token) (*node, error) {
	uriVars := make(map[string]string)
	current := r.root

	var lastStatic *node
	for _, tok := range sequence {
		child := current.childForValue(tok.name)

		if child == nil {
			if lastStatic != nil {
				current = lastStatic
				break

			} else {
				return nil, ErrRouteNotFound
			}
		}

		current = child
		if current.handler != nil && current.handler.staticHandler != nil {
			lastStatic = current
		}

		if child.value.isWildcard() {
			uriVars[child.value.parameter] = tok.name
		}
	}

	h := *current.handler
	h.uriVars = uriVars
	return &node{handler: &h}, nil
}

type token struct {
	name      string
	parameter string
}

func newToken(value string) token {
	var t token
	t.set(value)
	return t
}

func valueIsWildcard(value string) bool {
	if len(value) == 0 {
		return false
	}

	return value[0] == '{' && value[len(value)-1] == '}'
}

func (n *token) set(value string) {
	if valueIsWildcard(value) {
		n.parameter = value[1 : len(value)-1]
	} else {
		n.name = value
	}
}

func (n *token) isWildcard() bool {
	return len(n.parameter) > 0
}

func newTokenSequence(uri string) []token {
	uri = strings.TrimSpace(uri)

	// Make sure we are not appending the root ("/"), otherwise remove final slash
	if len(uri) > 1 && uri[len(uri)-1] == '/' {
		uri = uri[:len(uri)-1]
	}

	segments := strings.Split(uri, "/")
	sequence := make([]token, len(segments))

	for i, token := range segments {
		sequence[i] = newToken(token)
	}

	return sequence
}

type node struct {
	value         token
	handler       *adapter
	children      map[token]*node
	wildcardChild *node
}

func newNode(t token) *node {
	return &node{value: t, children: make(map[token]*node)}
}

func (n *node) hasHandler() bool {
	return n.handler != nil
}

func (n *node) addChild(newNode *node) {
	if newNode.value.isWildcard() {
		n.wildcardChild = newNode
	} else {
		n.children[newNode.value] = newNode
	}
}

func (n *node) child(t token) *node {
	if n.wildcardChild != nil && n.wildcardChild.value == t {
		return n.wildcardChild
	}

	return n.children[t]
}

func (n *node) childForValue(value string) *node {
	if n.wildcardChild != nil {
		return n.wildcardChild
	}

	return n.children[token{name: value}]
}

func (n *node) canAddTokenAsChild(t token) bool {
	return n.wildcardChild == nil && (!t.isWildcard() || len(n.children) == 0)
}
