package token

type Type string

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	// identifier and literal
	IDENT   = "IDENT"
	INTEGER = "INTEGER"

	// operators
	ASSIGN  = "ASSIGN"
	PLUS    = "PLUS"
	MINUS   = "MINUS"
	ASTER   = "ASTER"
	SLASH   = "SLASH"
	PERCENT = "PERCENT"
	ARROW   = "ARROW"
	BANG    = "BANG"
	LT      = "LT"
	GT      = "GT"
	EQ      = "EQ"
	NEQ     = "NEQ"

	// delimiters
	COMMA     = "COMMA"
	SEMICOLON = "SEMICOLON"
	LPAREN    = "LPAREN"
	RPAREN    = "RPAREN"
	LBRACE    = "LBRACE"
	RBRACE    = "RBRACE"
	LBRACKET  = "LBRACKET"
	RBRACKET  = "RBRACKET"
	BAR       = "BAR"

	// keywords
	VAR    = "VAR"
	RETURN = "RETURN"
	TRUE   = "TRUE"
	FALSE  = "FALSE"
	IF     = "IF"
	ELSE   = "ELSE"
)

type Token struct {
	Type    Type
	Literal string
	Line    int
}

func TypeByLiteral(literal string) Type {
	switch literal {
	case "var":
		return VAR
	case "return":
		return RETURN
	case "true":
		return TRUE
	case "false":
		return FALSE
	case "if":
		return IF
	case "else":
		return ELSE
	default:
		return IDENT
	}
}
