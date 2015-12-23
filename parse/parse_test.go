package parse

import (
	"testing"
)

type parseTest struct {
	name     string
	input    string
	expected *ModuleNode
	err      string
}

const noError = ""

// Position testing isnt implemented
var noPos = pos{0, 0}

func newParseTest(name, input string, expected *ModuleNode) parseTest {
	return parseTest{name, input, expected, noError}
}

func newErrorTest(name, input string, err string) parseTest {
	return parseTest{name, input, mkModule(), err}
}

func mkModule(nodes ...Node) *ModuleNode {
	l := newModuleNode()
	for _, n := range nodes {
		l.append(n)
	}

	return l
}

var parseTests = []parseTest{
	// Errors
	newErrorTest("unclosed block", "{% block test %}", "parse error: unclosed tag \"block\" starting on line 1, offset 3"),
	newErrorTest("unclosed if", "{% if test %}", "parse error: unclosed tag \"if\" starting on line 1, offset 3"),
	newErrorTest("unexpected end (function call)", "{{ func('arg1'", "parse error: unexpected end of input on line 1, offset 14"),
	newErrorTest("unclosed parenthesis", "{{ func(arg1 }}", "parse error: expected one of [PUNCTUATION, PARENS_CLOSE], got \"PRINT_CLOSE\" on line 1, offset 13"),
	newErrorTest("unexpected punctuation", "{{ func(arg1. arg2) }}", "parse error: unexpected punctuation \".\", expected \",\" on line 1, offset 12"),

	// Valid
	newParseTest("text", "some text", mkModule(newTextNode("some text", noPos))),
	newParseTest("hello", "Hello {{ name }}", mkModule(newTextNode("Hello ", noPos), newPrintNode(newNameExpr("name", noPos), noPos))),
	newParseTest("string expr", "Hello {{ 'Tyler' }}", mkModule(newTextNode("Hello ", noPos), newPrintNode(newStringExpr("Tyler", noPos), noPos))),
	newParseTest(
		"simple tag",
		"{% block something %}Body{% endblock %}",
		mkModule(newBlockNode("something", mkModule(newTextNode("Body", noPos)), noPos)),
	),
	newParseTest(
		"if",
		"{% if something %}Do Something{% endif %}",
		mkModule(newIfNode(newNameExpr("something", noPos), mkModule(newTextNode("Do Something", noPos)), mkModule(), noPos)),
	),
	newParseTest(
		"if else",
		"{% if something %}Do Something{% else %}Another thing{% endif %}",
		mkModule(newIfNode(newNameExpr("something", noPos), mkModule(newTextNode("Do Something", noPos)), mkModule(newTextNode("Another thing", noPos)), noPos)),
	),
	newParseTest(
		"if else if",
		"{% if something %}Do Something{% else if another %}Another thing{% endif %}",
		mkModule(newIfNode(
			newNameExpr("something", noPos),
			mkModule(newTextNode("Do Something", noPos)),
			mkModule(newIfNode(
				newNameExpr("another", noPos),
				mkModule(newTextNode("Another thing", noPos)),
				mkModule(),
				noPos)),
			noPos)),
	),
	newParseTest(
		"function expr",
		"{{ func('arg1', arg2) }}",
		mkModule(newPrintNode(newFuncExpr(newNameExpr("func", noPos), []Expr{newStringExpr("arg1", noPos), newNameExpr("arg2", noPos)}, noPos), noPos)),
	),
}

func nodeEqual(a, b Node) bool {
	if a.String() != b.String() {
		return false
	}

	return true
}

func evaluateTest(t *testing.T, test parseTest) {
	tree, err := Parse(test.input)
	if test.err != noError && err != nil && test.err != err.Error() {
		t.Errorf("%s: got error\n\t%+v\nexpected error\n\t%v", test.name, err, test.err)
	} else if !nodeEqual(tree.root, test.expected) {
		t.Errorf("%s: got\n\t%+v\nexpected\n\t%v", test.name, tree.root, test.expected)
		if err != nil {
			t.Errorf("%s: got error\n\t%v", test.name, err.Error())
		}
	}
}

func TestParse(t *testing.T) {
	for _, test := range parseTests {
		evaluateTest(t, test)
	}
}
