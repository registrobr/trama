package trama

import (
	"errors"
	"strings"
)

var (
	ErrRouteAlreadyExists = errors.New("Route already exists")
	ErrWildcardConflict   = errors.New("Wildcard node cannot have siblings")
)

type router struct {
	root *node
}

func (r *router) appendRoute(uri string, h adapter) error {
	sequence := newTokenSequence(uri)
	nod, remainingSequence := r.lastNodeThatMatches(sequence)

	if len(remainingSequence) == 0 {
		return ErrRouteAlreadyExists
	}

	if nod.wildcardChild != nil {
		return ErrWildcardConflict
	}

	for _, value := range remainingSequence {
		if value.isWildcard() && len(nod.children) > 0 {
			return ErrWildcardConflict
		}

		newNode := &node{value: value}
		nod.addChild(newNode)
		nod = newNode
	}

	nod.handler = h
	return nil
}

func (r *router) match(uri string) (adapter, error) {
	sequence := newTokenSequence(uri)
	node, err := r.findNode(sequence)

	if err != nil {
		return adapter{}, err
	}

	return node.handler, nil
}

func (r *router) lastNodeThatMatches(sequence []token) (*node, sequence) {
	current := r.root

	for i, tok := range sequence {
		child, ok := current.children[tok]

		if ok {
			current = child
		} else {
			sequence = sequence[i:]
			break
		}
	}

	return current, sequence
}

func (r *router) findNode(sequence sequence) (*node, error) {
	uriVars := make(map[string]string)
	current := r.root

	for _, value := range sequence {
		if current.wildcardChild != nil {
			uriVars[current.wildcardChild.value.parameter] = value.name
			current = wildcardChild
		} else {
			var ok bool
			current, ok = current.children[value]

			if !ok {
				return nil, ErrRouteNotFound
			}
		}
	}

	current.handler.uriVars = uriVars
	return current, nil
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
	handler       adapter
	children      map[token]*node
	wildcardChild *node
}

func (n *node) addChild(newNode *node) {
	if newNode.value.isWildcard() {
		n.wildcardChild = node
	} else {
		n.children[newNode.value] = newNode
	}
}
