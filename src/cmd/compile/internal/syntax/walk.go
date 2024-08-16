// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file implements syntax tree walking and changing.

package syntax

import "fmt"

// Inspect traverses an AST in pre-order: it starts by calling f(root);
// root must not be nil. If f returns true, Inspect invokes f recursively
// for each of the non-nil children of root, followed by a call of f(nil).
//
// See Walk for caveats about shared nodes.
func Inspect(root Node, f func(Node) bool) {
	Walk(root, inspector(f))
}

type inspector func(Node) bool

func (v inspector) Visit(node Node) Visitor {
	if v(node) {
		return v
	}
	return nil
}

// Walk traverses an AST in pre-order: It starts by calling
// v.Visit(node); node must not be nil. If the visitor w returned by
// v.Visit(node) is not nil, Walk is invoked recursively with visitor
// w for each of the non-nil children of node, followed by a call of
// w.Visit(nil).
//
// Some nodes may be shared among multiple parent nodes (e.g., types in
// field lists such as type T in "a, b, c T"). Such shared nodes are
// walked multiple times.
// TODO(gri) Revisit this design. It may make sense to walk those nodes
// only once. A place where this matters is types2.TestResolveIdents.
func Walk(root Node, v Visitor) {
	walker{v}.node(root)
}

// A Visitor's Visit method is invoked for each node encountered by Walk.
// If the result visitor w is not nil, Walk visits each of the children
// of node with the visitor w, followed by a call of w.Visit(nil).
type Visitor interface {
	Visit(node Node) (w Visitor)
}

type walker struct {
	v Visitor
}

