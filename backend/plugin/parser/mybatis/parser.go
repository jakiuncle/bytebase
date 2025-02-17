// Package mybatis defines the sql extractor for mybatis mapper xml.
package mybatis

import (
	"encoding/xml"
	"io"
	"strings"

	"github.com/pkg/errors"

	"github.com/bytebase/bytebase/backend/plugin/parser/mybatis/ast"
)

// Parser is the mybatis mapper xml parser.
type Parser struct {
	d           *xml.Decoder
	buf         []rune
	cursor      uint
	currentLine uint
}

// NewParser creates a new mybatis mapper xml parser.
func NewParser(stmt string) *Parser {
	reader := strings.NewReader(stmt)
	d := xml.NewDecoder(reader)
	return &Parser{
		d:      d,
		cursor: 0,
		buf:    nil,
	}
}

// Parse parses the mybatis mapper xml statements, building AST without recursion, returns the root node of the AST.
func (p *Parser) Parse() (ast.Node, error) {
	root := &ast.RootNode{}
	// To avoid recursion, we use stack to store the start element and node, and consume the token one by one.
	// The length of start element stack is always equal to the length of node stack - 1, because the root nod
	// is not in the start element stack.
	var startElementStack []*xml.StartElement
	nodeStack := []ast.Node{root}

	for {
		token, err := p.d.Token()
		if err != nil {
			if err == io.EOF {
				if len(startElementStack) == 0 {
					return root, nil
				}
				return nil, errors.Errorf("expected to read the end element of %q, but got EOF", startElementStack[len(startElementStack)-1].Name.Local)
			}
			return nil, errors.Wrapf(err, "failed to get token from xml decoder")
		}
		switch ele := token.(type) {
		case xml.StartElement:
			newNode := p.newNodeByStartElement(&ele)
			startElementStack = append(startElementStack, &ele)
			nodeStack = append(nodeStack, newNode)
		case xml.EndElement:
			if len(startElementStack) == 0 {
				return nil, errors.Errorf("unexpected end element %q", ele.Name.Local)
			}
			if ele.Name.Local != startElementStack[len(startElementStack)-1].Name.Local {
				return nil, errors.Errorf("expected to read the name of end element is %q, but got %q", startElementStack[len(startElementStack)-1].Name.Local, ele.Name.Local)
			}
			// We will pop the start element stack and node stack at the same time.
			startElementStack = startElementStack[:len(startElementStack)-1]
			popNode := nodeStack[len(nodeStack)-1]
			// To avoid keeping many empty node in AST, we only add the node which is not an empty node to the parent node.
			if _, ok := popNode.(*ast.EmptyNode); !ok {
				nodeStack[len(nodeStack)-2].AddChild(popNode)
			}
			nodeStack = nodeStack[:len(nodeStack)-1]
		case xml.CharData:
			for _, b := range ele {
				if b == '\n' {
					p.currentLine++
				}
			}
			trimmed := strings.TrimSpace(string(ele))
			if len(trimmed) == 0 {
				continue
			}
			dataNode := ast.NewDataNode([]byte(trimmed))
			if err := dataNode.Scan(); err != nil {
				return nil, errors.Wrapf(err, "cannot parse data node")
			}
			if len(nodeStack) == 0 {
				return nil, errors.Errorf("try to append data node to parent node, but node stack is empty")
			}
			nodeStack[len(nodeStack)-1].AddChild(dataNode)
		case xml.Comment:
			for _, b := range ele {
				if b == '\n' {
					p.currentLine++
				}
			}
		}
	}
}

// newNodeByStartElement returns the node related to the startElement, for example, returns QueryNode for
// start element which name is "select", "update", "insert", "delete". If the startElement is unacceptable,
// returns an emptyNode instead.
func (*Parser) newNodeByStartElement(startElement *xml.StartElement) ast.Node {
	switch startElement.Name.Local {
	case "mapper":
		return ast.NewMapperNode(startElement)
	case "select", "update", "insert", "delete":
		return ast.NewQueryNode(startElement)
	case "if":
		return ast.NewIfNode(startElement)
	case "choose":
		return ast.NewChooseNode(startElement)
	case "when":
		return ast.NewWhenNode(startElement)
	case "otherwise":
		return ast.NewOtherwiseNode(startElement)
	}
	return ast.NewEmptyNode()
}
