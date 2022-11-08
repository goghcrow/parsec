package example

import (
	"fmt"
	"testing"

	. "github.com/goghcrow/parsec"
	. "github.com/goghcrow/parsec/states/charstate"
)

func TestComment(t *testing.T) {
	comment := Right(Str("<!--"), ManyTill(Regex(`[\w\s]+`), Str("-->")))
	v, err := comment.Parse(NewState("<!--hello world-->"))
	if err != nil {
		panic(err)
	}
	expect := "[hello world]"
	actual := fmt.Sprintf("%s", v)
	if expect != actual {
		t.Errorf("expect %s actual %s", expect, actual)
	}
}
