package sqlfmt

import (
	"errors"
)

func Parse(lexer *sqlLex) (stmt *SelectStmt, err error) {
	if rc := yyParse(lexer); rc != 0 {
		return nil, errors.New("Parse failed")
	}

	return lexer.stmt, nil
}

type Expr interface {
	RenderTo(Renderer)
}

type PgType struct {
	Name         AnyName
	OptInterval  *OptInterval
	Setof        bool
	ArrayWord    bool
	ArrayBounds  []IntegerConst
	TypeMods     []Expr
	CharSet      string
	WithTimeZone bool
}

func (t PgType) RenderTo(r Renderer) {
	if t.Setof {
		r.Text("setof", "keyword")
		r.Space()
	}

	t.Name.RenderTo(r)

	if t.OptInterval != nil {
		r.Space()
		t.OptInterval.RenderTo(r)
	}

	if t.ArrayWord {
		r.Space()
		r.Text("array", "keyword")
	}

	for _, ab := range t.ArrayBounds {
		r.Text("[", "lbracket")
		r.Text(string(ab), "IntegerConst")
		r.Text("]", "lbracket")
	}

	if len(t.TypeMods) > 0 {
		r.Text("(", "lparen")
		for i, e := range t.TypeMods {
			e.RenderTo(r)
			if i < len(t.TypeMods)-1 {
				r.Text(",", "comma")
				r.Space()
			}
		}
		r.Text(")", "rparen")
	}

	if t.WithTimeZone {
		r.Space()
		r.Text("with time zone", "keyword")
	}

	if t.CharSet != "" {
		r.Space()
		r.Text("character set", "keyword")
		r.Space()
		r.Text(t.CharSet, "charset")
	}
}

type AnyName []string

func (an AnyName) RenderTo(r Renderer) {
	for i, n := range an {
		r.Text(n, "identifer")
		if i < len(an)-1 {
			r.Text(".", "period")
		}
	}
}

type ColumnRef struct {
	Name        string
	Indirection Indirection
}

func (cr ColumnRef) RenderTo(r Renderer) {
	r.Text(cr.Name, "identifer")
	if cr.Indirection != nil {
		cr.Indirection.RenderTo(r)
	}
}

type Indirection []IndirectionEl

func (i Indirection) RenderTo(r Renderer) {
	for _, e := range i {
		e.RenderTo(r)
	}
}

type IndirectionEl struct {
	Name           string
	LowerSubscript Expr
	UpperSubscript Expr
}

func (ie IndirectionEl) RenderTo(r Renderer) {
	if ie.LowerSubscript != nil {
		r.Text("[", "lbracket")
		ie.LowerSubscript.RenderTo(r)
		if ie.UpperSubscript != nil {
			r.Text(":", "colon")
			ie.UpperSubscript.RenderTo(r)
		}
		r.Text("]", "rbracket")
	} else {
		r.Text(".", "period")
		r.Text(ie.Name, "identifier")
	}
}

type StringConst string

func (s StringConst) RenderTo(r Renderer) {
	r.Text(string(s), "StringConst")
}

type IntegerConst string

func (s IntegerConst) RenderTo(r Renderer) {
	r.Text(string(s), "IntegerConst")
}

type FloatConst string

func (s FloatConst) RenderTo(r Renderer) {
	r.Text(string(s), "floatConstant")
}

type BoolConst bool

func (b BoolConst) RenderTo(r Renderer) {
	if b {
		r.Text("true", "BoolConst")
	} else {
		r.Text("false", "BoolConst")
	}
}

type NullConst struct{}

func (n NullConst) RenderTo(r Renderer) {
	r.Text("null", "NullConst")
}

type BitConst string

func (b BitConst) RenderTo(r Renderer) {
	r.Text(string(b), "bitConstant")
}

type BooleanExpr struct {
	Left     Expr
	Operator string
	Right    Expr
}

func (e BooleanExpr) RenderTo(r Renderer) {
	e.Left.RenderTo(r)
	r.NewLine()
	r.Text(e.Operator, "operator")
	r.Space()
	e.Right.RenderTo(r)
}

type BinaryExpr struct {
	Left     Expr
	Operator AnyName
	Right    Expr
}

func (e BinaryExpr) RenderTo(r Renderer) {
	e.Left.RenderTo(r)
	r.Space()
	e.Operator.RenderTo(r)
	r.Space()
	e.Right.RenderTo(r)
}

type ArrayConstructorExpr ArrayExpr

func (ace ArrayConstructorExpr) RenderTo(r Renderer) {
	r.Text("array", "keyword")
	ArrayExpr(ace).RenderTo(r)
}

type ArrayExpr []Expr

