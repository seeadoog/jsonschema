// expr.y
%{
package ast

import (
)

%}

%token IDENT NUMBER STRING BOOL NIL EQ AND OR NOTEQ GT GTE LT LTE ORR ACC
%left '='
%left ORR
%left '+' '-'
%left '*' '/'
%left '&' '|'
%left '?'
%left ':'
%left AND OR
%left EQ   NOTEQ  GT GTE LT LTE
%right '!'
%right '^'
%left ACC
%right UMINUS


%%

Input:
	   Expr { yylex.(Setter).SetRoot(yyS[yypt-0].node); yyVAL.node = yyS[yypt-0].node }
	;

Expr:
	  Expr '+' Expr       { yyVAL.node = &Binary{Op:"+", L: yyS[yypt-2].node, R: yyS[yypt-0].node} }
	| Expr '-' Expr       { yyVAL.node = &Binary{Op:"-", L: yyS[yypt-2].node, R: yyS[yypt-0].node} }
	| Expr '*' Expr       { yyVAL.node = &Binary{Op:"*", L: yyS[yypt-2].node, R: yyS[yypt-0].node} }
	| Expr '/' Expr       { yyVAL.node = &Binary{Op:"/", L: yyS[yypt-2].node, R: yyS[yypt-0].node} }
	| Expr '^' Expr       { yyVAL.node = &Binary{Op:"^", L: yyS[yypt-2].node, R: yyS[yypt-0].node} }
	| Expr '&' Expr        { yyVAL.node = &Binary{Op:"&", L: yyS[yypt-2].node, R: yyS[yypt-0].node} }
	| Expr '|' Expr        { yyVAL.node = &Binary{Op:"|", L: yyS[yypt-2].node, R: yyS[yypt-0].node} }
	| Expr  EQ Expr        { yyVAL.node = &Binary{Op:"==", L: yyS[yypt-2].node, R: yyS[yypt-0].node} }
	| IDENT  '=' Expr        { yyVAL.node = &Set{L: yyS[yypt-2].str, R: yyS[yypt-0].node} }
	| Expr  AND Expr        { yyVAL.node = &Binary{Op:"&&", L: yyS[yypt-2].node, R: yyS[yypt-0].node} }
	| Expr  OR Expr        { yyVAL.node = &Binary{Op:"||", L: yyS[yypt-2].node, R: yyS[yypt-0].node} }
	| Expr  NOTEQ Expr        { yyVAL.node = &Binary{Op:"!=", L: yyS[yypt-2].node, R: yyS[yypt-0].node} }
	| Expr  GT Expr        { yyVAL.node = &Binary{Op:">", L: yyS[yypt-2].node, R: yyS[yypt-0].node} }
	| Expr  GTE Expr        { yyVAL.node = &Binary{Op:">=", L: yyS[yypt-2].node, R: yyS[yypt-0].node} }
	| Expr  LT Expr        { yyVAL.node = &Binary{Op:"<", L: yyS[yypt-2].node, R: yyS[yypt-0].node} }
	| Expr  LTE Expr        { yyVAL.node = &Binary{Op:"<=", L: yyS[yypt-2].node, R: yyS[yypt-0].node} }
	| Expr  ORR Expr        { yyVAL.node = &Call{Name: "orr", Args: []Node{yyS[yypt-2].node,yyS[yypt-0].node}}  }
	| Expr ACC Expr   { yyVAL.node = &Access{ L: yyS[yypt-2].node, R: yyS[yypt-0].node} }
	| '!' Expr        { yyVAL.node = &Unary{Op:"!", X: yyS[yypt-0].node}  }
	| '-' Expr  %prec UMINUS { yyVAL.node = &Unary{Op:"-", X: yyS[yypt-0].node} }
	| Expr '?' Expr ':' Expr { yyVAL.node = &Call{Name: "ternary", Args: []Node{yyS[yypt-4].node,yyS[yypt-2].node,yyS[yypt-0].node}} }
	| Primary             { yyVAL.node = yyS[yypt-0].node }

	;

Primary:
	  NUMBER              { yyVAL.node = &Number{Val: yyS[yypt-0].num} }
	| BOOL                { yyVAL.node = &Bool{Val:yyS[yypt-0].boolean} }
	| STRING {yyVAL.node = &String{Val: yyS[yypt-0].str}}
	| NIL    {yyVAL.node = &Nil{} }
	| IDENT               { yyVAL.node = &Variable{Name: yyS[yypt-0].str} }
	| IDENT '(' ArgListOpt ')' { yyVAL.node = &Call{Name: yyS[yypt-3].str, Args: yyS[yypt-1].nodes} }
	| '(' Expr ')'        { yyVAL.node = yyS[yypt-1].node }


	;

ArgListOpt:
	  /* empty */         { yyVAL.nodes = nil }
	| ArgList             { yyVAL.nodes = yyS[yypt-0].nodes }
	;

ArgList:
	  Expr                { yyVAL.nodes = []Node{ yyS[yypt-0].node } }
	| ArgList ',' Expr    { yyVAL.nodes = append(yyS[yypt-2].nodes, yyS[yypt-0].node)  }
	;

%%

