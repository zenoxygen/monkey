package parser

import (
	"fmt"

	"github.com/zenoxygen/monkey/ast"
	"github.com/zenoxygen/monkey/lexer"
	"github.com/zenoxygen/monkey/token"
)

const (
	_ int = iota
	LOWEST
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL        // myFunction(X)
)

type Parser struct {
	l         *lexer.Lexer // pointer to instance of lexer
	curToken  token.Token  // pointer to current token
	peekToken token.Token  // pointer to next token
	errors    []string

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

// Initialize a new parser
func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l, errors: []string{}}
	p.nextToken()
	p.nextToken()

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)

	return p
}

// Advance both current token and peek token
func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

// Parse program
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for p.curToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}
	return program
}

// Parse statement
func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

// Parse LET statement
func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken}
	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	for !p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// Parse RETURN statement
func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}
	p.nextToken()
	for !p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// Parse EXPRESSION statement
func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	stmt.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// Parse EXPRESSION
func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		return nil
	}
	leftExp := prefix()

	return leftExp
}

// Parse IDENTIFIER
func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

// Check current token type
func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

// Check peek token type
func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

// Check peek token type and advance tokens
func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

// Return errors
func (p *Parser) Errors() []string {
	return p.errors
}

// Add error to errors when type of peek token doesn’t match expectation
func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("Expected next token to be `%s`. Got `%s` instead", t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

// Add entry to prefix map
func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

// Add entry to infix map
func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}