func (a ArrayExpr) RenderTo(r Renderer) {
	r.Text("[", "lbracket")
	for i, e := range a {
		e.RenderTo(r)
		if i < len(a)-1 {
			r.Text(",", "comma")
			r.Space()
		}
	}
	r.Text("]", "rbracket")
}

type TextOpWithEscapeExpr struct {
	Left     Expr
	Operator string
	Right    Expr
	Escape   Expr
}

func (e TextOpWithEscapeExpr) RenderTo(r Renderer) {
	e.Left.RenderTo(r)
	r.Space()
	r.Text(e.Operator, "operator")
	r.Space()
	e.Right.RenderTo(r)

	if e.Escape != nil {
		r.Space()
		r.Text("escape", "keyword")
		r.Space()
		e.Escape.RenderTo(r)
	}
}

type UnaryExpr struct {
	Operator AnyName
	Expr     Expr
}

func (e UnaryExpr) RenderTo(r Renderer) {
	e.Operator.RenderTo(r)
	e.Expr.RenderTo(r)
}

type PostfixExpr struct {
	Expr     Expr
	Operator AnyName
}

func (e PostfixExpr) RenderTo(r Renderer) {
	e.Expr.RenderTo(r)
	r.Space()
	e.Operator.RenderTo(r)
}

type SubqueryOpExpr struct {
	Value Expr
	Op    SubqueryOp
	Type  string
	Query Expr
}

func (s SubqueryOpExpr) RenderTo(r Renderer) {
	s.Value.RenderTo(r)
	r.Space()
	s.Op.RenderTo(r)
	r.Space()
	r.Text(s.Type, "keyword")
	r.Space()
	s.Query.RenderTo(r)
}

type SubqueryOp struct {
	Operator bool
	Name     AnyName
}

func (s SubqueryOp) RenderTo(r Renderer) {
	if s.Operator {
		r.Text("operator", "keyword")
		r.Text("(", "lparen")
	}
	s.Name.RenderTo(r)
	if s.Operator {
		r.Text(")", "lparen")
	}
}

type WhenClause struct {
	When Expr
	Then Expr
}

func (w WhenClause) RenderTo(r Renderer) {
	r.Text("when", "keyword")
	r.Space()
	w.When.RenderTo(r)
	r.Space()
	r.Text("then", "keyword")
	r.NewLine()
	r.Indent()
	w.Then.RenderTo(r)
	r.NewLine()
	r.Unindent()
}

type InExpr struct {
	Value Expr
	Not   bool
	In    Expr
}

func (i InExpr) RenderTo(r Renderer) {
	i.Value.RenderTo(r)
	r.Space()

	if i.Not {
		r.Text("not", "keyword")
		r.Space()
	}

	r.Text("in", "keyword")
	r.Space()

	i.In.RenderTo(r)
}

type BetweenExpr struct {
	Expr      Expr
	Not       bool
	Symmetric bool
	Left      Expr
	Right     Expr
}

func (b BetweenExpr) RenderTo(r Renderer) {
	b.Expr.RenderTo(r)
	r.Space()

	if b.Not {
		r.Text("not", "keyword")
		r.Space()
	}

	r.Text("between", "keyword")
	r.Space()

	if b.Symmetric {
		r.Text("symmetric", "keyword")
		r.Space()
	}

	b.Left.RenderTo(r)
	r.Space()
	r.Text("and", "keyword")
	r.Space()
	b.Right.RenderTo(r)
}

type CaseExpr struct {
	CaseArg     Expr
	WhenClauses []WhenClause
	Default     Expr
}

func (c CaseExpr) RenderTo(r Renderer) {
	r.Text("case", "keyword")

	if c.CaseArg != nil {
		r.Space()
		c.CaseArg.RenderTo(r)
	}

	r.NewLine()

	for _, w := range c.WhenClauses {
		w.RenderTo(r)
	}

	if c.Default != nil {
		r.Text("else", "keyword")
		r.NewLine()
		r.Indent()
		c.Default.RenderTo(r)
		r.NewLine()
		r.Unindent()
	}

	r.Text("end", "keyword")
	r.NewLine()
}

type ParenExpr struct {
	Expr        Expr
	Indirection Indirection
}

func (e ParenExpr) RenderTo(r Renderer) {
	r.Text("(", "lparen")
	e.Expr.RenderTo(r)
	r.Text(")", "rparen")
	if e.Indirection != nil {
		e.Indirection.RenderTo(r)
	}
}

type TypecastExpr struct {
	Expr     Expr
	Typename PgType
}

func (t TypecastExpr) RenderTo(r Renderer) {
	t.Expr.RenderTo(r)
	r.Text("::", "typecast")
	t.Typename.RenderTo(r)
}

type ConstTypeExpr struct {
	Typename PgType
	Expr     Expr
}

func (t ConstTypeExpr) RenderTo(r Renderer) {
	t.Typename.RenderTo(r)
	r.Space()
	t.Expr.RenderTo(r)
}