func (w walker) node(n Node) {
	if n == nil {
		panic("nil node")
	}

	w.v = w.v.Visit(n)
	if w.v == nil {
		return
	}

	switch n := n.(type) {
	// packages
	case *File:
		w.node(n.PkgName)
		w.declList(n.DeclList)

	// declarations
	case *ImportDecl:
		if n.LocalPkgName != nil {
			w.node(n.LocalPkgName)
		}
		w.node(n.Path)

	case *ConstDecl:
		w.nameList(n.NameList)
		if n.Type != nil {
			w.node(n.Type)
		}
		if n.Values != nil {
			w.node(n.Values)
		}

	case *TypeDecl:
		w.node(n.Name)
		w.fieldList(n.TParamList)
		w.node(n.Type)

	case *VarDecl:
		w.nameList(n.NameList)
		if n.Type != nil {
			w.node(n.Type)
		}
		if n.Values != nil {
			w.node(n.Values)
		}

	case *FuncDecl:
		if n.Recv != nil {
			w.node(n.Recv)
		}
		w.node(n.Name)
		w.fieldList(n.TParamList)
		w.node(n.Type)
		if n.Body != nil {
			w.node(n.Body)
		}

	// expressions
	case *BadExpr: // nothing to do
	case *Name: // nothing to do
	case *BasicLit: // nothing to do

	case *CompositeLit:
		if n.Type != nil {
			w.node(n.Type)
		}
		w.exprList(n.ElemList)

	case *KeyValueExpr:
		w.node(n.Key)
		w.node(n.Value)

	case *FuncLit:
		w.node(n.Type)
		w.node(n.Body)

	case *ParenExpr:
		w.node(n.X)

	case *SelectorExpr:
		w.node(n.X)
		w.node(n.Sel)

	case *IndexExpr:
		w.node(n.X)
		w.node(n.Index)

	case *SliceExpr:
		w.node(n.X)
		for _, x := range n.Index {
			if x != nil {
				w.node(x)
			}
		}

	case *AssertExpr:
		w.node(n.X)
		w.node(n.Type)

	case *TypeSwitchGuard:
		if n.Lhs != nil {
			w.node(n.Lhs)
		}
		w.node(n.X)

	case *Operation:
		w.node(n.X)
		if n.Y != nil {
			w.node(n.Y)
		}

	case *CallExpr:
		w.node(n.Fun)
		w.exprList(n.ArgList)

	case *ListExpr:
		w.exprList(n.ElemList)

	// types
	case *ArrayType:
		if n.Len != nil {
			w.node(n.Len)
		}
		w.node(n.Elem)

	case *SliceType:
		w.node(n.Elem)

	case *DotsType:
		w.node(n.Elem)

	case *StructType:
		w.fieldList(n.FieldList)
		for _, t := range n.TagList {
			if t != nil {
				w.node(t)
			}
		}

	case *Field:
		if n.Name != nil {
			w.node(n.Name)
		}
		w.node(n.Type)

	case *InterfaceType:
		w.fieldList(n.MethodList)

	case *FuncType:
		w.fieldList(n.ParamList)
		w.fieldList(n.ResultList)

	case *MapType:
		w.node(n.Key)
		w.node(n.Value)

	case *ChanType:
		w.node(n.Elem)

	// statements
	case *EmptyStmt: // nothing to do

	case *LabeledStmt:
		w.node(n.Label)
		w.node(n.Stmt)

	case *BlockStmt:
		w.stmtList(n.List)

	case *ExprStmt:
		w.node(n.X)

	case *SendStmt:
		w.node(n.Chan)
		w.node(n.Value)

	case *DeclStmt:
		w.declList(n.DeclList)

	case *AssignStmt:
		w.node(n.Lhs)
		if n.Rhs != nil {
			w.node(n.Rhs)
		}

	case *BranchStmt:
		if n.Label != nil {
			w.node(n.Label)
		}
		// Target points to nodes elsewhere in the syntax tree

	case *CallStmt:
		w.node(n.Call)

	case *ReturnStmt:
		if n.Results != nil {
			w.node(n.Results)
		}

	case *IfStmt:
		if n.Init != nil {
			w.node(n.Init)
		}
		w.node(n.Cond)
		w.node(n.Then)
		if n.Else != nil {
			w.node(n.Else)
		}

	case *ForStmt:
		if n.Init != nil {
			w.node(n.Init)
		}
		if n.Cond != nil {
			w.node(n.Cond)
		}
		if n.Post != nil {
			w.node(n.Post)
		}
		w.node(n.Body)

	case *SwitchStmt:
		if n.Init != nil {
			w.node(n.Init)
		}
		if n.Tag != nil {
			w.node(n.Tag)
		}
		for _, s := range n.Body {
			w.node(s)
		}

	case *SelectStmt:
		for _, s := range n.Body {
			w.node(s)
		}

	// helper nodes
	case *RangeClause:
		if n.Lhs != nil {
			w.node(n.Lhs)
		}
		w.node(n.X)

	case *CaseClause:
		if n.Cases != nil {
			w.node(n.Cases)
		}
		w.stmtList(n.Body)

	case *CommClause:
		if n.Comm != nil {
			w.node(n.Comm)
		}
		w.stmtList(n.Body)

	default:
		panic(fmt.Sprintf("internal error: unknown node type %T", n))
	}

	w.v.Visit(nil)
}

func (w walker) declList(list []Decl) {
	for _, n := range list {
		w.node(n)
	}
}

func (w walker) exprList(list []Expr) {
	for _, n := range list {
		w.node(n)
	}
}

func (w walker) stmtList(list []Stmt) {
	for _, n := range list {
		w.node(n)
	}
}

func (w walker) nameList(list []*Name) {
	for _, n := range list {
		w.node(n)
	}
}

func (w walker) fieldList(list []*Field) {
	for _, n := range list {
		w.node(n)
	}
}

func WalkAndChange(root Node, f func(*Node) bool) Node {
	return ASTChanger{changer(f)}.node(root)
}

type changer func(*Node) bool

func (v changer) Change(node *Node) NodeChanger {
	if v(node) {
		return v
	}
	return nil
}

type NodeChanger interface {
	Change(node *Node) NodeChanger
}

type ASTChanger struct {
	changer NodeChanger
}

