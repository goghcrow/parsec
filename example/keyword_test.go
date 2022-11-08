package example

import (
	"testing"

	. "github.com/goghcrow/parsec"
	. "github.com/goghcrow/parsec/states/charstate"
)

func TestKeyword(t *testing.T) {
	let := Left(Str("let"), NotFollowedBy(Regex(`[\d\w]+`)))
	parse, err := let.Parse(NewState("let a = 1"))
	if err != nil {
		panic(err)
	}
	if parse.(string) != "let" {
		t.Errorf("expect let actual %s", parse)
	}

	parse, err = let.Parse(NewState("lets go"))
	if parse != nil {
		t.Errorf("expect error actual %s", parse)
	}
	expect := "unexpect `s` in pos 4 line 1 col 4"
	actual := err.Error()
	if actual != expect {
		t.Errorf("expect error %s actual %s", expect, actual)
	}
}