type ConstIntervalExpr struct {
	Precision   IntegerConst
	Value       Expr
	OptInterval *OptInterval
}

func (i ConstIntervalExpr) RenderTo(r Renderer) {
	r.Text("interval", "keyword")
	if i.Precision != "" {
		r.Text("(", "lparen")
		i.Precision.RenderTo(r)
		r.Text(")", "lparen")
	}

	r.Space()
	i.Value.RenderTo(r)

	if i.OptInterval != nil {
		r.Space()
		i.OptInterval.RenderTo(r)
	}
}

type OptInterval struct {
	Left   string
	Right  string
	Second *IntervalSecond
}

func (oi OptInterval) RenderTo(r Renderer) {
	if oi.Left != "" {
		r.Text(oi.Left, "keyword")
	}

	if oi.Right != "" {
		r.Space()
		r.Text("to", "keyword")
		r.Space()
		r.Text(oi.Right, "keyword")
	}

	if oi.Second != nil {
		if oi.Left != "" {
			r.Space()
		}
		oi.Second.RenderTo(r)
	}
}

type IntervalSecond struct {
	Precision IntegerConst
}

func (is IntervalSecond) RenderTo(r Renderer) {
	r.Text("second", "keyword")
	if is.Precision != "" {
		r.Text("(", "lparen")
		is.Precision.RenderTo(r)
		r.Text(")", "rparen")
	}
}

type ExtractExpr ExtractList

func (ee ExtractExpr) RenderTo(r Renderer) {
	r.Text("extract", "keyword")
	r.Text("(", "lparen")
	ExtractList(ee).RenderTo(r)
	r.Text(")", "rparen")
}

type ExtractList struct {
	Extract Expr
	Time    Expr
}

func (el ExtractList) RenderTo(r Renderer) {
	el.Extract.RenderTo(r)
	r.Space()
	r.Text("from", "keyword")
	r.Space()
	el.Time.RenderTo(r)
}

type OverlayExpr OverlayList

func (oe OverlayExpr) RenderTo(r Renderer) {
	r.Text("overlay", "keyword")
	r.Text("(", "lparen")
	OverlayList(oe).RenderTo(r)
	r.Text(")", "rparen")
}

type OverlayList struct {
	Dest    Expr
	Placing Expr
	From    Expr
	For     Expr
}

func (ol OverlayList) RenderTo(r Renderer) {
	ol.Dest.RenderTo(r)
	r.Space()
	r.Text("placing", "keyword")
	r.Space()
	ol.Placing.RenderTo(r)
	r.Space()
	r.Text("from", "keyword")
	r.Space()
	ol.From.RenderTo(r)

	if ol.For != nil {
		r.Space()
		r.Text("for", "keyword")
		r.Space()
		ol.For.RenderTo(r)
	}
}

type PositionExpr PositionList

func (pe PositionExpr) RenderTo(r Renderer) {
	r.Text("position", "keyword")
	r.Text("(", "lparen")
	PositionList(pe).RenderTo(r)
	r.Text(")", "rparen")
}

type PositionList struct {
	Substring Expr
	String    Expr
}

func (pl PositionList) RenderTo(r Renderer) {
	pl.Substring.RenderTo(r)
	r.Space()
	r.Text("in", "keyword")
	r.Space()
	pl.String.RenderTo(r)
}

type SubstrExpr SubstrList

func (se SubstrExpr) RenderTo(r Renderer) {
	r.Text("substring", "keyword")
	r.Text("(", "lparen")
	SubstrList(se).RenderTo(r)
	r.Text(")", "rparen")
}

type SubstrList struct {
	Source Expr
	From   Expr
	For    Expr
}

func (sl SubstrList) RenderTo(r Renderer) {
	sl.Source.RenderTo(r)
	r.Space()
	r.Text("from", "keyword")
	r.Space()
	sl.From.RenderTo(r)

	if sl.For != nil {
		r.Space()
		r.Text("for", "keyword")
		r.Space()
		sl.For.RenderTo(r)
	}
}

type TrimExpr struct {
	Direction string
	TrimList
}

func (te TrimExpr) RenderTo(r Renderer) {
	r.Text("trim", "keyword")
	r.Text("(", "lparen")
	if te.Direction != "" {
		r.Text(te.Direction, "keyword")
		r.Space()
	}
	te.TrimList.RenderTo(r)
	r.Text(")", "rparen")
}

type TrimList struct {
	Left  Expr
	From  bool
	Right []Expr
}

func (tl TrimList) RenderTo(r Renderer) {
	if tl.Left != nil {
		tl.Left.RenderTo(r)
		r.Space()
	}

	if tl.From {
		r.Text("from", "keyword")
		r.Space()
	}

	for i, e := range tl.Right {
		e.RenderTo(r)
		if i+1 < len(tl.Right) {
			r.Text(",", "comma")
			r.Space()
		}
	}
}

