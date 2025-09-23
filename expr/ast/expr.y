// expr.y
%{
package ast

import (
)

%}

%token IDENT NUMBER STRING BOOL NIL EQ AND OR NOTEQ GT GTE LT LTE ORR ACC IF ELSE FOR IN ACC2
%left IDENT
%left IF ELSE
%left ';'
%right '='
%left '?'
%left ':'
%left ORR
%left '+' '-'
%left '*' '/'
%left '&' '|'
%left OR
%left AND
%left EQ   NOTEQ  GT GTE LT LTE
%right '!'
%right '^'
%left ACC '[' ']'
%right UMINUS
%right ACC2



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
	| Expr  EQ Expr        { yyVAL.node = &Binary{Op:"==", L: yyS[yypt-2].node, R: yyS[yypt-0].node} }
	| Expr  ';' Expr        { yyVAL.node = &Binary{Op:";", L: yyS[yypt-2].node, R: yyS[yypt-0].node} }
	| Ident  '=' Expr        { yyVAL.node = &Set{L: yyS[yypt-2].node, R: yyS[yypt-0].node} }
	| Var  '=' Expr        { yyVAL.node = &Set{L: yyS[yypt-2].node, R: yyS[yypt-0].node} }
	| ArrIndex  '=' Expr        { yyVAL.node = &Set{L: yyS[yypt-2].node, R: yyS[yypt-0].node} }
	| Expr  AND Expr        { yyVAL.node = &Binary{Op:"&&", L: yyS[yypt-2].node, R: yyS[yypt-0].node} }
	| Expr  OR Expr        { yyVAL.node = &Binary{Op:"||", L: yyS[yypt-2].node, R: yyS[yypt-0].node} }
	| Expr  NOTEQ Expr        { yyVAL.node = &Binary{Op:"!=", L: yyS[yypt-2].node, R: yyS[yypt-0].node} }
	| Expr  GT Expr        { yyVAL.node = &Binary{Op:">", L: yyS[yypt-2].node, R: yyS[yypt-0].node} }
	| Expr  GTE Expr        { yyVAL.node = &Binary{Op:">=", L: yyS[yypt-2].node, R: yyS[yypt-0].node} }
	| Expr  LT Expr        { yyVAL.node = &Binary{Op:"<", L: yyS[yypt-2].node, R: yyS[yypt-0].node} }
	| Expr  LTE Expr        { yyVAL.node = &Binary{Op:"<=", L: yyS[yypt-2].node, R: yyS[yypt-0].node} }
	| Expr  ORR Expr        { yyVAL.node = &Call{Name: "orr", Args: []Node{yyS[yypt-2].node,yyS[yypt-0].node}}  }
	| '!' Expr        { yyVAL.node = &Unary{Op:"!", X: yyS[yypt-0].node}  }
	| '-' Expr  %prec UMINUS { yyVAL.node = &Unary{Op:"-", X: yyS[yypt-0].node} }
	| Expr '?' Expr ':' Expr { yyVAL.node = &Call{Name: "ternary", Args: []Node{yyS[yypt-4].node,yyS[yypt-2].node,yyS[yypt-0].node}} }
	| Primary             { yyVAL.node = yyS[yypt-0].node }

	;

Var:
    Expr ACC Ident  %prec ACC  { $$.node = &Access{L: $1.node,R:$3.node}}
    |Expr ACC ArrIndex  %prec ACC  { $$.node = &Access{L: $1.node,R:$3.node}}

Acc:
    Expr ACC Expr %prec ACC     {  $$.node = &Access{L: $1.node,R:$3.node} }

Ident:
    IDENT                       { $$.node = &Variable{Name: $1.str} }

Primary:
	  NUMBER              { yyVAL.node = &Number{Val: yyS[yypt-0].num} }
	| BOOL                { yyVAL.node = &Bool{Val:yyS[yypt-0].boolean} }
	| STRING {yyVAL.node = &String{Val: yyS[yypt-0].str}}
	| NIL    {yyVAL.node = &Nil{} }
	| Ident               { $$.node = $1.node }
	| IDENT '(' ArgListOpt ')' { yyVAL.node = &Call{Name: yyS[yypt-3].str, Args: yyS[yypt-1].nodes} }
	| '(' Expr ')'        { yyVAL.node = yyS[yypt-1].node }
	| '{' KvsOpt '}'  { $$.node = &MapSet{Kvs: $2.kvs}  }
	| '[' ArgListOpt ']' { $$.node = &ArrDef{V:$2.nodes}  }
	| ArrIndex {  $$.node = $1.node }
	| Acc   {   $$.node = $1.node }
	| Expr '[' SliceN ':' SliceN ']' {  $$.node = &SliceCut{V: $1.node,St: $3.node,Ed:$5.node} }
	;

SliceN:
    /*empty*/ { $$.node = nil }
    |Expr { $$.node = $1.node}

ArrIndex:
    Expr '[' Expr ']'	 {  $$.node = &ArrAccess{L:$1.node ,R:$3.node} }

KvsOpt:
    /*empty*/ {  $$.kvs = nil  }
    |Kvs { $$.kvs = $1.kvs  }
    ;
Kvs:
    Kv  {   $$.kvs = []KV{$1.kv}}
    |Kvs ',' Kv { $$.kvs  = append($1.kvs,$3.kv) }

    ;
Kv:
    Expr ':' Expr {  $$.kv = KV{ K:$1.node, V: $3.node} }
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

