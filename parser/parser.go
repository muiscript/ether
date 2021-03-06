package parser

import (
	"fmt"
	"github.com/muiscript/ether/ast"
	"github.com/muiscript/ether/lexer"
	"github.com/muiscript/ether/token"
	"strconv"
)

type Precedence int

const (
	LOWEST Precedence = iota
	EQUAL
	COMPARISON
	ARROW
	ADDITION
	MULTIPLICATION
	PREFIX
	CALL
	INDEX
)

var (
	TRUE_NODE  = &ast.BooleanLiteral{Value: true}
	FALSE_NODE = &ast.BooleanLiteral{Value: false}
)

func precedence(t token.Token) Precedence {
	switch t.Type {
	case token.ARROW:
		return ARROW
	case token.EQ, token.NEQ:
		return EQUAL
	case token.GT, token.LT:
		return COMPARISON
	case token.PLUS, token.MINUS:
		return ADDITION
	case token.ASTER, token.SLASH, token.PERCENT:
		return MULTIPLICATION
	case token.LPAREN:
		return CALL
	case token.LBRACKET:
		return INDEX
	default:
		return LOWEST
	}
}

type Parser struct {
	lexer        *lexer.Lexer
	currentToken token.Token
	peekToken    token.Token
	errors       []*ParserError
}

func New(lexer *lexer.Lexer) *Parser {
	parser := &Parser{lexer: lexer}
	parser.consumeToken()
	parser.consumeToken()

	return parser
}

func (p *Parser) ParseProgram() (*ast.Program, error) {
	statements := make([]ast.Statement, 0)

	for p.currentToken.Type != token.EOF {
		statement, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		statements = append(statements, statement)
		p.consumeToken()
	}

	return &ast.Program{Statements: statements}, nil
}

func (p *Parser) consumeToken() {
	p.currentToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
}

func (p *Parser) expectToken(tokenType token.Type) error {
	if p.peekToken.Type != tokenType {
		return &ParserError{
			line: p.peekToken.Line,
			msg:  fmt.Sprintf("unexpected token.\nwant=%v\ngot=%v (%+v)\n", tokenType, p.peekToken.Type, p.peekToken),
		}
	}
	p.consumeToken()
	return nil
}

func (p *Parser) currentPrecedence() Precedence {
	return precedence(p.currentToken)
}

func (p *Parser) peekPrecedence() Precedence {
	return precedence(p.peekToken)
}