type XmlElement struct {
	Name       string
	Attributes XmlAttributes
	Body       []Expr
}

func (el XmlElement) RenderTo(r Renderer) {
	r.Text("xmlelement", "keyword")
	r.Text("(", "lparen")
	r.Text("name", "keyword")
	r.Space()
	r.Text(el.Name, "identifier")

	if el.Attributes != nil {
		r.Text(",", "comma")
		r.Space()
		el.Attributes.RenderTo(r)
	}

	if el.Body != nil {
		for _, e := range el.Body {
			r.Text(",", "comma")
			r.Space()
			e.RenderTo(r)
		}
	}

	r.Text(")", "rparen")
}

type XmlAttributes []XmlAttributeEl

func (attrs XmlAttributes) RenderTo(r Renderer) {
	r.Text("xmlattributes", "keyword")
	r.Text("(", "lparen")
	xmlAttributes(attrs).RenderTo(r)
	r.Text(")", "rparen")
}

type XmlAttributeEl struct {
	Value Expr
	Name  string
}

func (el XmlAttributeEl) RenderTo(r Renderer) {
	el.Value.RenderTo(r)
	if el.Name != "" {
		r.Space()
		r.Text("as", "keyword")
		r.Space()
		r.Text(el.Name, "identifier")
	}
}

type XmlExists struct {
	Path Expr
	Body XmlExistsArgument
}

func (e XmlExists) RenderTo(r Renderer) {
	r.Text("xmlexists", "keyword")
	r.Text("(", "lparen")
	e.Path.RenderTo(r)
	r.Space()
	e.Body.RenderTo(r)
	r.Text(")", "rparen")
}

type XmlExistsArgument struct {
	LeftByRef  bool
	Arg        Expr
	RightByRef bool
}

func (a XmlExistsArgument) RenderTo(r Renderer) {
	r.Text("passing", "keyword")
	r.Space()

	if a.LeftByRef {
		r.Text("by ref", "keyword")
		r.Space()
	}

	a.Arg.RenderTo(r)

	if a.RightByRef {
		r.Space()
		r.Text("by ref", "keyword")
	}
}

type XmlForest []XmlAttributeEl

func (f XmlForest) RenderTo(r Renderer) {
	r.Text("xmlforest", "keyword")
	r.Text("(", "lparen")
	xmlAttributes(f).RenderTo(r)
	r.Text(")", "rparen")
}

type xmlAttributes []XmlAttributeEl

func (attrs xmlAttributes) RenderTo(r Renderer) {
	for i, a := range attrs {
		a.RenderTo(r)
		if i+1 < len(attrs) {
			r.Text(",", "comma")
			r.Space()
		}
	}
}

type XmlParse struct {
	Type             string
	Content          Expr
	WhitespaceOption string
}

func (p XmlParse) RenderTo(r Renderer) {
	r.Text("xmlparse", "keyword")
	r.Text("(", "lparen")
	r.Text(p.Type, "keyword")
	r.Space()
	p.Content.RenderTo(r)
	if p.WhitespaceOption != "" {
		r.Space()
		r.Text(p.WhitespaceOption, "keyword")
	}

	r.Text(")", "rparen")
}

type XmlPi struct {
	Name    string
	Content Expr
}

func (p XmlPi) RenderTo(r Renderer) {
	r.Text("xmlpi", "keyword")
	r.Text("(", "lparen")
	r.Text("name", "keyword")
	r.Space()
	r.Text(p.Name, "identifier")

	if p.Content != nil {
		r.Text(",", "comma")
		r.Space()
		p.Content.RenderTo(r)
	}
	r.Text(")", "rparen")
}

type XmlRoot struct {
	Xml        Expr
	Version    XmlRootVersion
	Standalone string
}

func (x XmlRoot) RenderTo(r Renderer) {
	r.Text("xmlroot", "keyword")
	r.Text("(", "lparen")
	x.Xml.RenderTo(r)
	r.Text(",", "comma")
	r.Space()
	x.Version.RenderTo(r)
	if x.Standalone != "" {
		r.Text(",", "comma")
		r.Space()
		r.Text("standalone", "keyword")
		r.Space()
		r.Text(x.Standalone, "keyword")
	}
	r.Text(")", "rparen")
}

type XmlRootVersion struct {
	Expr Expr
}

func (rv XmlRootVersion) RenderTo(r Renderer) {
	r.Text("version", "keyword")
	r.Space()
	if rv.Expr != nil {
		rv.Expr.RenderTo(r)
	} else {
		r.Text("no value", "keyword")
	}
}

type XmlSerialize struct {
	XmlType string
	Content Expr
	Type    PgType
}