func (c ASTChanger) node(o Node) Node {
	if o == nil {
		panic("nil node")
	}

	c.changer = c.changer.Change(&o)
	if c.changer == nil {
		return o
	}

	switch n := (o).(type) {
	// packages
	case *File:
		n.PkgName = c.node(n.PkgName).(*Name)
		n.DeclList = c.declList(n.DeclList)

	// declarations
	case *ImportDecl:
		if n.LocalPkgName != nil {
			n.LocalPkgName = c.node(n.LocalPkgName).(*Name)
		}
		n.Path = c.node(n.Path).(*BasicLit)

	case *ConstDecl:
		n.NameList = c.nameList(n.NameList)
		if n.Type != nil {
			n.Type = c.node(n.Type).(Expr)
		}
		if n.Values != nil {
			n.Values = c.node(n.Values).(Expr)
		}

	case *TypeDecl:
		n.Name = c.node(n.Name).(*Name)
		n.TParamList = c.fieldList(n.TParamList)
		n.Type = c.node(n.Type).(Expr)

	case *VarDecl:
		n.NameList = c.nameList(n.NameList)
		if n.Type != nil {
			n.Type = c.node(n.Type).(Expr)
		}
		if n.Values != nil {
			n.Values = c.node(n.Values).(Expr)
		}

	case *FuncDecl:
		if n.Recv != nil {
			n.Recv = c.node(n.Recv).(*Field)
		}
		n.Name = c.node(n.Name).(*Name)
		n.TParamList = c.fieldList(n.TParamList)
		n.Type = c.node(n.Type).(*FuncType)
		if n.Body != nil {
			n.Body = c.node(n.Body).(*BlockStmt)
		}

	// expressions
	case *BadExpr: // nothing to do
	case *Name: // nothing to do
	case *BasicLit: // nothing to do

	case *CompositeLit:
		if n.Type != nil {
			n.Type = c.node(n.Type).(Expr)
		}
		n.ElemList = c.exprList(n.ElemList)

	case *KeyValueExpr:
		n.Key = c.node(n.Key).(Expr)
		n.Value = c.node(n.Value).(Expr)

	case *FuncLit:
		n.Type = c.node(n.Type).(*FuncType)
		n.Body = c.node(n.Body).(*BlockStmt)

	case *ParenExpr:
		n.X = c.node(n.X).(Expr)

	case *SelectorExpr:
		n.X = c.node(n.X).(Expr)
		n.Sel = c.node(n.Sel).(*Name)

	case *IndexExpr:
		n.X = c.node(n.X).(Expr)
		n.Index = c.node(n.Index).(Expr)

	case *SliceExpr:
		n.X = c.node(n.X).(Expr)
		for i, x := range n.Index {
			if x != nil {
				n.Index[i] = c.node(x).(Expr)
			}
		}

	case *AssertExpr:
		n.X = c.node(n.X).(Expr)
		n.Type = c.node(n.Type).(Expr)

	case *TypeSwitchGuard:
		if n.Lhs != nil {
			n.Lhs = c.node(n.Lhs).(*Name)
		}
		n.X = c.node(n.X).(Expr)

	case *Operation:
		n.X = c.node(n.X).(Expr)
		if n.Y != nil {
			n.Y = c.node(n.Y).(Expr)
		}

	case *CallExpr:
		n.Fun = c.node(n.Fun).(Expr)
		n.ArgList = c.exprList(n.ArgList)

	case *ListExpr:
		n.ElemList = c.exprList(n.ElemList)

	// types
	case *ArrayType:
		if n.Len != nil {
			n.Len = c.node(n.Len).(Expr)
		}
		n.Elem = c.node(n.Elem).(Expr)

	case *SliceType:
		n.Elem = c.node(n.Elem).(Expr)

	case *DotsType:
		n.Elem = c.node(n.Elem).(Expr)

	case *StructType:
		n.FieldList = c.fieldList(n.FieldList)
		for i, t := range n.TagList {
			if t != nil {
				n.TagList[i] = c.node(t).(*BasicLit)
			}
		}

	case *Field:
		if n.Name != nil {
			n.Name = c.node(n.Name).(*Name)
		}
		n.Type = c.node(n.Type).(Expr)

	case *InterfaceType:
		n.MethodList = c.fieldList(n.MethodList)

	case *FuncType:
		n.ParamList = c.fieldList(n.ParamList)
		n.ResultList = c.fieldList(n.ResultList)

	case *MapType:
		n.Key = c.node(n.Key).(Expr)
		n.Value = c.node(n.Value).(Expr)

	case *ChanType:
		n.Elem = c.node(n.Elem).(Expr)

	// statements
	case *EmptyStmt: // nothing to do

	case *LabeledStmt:
		n.Label = c.node(n.Label).(*Name)
		n.Stmt = c.node(n.Stmt).(Stmt)

	case *BlockStmt:
		n.List = c.stmtList(n.List)

	case *ExprStmt:
		n.X = c.node(n.X).(Expr)

	case *SendStmt:
		n.Chan = c.node(n.Chan).(Expr)
		n.Value = c.node(n.Value).(Expr)

	case *DeclStmt:
		n.DeclList = c.declList(n.DeclList)

	case *AssignStmt:
		n.Lhs = c.node(n.Lhs).(Expr)
		if n.Rhs != nil {
			n.Rhs = c.node(n.Rhs).(Expr)
		}

	case *BranchStmt:
		if n.Label != nil {
			n.Label = c.node(n.Label).(*Name)
		}
		// Target points to nodes elsewhere in the syntax tree

	case *CallStmt:
		n.Call = c.node(n.Call).(Expr)

	case *ReturnStmt:
		if n.Results != nil {
			n.Results = c.node(n.Results).(Expr)
		}

	case *IfStmt:
		if n.Init != nil {
			v := c.node(n.Init).(SimpleStmt)
			n.Init = v
		}
		n.Cond = c.node(n.Cond).(Expr)
		n.Then = c.node(n.Then).(*BlockStmt)
		if n.Else != nil {
			n.Else = c.node(n.Else).(Stmt)
		}

	case *ForStmt:
		if n.Init != nil {
			n.Init = c.node(n.Init).(SimpleStmt)
		}
		if n.Cond != nil {
			n.Cond = c.node(n.Cond).(Expr)
		}
		if n.Post != nil {
			n.Post = c.node(n.Post).(SimpleStmt)
		}
		n.Body = c.node(n.Body).(*BlockStmt)

	case *SwitchStmt:
		if n.Init != nil {
			n.Init = c.node(n.Init).(SimpleStmt)
		}
		if n.Tag != nil {
			n.Tag = c.node(n.Tag).(Expr)
		}
		for i, s := range n.Body {
			n.Body[i] = c.node(s).(*CaseClause)
		}

	case *SelectStmt:
		for i, s := range n.Body {
			n.Body[i] = c.node(s).(*CommClause)
		}

	// helper nodes
	case *RangeClause:
		if n.Lhs != nil {
			n.Lhs = c.node(n.Lhs).(Expr)
		}
		n.X = c.node(n.X).(Expr)

	case *CaseClause:
		if n.Cases != nil {
			n.Cases = c.node(n.Cases).(Expr)
		}
		n.Body = c.stmtList(n.Body)

	case *CommClause:
		if n.Comm != nil {
			n.Comm = c.node(n.Comm).(SimpleStmt)
		}
		n.Body = c.stmtList(n.Body)

	default:
		panic(fmt.Sprintf("internal error: unknown node type %T", n))
	}

	c.changer.Change(nil)
	return o
}

func (c ASTChanger) declList(list []Decl) []Decl {
	for i, n := range list {
		list[i] = c.node(n).(Decl)
	}
	return list
}

func (c ASTChanger) exprList(list []Expr) []Expr {
	for i, n := range list {
		list[i] = c.node(n).(Expr)
	}
	return list
}

func (c ASTChanger) stmtList(list []Stmt) []Stmt {
	for i, n := range list {
		list[i] = c.node(n).(Stmt)
	}
	return list
}

func (c ASTChanger) nameList(list []*Name) []*Name {
	for i, n := range list {
		list[i] = c.node(n).(*Name)
	}
	return list
}

func (c ASTChanger) fieldList(list []*Field) []*Field {
	for i, n := range list {
		list[i] = c.node(n).(*Field)
	}
	return list
}