func (p *Parser) parseStatement() (ast.Statement, error) {
	switch p.currentToken.Type {
	case token.VAR:
		return p.parseVarStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseVarStatement() (*ast.VarStatement, error) {
	line := p.currentToken.Line
	p.consumeToken()

	identifier, err := p.parseIdentifier()
	if err != nil {
		return nil, err
	}

	if err := p.expectToken(token.ASSIGN); err != nil {
		return nil, err
	}
	p.consumeToken()

	expression, err := p.parseExpression(LOWEST)
	if err != nil {
		return nil, err
	}
	if p.peekToken.Type == token.SEMICOLON {
		p.consumeToken()
	}

	return ast.NewVarStatement(identifier, expression, line), nil
}

func (p *Parser) parseReturnStatement() (*ast.ReturnStatement, error) {
	line := p.currentToken.Line
	p.consumeToken()

	expression, err := p.parseExpression(LOWEST)
	if err != nil {
		return nil, err
	}
	if p.peekToken.Type == token.SEMICOLON {
		p.consumeToken()
	}

	return ast.NewReturnStatement(expression, line), nil
}

func (p *Parser) parseExpressionStatement() (*ast.ExpressionStatement, error) {
	line := p.currentToken.Line
	expression, err := p.parseExpression(LOWEST)
	if err != nil {
		return nil, err
	}
	if p.peekToken.Type == token.SEMICOLON {
		p.consumeToken()
	}

	return ast.NewExpressionStatement(expression, line), nil
}

func (p *Parser) parseBlockStatement() (*ast.BlockStatement, error) {
	line := p.currentToken.Line
	p.consumeToken()
	statements := make([]ast.Statement, 0)

	for p.currentToken.Type != token.RBRACE {
		statement, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		statements = append(statements, statement)
		p.consumeToken()
	}

	return ast.NewBlockStatement(statements, line), nil
}

func (p *Parser) parseExpression(precedence Precedence) (ast.Expression, error) {
	var left ast.Expression
	var err error
	switch p.currentToken.Type {
	case token.INTEGER:
		left, err = p.parseIntegerLiteral()
	case token.TRUE, token.FALSE:
		left, err = p.parseBooleanLiteral()
	case token.IDENT:
		left, err = p.parseIdentifier()
	case token.MINUS, token.BANG:
		left, err = p.parsePrefixExpression()
	case token.LPAREN:
		left, err = p.parseGroupedExpression()
	case token.BAR:
		left, err = p.parseFunctionLiteral()
	case token.IF:
		left, err = p.parseIfExpression()
	case token.LBRACKET:
		left, err = p.parseArrayLiteral()
	default:
		return nil, &ParserError{line: p.currentToken.Line, msg: fmt.Sprintf("unable to parse prefix token %+v\n", p.currentToken)}
	}
	if err != nil {
		return nil, err
	}

	for precedence < p.peekPrecedence() {
		p.consumeToken()
		switch p.currentToken.Type {
		case token.LPAREN:
			left, err = p.parseFunctionCall(left)
		case token.LBRACKET:
			left, err = p.parseIndexExpression(left)
		case token.ARROW:
			left, err = p.parseArrowExpression(left)
		default:
			left, err = p.parseInfixExpression(left)
		}
		if err != nil {
			return nil, err
		}
	}

	return left, nil
}

func (p *Parser) parseIntegerLiteral() (*ast.IntegerLiteral, error) {
	line := p.currentToken.Line
	v, err := strconv.Atoi(p.currentToken.Literal)
	if err != nil {
		return nil, &ParserError{line: line, msg: err.Error()}
	}
	return ast.NewIntegerLiteral(v, line), nil
}

func (p *Parser) parseBooleanLiteral() (*ast.BooleanLiteral, error) {
	line := p.currentToken.Line
	switch p.currentToken.Type {
	case token.TRUE:
		return TRUE_NODE, nil
	case token.FALSE:
		return FALSE_NODE, nil
	default:
		return nil, &ParserError{line: line, msg: fmt.Sprintf("not boolean: %+v", p.currentToken)}
	}
}

func (p *Parser) parseIdentifier() (*ast.Identifier, error) {
	line := p.currentToken.Line
	if p.currentToken.Type != token.IDENT {
		return nil, &ParserError{line: line, msg: fmt.Sprintf("not identifier: %+v", p.currentToken)}
	}
	return ast.NewIdentifier(p.currentToken.Literal, line), nil
}

func (p *Parser) parsePrefixExpression() (*ast.PrefixExpression, error) {
	line := p.currentToken.Line
	operator := p.currentToken.Literal
	p.consumeToken()
	right, err := p.parseExpression(PREFIX)
	if err != nil {
		return nil, err
	}
	return ast.NewPrefixExpression(operator, right, line), nil
}

func (p *Parser) parseGroupedExpression() (ast.Expression, error) {
	p.consumeToken()
	expression, err := p.parseExpression(LOWEST)
	if err != nil {
		return nil, err
	}
	if err := p.expectToken(token.RPAREN); err != nil {
		return nil, err
	}
	return expression, nil
}

func (p *Parser) parseIfExpression() (ast.Expression, error) {
	line := p.currentToken.Line

	if err := p.expectToken(token.LPAREN); err != nil {
		return nil, err
	}
	p.consumeToken()
	condition, err := p.parseExpression(LOWEST)
	if err != nil {
		return nil, err
	}
	if err := p.expectToken(token.RPAREN); err != nil {
		return nil, err
	}

	if err := p.expectToken(token.LBRACE); err != nil {
		return nil, err
	}
	consequence, err := p.parseBlockStatement()
	if err != nil {
		return nil, err
	}

	if p.peekToken.Type != token.ELSE {
		return ast.NewIfExpression(condition, consequence, nil, line), nil
	}
	p.consumeToken()
	p.consumeToken()
	alternative, err := p.parseBlockStatement()
	if err != nil {
		return nil, err
	}
	return ast.NewIfExpression(condition, consequence, alternative, line), nil
}

func (p *Parser) parseFunctionLiteral() (ast.Expression, error) {
	line := p.currentToken.Line
	expressions, err := p.parseCommaSeparatedExpressions(token.BAR)
	if err != nil {
		return nil, err
	}
	var parameters []*ast.Identifier
	for _, expression := range expressions {
		if parameter, ok := expression.(*ast.Identifier); ok {
			parameters = append(parameters, parameter)
		} else {
			return nil, &ParserError{line: line, msg: fmt.Sprintf("unable to parse function parameter: %+v", expression)}
		}
	}

	if err := p.expectToken(token.LBRACE); err != nil {
		return nil, err
	}
	body, err := p.parseBlockStatement()
	if err != nil {
		return nil, err
	}

	return ast.NewFunctionLiteral(parameters, body, line), nil
}

func (p *Parser) parseArrayLiteral() (ast.Expression, error) {
	line := p.currentToken.Line
	elements, err := p.parseCommaSeparatedExpressions(token.RBRACKET)
	if err != nil {
		return nil, err
	}

	return ast.NewArrayLiteral(elements, line), nil
}

func (p *Parser) parseInfixExpression(left ast.Expression) (*ast.InfixExpression, error) {
	line := p.currentToken.Line
	precedence := p.currentPrecedence()
	operator := p.currentToken.Literal
	p.consumeToken()
	right, err := p.parseExpression(precedence)
	if err != nil {
		return nil, err
	}
	return ast.NewInfixExpression(operator, left, right, line), nil
}

func (p *Parser) parseFunctionCall(left ast.Expression) (*ast.FunctionCall, error) {
	line := p.currentToken.Line
	arguments, err := p.parseCommaSeparatedExpressions(token.RPAREN)
	if err != nil {
		return nil, err
	}
	return ast.NewFunctionCall(left, arguments, line), nil
}

func (p *Parser) parseIndexExpression(left ast.Expression) (*ast.IndexExpression, error) {
	line := p.currentToken.Line
	p.consumeToken()
	index, err := p.parseExpression(LOWEST)
	if err != nil {
		return nil, err
	}
	if err := p.expectToken(token.RBRACKET); err != nil {
		return nil, err
	}
	return ast.NewIndexExpression(left, index, line), nil
}

func (p *Parser) parseArrowExpression(left ast.Expression) (*ast.FunctionCall, error) {
	line := p.currentToken.Line
	p.consumeToken()
	right, err := p.parseExpression(ARROW)
	if err != nil {
		return nil, err
	}
	rightCall, ok := right.(*ast.FunctionCall)
	if !ok {
		return nil, &ParserError{line: line, msg: fmt.Sprintf("right of '->' should be function call. got=%+v (%T)\n", right, right)}
	}
	rightCall.Arguments = append([]ast.Expression{left}, rightCall.Arguments...)

	return rightCall, nil
}

func (p *Parser) parseCommaSeparatedExpressions(endTokenType token.Type) ([]ast.Expression, error) {
	p.consumeToken()
	if p.currentToken.Type == endTokenType {
		return []ast.Expression{}, nil
	}

	first, err := p.parseExpression(LOWEST)
	if err != nil {
		return nil, err
	}
	expressions := []ast.Expression{first}

	for p.peekToken.Type != endTokenType {
		if err := p.expectToken(token.COMMA); err != nil {
			return nil, err
		}
		p.consumeToken()

		expression, err := p.parseExpression(LOWEST)
		if err != nil {
			return nil, err
		}
		expressions = append(expressions, expression)
	}
	if err := p.expectToken(endTokenType); err != nil {
		return nil, err
	}

	return expressions, nil
}