func (s XmlSerialize) RenderTo(r Renderer) {
	r.Text("xmlserialize", "keyword")
	r.Text("(", "lparen")
	r.Text(s.XmlType, "keyword")
	r.Space()
	s.Content.RenderTo(r)
	r.Space()
	r.Text("as", "keyword")
	r.Space()
	s.Type.RenderTo(r)
	r.Text(")", "rparen")
}

type CollateExpr struct {
	Expr      Expr
	Collation AnyName
}

func (c CollateExpr) RenderTo(r Renderer) {
	c.Expr.RenderTo(r)
	r.Space()
	r.Text("collate", "keyword")
	r.Space()
	c.Collation.RenderTo(r)
}

type NotExpr struct {
	Expr Expr
}

func (e NotExpr) RenderTo(r Renderer) {
	r.Text("not", "keyword")
	r.Space()
	e.Expr.RenderTo(r)
}

type IsExpr struct {
	Expr Expr
	Not  bool
	Op   string // null, document, true, false, etc.
}

func (e IsExpr) RenderTo(r Renderer) {
	e.Expr.RenderTo(r)
	r.Space()
	r.Text("is", "keyword")
	r.Space()
	if e.Not {
		r.Text("not", "keyword")
		r.Space()
	}
	r.Text(e.Op, "keyword")
}

type AliasedExpr struct {
	Expr  Expr
	Alias string
}

func (e AliasedExpr) RenderTo(r Renderer) {
	e.Expr.RenderTo(r)
	r.Space()
	r.Text("as", "keyword")
	r.Space()
	r.Text(e.Alias, "identifier")
}

type IntoClause struct {
	Options  string
	OptTable bool
	Target   AnyName
}

func (i IntoClause) RenderTo(r Renderer) {
	r.Text("into", "keyword")
	r.Space()

	if i.Options != "" {
		r.Text(i.Options, "keyword")
		r.Space()
	}

	if i.OptTable {
		r.Text("table", "keyword")
		r.Space()
	}

	i.Target.RenderTo(r)
	r.NewLine()
}

type FromClause struct {
	Expr Expr
}

func (e FromClause) RenderTo(r Renderer) {
	r.Text("from", "keyword")
	r.NewLine()
	r.Indent()
	e.Expr.RenderTo(r)
	r.NewLine()
	r.Unindent()
}

type JoinExpr struct {
	Left  Expr
	Join  string
	Right Expr
	Using []string
	On    Expr
}

func (s JoinExpr) RenderTo(r Renderer) {
	s.Left.RenderTo(r)

	if s.Join == "," {
		r.Text(",", "comma")
		r.NewLine()
	} else {
		r.NewLine()
		r.Text(s.Join, "keyword")
		r.Space()
	}

	s.Right.RenderTo(r)

	if len(s.Using) > 0 {
		r.Space()
		r.Text("using", "keyword")
		r.Text("(", "lparen")

		for i, u := range s.Using {
			r.Text(u, "identifier")
			if i+1 < len(s.Using) {
				r.Text(",", "comma")
				r.Space()
			}
		}

		r.Text(")", "rparen")
	}

	if s.On != nil {
		r.Space()
		r.Text("on", "keyword")
		r.Space()
		s.On.RenderTo(r)
	}
}

type WhereClause struct {
	Expr Expr
}

func (e WhereClause) RenderTo(r Renderer) {
	r.Text("where", "keyword")
	r.NewLine()
	r.Indent()
	e.Expr.RenderTo(r)
	r.NewLine()
	r.Unindent()
}

type OrderExpr struct {
	Expr  Expr
	Order string
	Using AnyName
	Nulls string
}

func (e OrderExpr) RenderTo(r Renderer) {
	e.Expr.RenderTo(r)
	if e.Order != "" {
		r.Space()
		r.Text(e.Order, "keyword")
	}
	if len(e.Using) > 0 {
		r.Space()
		r.Text("using", "keyword")
		r.Space()
		e.Using.RenderTo(r)
	}
	if e.Nulls != "" {
		r.Space()
		r.Text("nulls", "keyword")
		r.Space()
		r.Text(e.Nulls, "keyword")
	}
}

type OrderClause struct {
	Exprs []OrderExpr
}

func (e OrderClause) RenderTo(r Renderer) {
	r.Text("order by", "keyword")
	r.NewLine()
	r.Indent()

	for i, f := range e.Exprs {
		f.RenderTo(r)
		if i < len(e.Exprs)-1 {
			r.Text(",", "comma")
		}
		r.NewLine()
	}
	r.Unindent()
}

type GroupByClause struct {
	Exprs []Expr
}

func (e GroupByClause) RenderTo(r Renderer) {
	r.Text("group by", "keyword")
	r.NewLine()
	r.Indent()

	for i, f := range e.Exprs {
		f.RenderTo(r)
		if i < len(e.Exprs)-1 {
			r.Text(",", "comma")
		}
		r.NewLine()
	}
	r.Unindent()
}

type LimitClause struct {
	Limit  Expr
	Offset Expr
}

func (e LimitClause) RenderTo(r Renderer) {
	if e.Limit != nil {
		r.Text("limit", "keyword")
		r.Space()
		e.Limit.RenderTo(r)
		r.NewLine()
	}
	if e.Offset != nil {
		r.Text("offset", "keyword")
		r.Space()
		e.Offset.RenderTo(r)
		r.NewLine()
	}
}

type AtTimeZoneExpr struct {
	Expr     Expr
	TimeZone Expr
}

func (e AtTimeZoneExpr) RenderTo(r Renderer) {
	e.Expr.RenderTo(r)
	r.Space()
	r.Text("at time zone", "keyword")
	r.Space()
	e.TimeZone.RenderTo(r)
}

type LockingItem struct {
	Strength   string
	LockedRels []AnyName
	WaitPolicy string
}

func (li LockingItem) RenderTo(r Renderer) {
	r.Text("for", "keyword")
	r.Space()
	r.Text(li.Strength, "keyword")

	if li.LockedRels != nil {
		r.Space()
		r.Text("of", "keyword")
		r.Space()

		for i, lr := range li.LockedRels {
			lr.RenderTo(r)
			if i < len(li.LockedRels)-1 {
				r.Text(",", "comma")
				r.Space()
			}
		}
	}

	if li.WaitPolicy != "" {
		r.Space()
		r.Text(li.WaitPolicy, "keyword")
	}

	r.NewLine()
}

type LockingClause struct {
	Locks []LockingItem
}

func (lc LockingClause) RenderTo(r Renderer) {
	for _, li := range lc.Locks {
		li.RenderTo(r)
	}
}

type FuncExprNoParens string

func (fe FuncExprNoParens) RenderTo(r Renderer) {
	r.Text(string(fe), "keyword")
}

type FuncExpr struct {
	FuncApplication
	WithinGroupClause *WithinGroupClause
	FilterClause      *FilterClause
	OverClause        *OverClause
}

func (fe FuncExpr) RenderTo(r Renderer) {
	fe.FuncApplication.RenderTo(r)

	if fe.WithinGroupClause != nil {
		r.Space()
		fe.WithinGroupClause.RenderTo(r)
	}

	if fe.FilterClause != nil {
		r.Space()
		fe.FilterClause.RenderTo(r)
	}

	if fe.OverClause != nil {
		r.Space()
		fe.OverClause.RenderTo(r)
	}
}

type FuncApplication struct {
	Name AnyName

	Distinct bool

	Star        bool
	Args        []FuncArg
	VariadicArg *FuncArg

	OrderClause *OrderClause
}

func (fa FuncApplication) RenderTo(r Renderer) {
	fa.Name.RenderTo(r)
	r.Text("(", "lparen")

	if fa.Distinct {
		r.Text("distinct", "keyword")
		r.Space()
	}

	if fa.Star {
		r.Text("*", "star")
	} else if len(fa.Args) > 0 {
		for i, a := range fa.Args {
			a.RenderTo(r)
			if i < len(fa.Args)-1 {
				r.Text(",", "comma")
				r.Space()
			}
		}
	}

	if fa.VariadicArg != nil {
		if len(fa.Args) > 0 {
			r.Text(",", "comma")
			r.Space()
		}

		r.Text("variadic", "keyword")
		r.Space()
		fa.VariadicArg.RenderTo(r)
	}

	if fa.OrderClause != nil {
		r.Space()
		fa.OrderClause.RenderTo(r)
	}

	r.Text(")", "lparen")
}

type FuncArg struct {
	Name   string
	NameOp string
	Expr   Expr
}

func (fa FuncArg) RenderTo(r Renderer) {
	if fa.Name != "" {
		r.Text(fa.Name, "identifier")
		r.Space()
		r.Text(fa.NameOp, "operator")
		r.Space()
	}
	fa.Expr.RenderTo(r)
}

type CastFunc struct {
	Name string
	Expr Expr
	Type PgType
}

func (cf CastFunc) RenderTo(r Renderer) {
	r.Text(cf.Name, "keyword")
	r.Text("(", "lparen")
	cf.Expr.RenderTo(r)
	r.Space()
	r.Text("as", "keyword")
	r.Space()
	cf.Type.RenderTo(r)
	r.Text(")", "rparen")
}

type IsOfExpr struct {
	Expr  Expr
	Not   bool
	Types []PgType
}

func (io IsOfExpr) RenderTo(r Renderer) {
	io.Expr.RenderTo(r)
	r.Space()
	r.Text("is", "keyword")
	r.Space()

	if io.Not {
		r.Text("not", "keyword")
		r.Space()
	}

	r.Text("of", "keyword")
	r.Space()
	r.Text("(", "lparen")

	for i, t := range io.Types {
		t.RenderTo(r)

		if i < len(io.Types)-1 {
			r.Text(",", "comma")
			r.Space()
		}
	}

	r.Text(")", "rparen")
}

type WithinGroupClause OrderClause

func (w WithinGroupClause) RenderTo(r Renderer) {
	r.Text("within group", "keyword")
	r.Space()
	r.Text("(", "lparen")
	OrderClause(w).RenderTo(r)
	r.Text(")", "rparen")
}

type FilterClause struct {
	Expr
}

func (f FilterClause) RenderTo(r Renderer) {
	r.Text("filter", "keyword")
	r.Space()
	r.Text("(", "lparen")
	r.Text("where", "keyword")
	r.Space()
	f.Expr.RenderTo(r)
	r.Text(")", "rparen")
}

type DefaultExpr bool

func (d DefaultExpr) RenderTo(r Renderer) {
	r.Text("default", "keyword")
}

type Row struct {
	RowWord bool
	Exprs   []Expr
}

func (row Row) RenderTo(r Renderer) {
	if row.RowWord {
		r.Text("row", "keyword")
		r.Space()
	}

	r.Text("(", "lparen")

	for i, e := range row.Exprs {
		e.RenderTo(r)
		if i < len(row.Exprs)-1 {
			r.Text(",", "comma")
			r.Space()
		}
	}

	r.Text(")", "rparen")
}

type ValuesRow []Expr

func (vr ValuesRow) RenderTo(r Renderer) {
	r.Text("(", "lparen")

	for i, e := range vr {
		e.RenderTo(r)
		if i < len(vr)-1 {
			r.Text(",", "comma")
			r.Space()
		}
	}

	r.Text(")", "rparen")
}

type ValuesClause []ValuesRow

func (vc ValuesClause) RenderTo(r Renderer) {
	r.Text("values", "keyword")
	r.NewLine()
	r.Indent()

	for i, row := range vc {
		row.RenderTo(r)
		if i < len(vc)-1 {
			r.Text(",", "comma")
		}
		r.NewLine()
	}

	r.Unindent()
}

type OverClause struct {
	Name          string
	Specification *WindowSpecification
}

func (oc *OverClause) RenderTo(r Renderer) {
	r.Text("over", "keyword")
	r.Space()
	if oc.Name != "" {
		r.Text(oc.Name, "identifier")
	} else {
		oc.Specification.RenderTo(r)
	}
}

type WindowClause []WindowDefinition

func (wc WindowClause) RenderTo(r Renderer) {
	r.Text("window", "keyword")
	r.NewLine()
	r.Indent()

	for i, wd := range wc {
		wd.RenderTo(r)
		if i < len(wc)-1 {
			r.Text(",", "comma")
		}
		r.NewLine()
	}

	r.Unindent()
}

type WindowDefinition struct {
	Name          string
	Specification WindowSpecification
}

func (wd WindowDefinition) RenderTo(r Renderer) {
	r.Text(wd.Name, "identifier")
	r.Space()
	r.Text("as", "keyword")
	r.Space()
	wd.Specification.RenderTo(r)
}

type WindowSpecification struct {
	ExistingName    string
	PartitionClause PartitionClause
	OrderClause     *OrderClause
	FrameClause     *FrameClause
}

func (ws WindowSpecification) RenderTo(r Renderer) {
	r.Text("(", "lparen")

	if ws.ExistingName != "" {
		r.Text(ws.ExistingName, "identifier")
		r.Space()
	}

	if ws.PartitionClause != nil {
		ws.PartitionClause.RenderTo(r)

		// TODO figure better way to handle spaces
		if ws.OrderClause != nil || ws.FrameClause != nil {
			r.Space()
		}
	}

	if ws.OrderClause != nil {
		ws.OrderClause.RenderTo(r)
		if ws.FrameClause != nil {
			r.Space()
		}
	}

	if ws.FrameClause != nil {
		ws.FrameClause.RenderTo(r)
	}

	r.Text(")", "rparen")
}

type PartitionClause []Expr

func (pc PartitionClause) RenderTo(r Renderer) {
	r.Text("partition by", "keyword")
	r.Space()

	for i, e := range pc {
		e.RenderTo(r)
		if i < len(pc)-1 {
			r.Text(",", "comma")
			r.Space()
		}
	}
}

type FrameClause struct {
	Mode  string
	Start *FrameBound
	End   *FrameBound
}

func (fc *FrameClause) RenderTo(r Renderer) {
	r.Text(fc.Mode, "keyword")
	r.Space()

	if fc.End != nil {
		r.Text("between", "keyword")
		r.Space()
		fc.Start.RenderTo(r)
		r.Space()
		r.Text("and", "keyword")
		r.Space()
		fc.End.RenderTo(r)
	} else {
		fc.Start.RenderTo(r)
	}
}

type FrameBound struct {
	CurrentRow bool

	BoundExpr Expr
	Direction string
}

func (fb FrameBound) RenderTo(r Renderer) {
	if fb.CurrentRow {
		r.Text("current row", "keyword")
		return
	}

	if fb.BoundExpr != nil {
		fb.BoundExpr.RenderTo(r)
	} else {
		r.Text("unbounded", "keyword")
	}

	r.Space()

	r.Text(fb.Direction, "keyword")
}

type RelationExpr struct {
	Name AnyName
	Star bool
	Only bool
}

func (re RelationExpr) RenderTo(r Renderer) {
	if re.Only {
		r.Text("only", "keyword")
		r.Space()
	}

	re.Name.RenderTo(r)

	if re.Star {
		r.Space()
		r.Text("*", "star")
	}

	r.NewLine()
}

type SimpleSelect struct {
	DistinctList  []Expr
	TargetList    []Expr
	IntoClause    *IntoClause
	FromClause    *FromClause
	WhereClause   *WhereClause
	GroupByClause *GroupByClause
	HavingClause  Expr
	WindowClause  WindowClause

	ValuesClause ValuesClause

	LeftSelect  *SelectStmt
	SetOp       string
	SetAll      bool
	RightSelect *SelectStmt

	Table *RelationExpr
}

func (s SimpleSelect) RenderTo(r Renderer) {
	if s.Table != nil {
		r.Text("table", "keyword")
		r.Space()
		s.Table.RenderTo(r)
		return
	}

	if s.ValuesClause != nil {
		s.ValuesClause.RenderTo(r)
		return
	}

	if s.LeftSelect != nil {
		s.LeftSelect.RenderTo(r)
		r.NewLine()
		r.Text(s.SetOp, "keyword")

		if s.SetAll {
			r.Space()
			r.Text("all", "keyword")
		}

		r.NewLine()

		s.RightSelect.RenderTo(r)

		return
	}

	r.Text("select", "keyword")

	if s.DistinctList != nil {
		r.Space()
		r.Text("distinct", "keyword")

		if len(s.DistinctList) > 0 {
			r.Space()
			r.Text("on", "keyword")
			r.Text("(", "lparen")

			for i, f := range s.DistinctList {
				f.RenderTo(r)
				if i < len(s.DistinctList)-1 {
					r.Text(",", "comma")
					r.Space()
				}
			}
			r.Text(")", "rparen")
		}

	}

	r.NewLine()
	r.Indent()
	for i, f := range s.TargetList {
		f.RenderTo(r)
		if i < len(s.TargetList)-1 {
			r.Text(",", "comma")
		}
		r.NewLine()
	}
	r.Unindent()

	if s.IntoClause != nil {
		s.IntoClause.RenderTo(r)
	}

	if s.FromClause != nil {
		s.FromClause.RenderTo(r)
	}

	if s.WhereClause != nil {
		s.WhereClause.RenderTo(r)
	}

	if s.GroupByClause != nil {
		s.GroupByClause.RenderTo(r)
	}

	if s.HavingClause != nil {
		r.Text("having", "keyword")
		r.NewLine()
		r.Indent()
		s.HavingClause.RenderTo(r)
		r.NewLine()
	}

	if s.WindowClause != nil {
		s.WindowClause.RenderTo(r)
	}
}

type SelectStmt struct {
	SimpleSelect
	OrderClause   *OrderClause
	LimitClause   *LimitClause
	LockingClause *LockingClause

	ParenWrapped bool
	Semicolon    bool
}

func (s SelectStmt) RenderTo(r Renderer) {
	if s.ParenWrapped {
		r.Text("(", "lparen")
	}

	s.SimpleSelect.RenderTo(r)

	if s.OrderClause != nil {
		s.OrderClause.RenderTo(r)
	}

	if s.LimitClause != nil {
		s.LimitClause.RenderTo(r)
	}

	if s.LockingClause != nil {
		s.LockingClause.RenderTo(r)
	}

	if s.ParenWrapped {
		r.Text(")", "rparen")
		r.NewLine()
	}

	if s.Semicolon {
		r.Text(";", "semicolon")
		r.NewLine()
	}
}

type ExistsExpr SelectStmt

func (e ExistsExpr) RenderTo(r Renderer) {
	r.Text("exists", "keyword")

	SelectStmt(e).RenderTo(r)
}

type ArraySubselect SelectStmt

func (a ArraySubselect) RenderTo(r Renderer) {
	r.Text("array", "keyword")

	SelectStmt(a).RenderTo(r)
}
