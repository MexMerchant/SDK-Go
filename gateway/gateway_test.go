package gateway

import (
	"fmt"
	"net/url"
	"strings"
	"testing"
)

func TestSigning(t *testing.T) {
	simpleFields := url.Values{`b`: {`two`}, `a`: {`one`}}
	if pass, err := signAndTest(simpleFields, `86cdc`); !pass {
		t.Error(err)
	}

	newLines := url.Values{`a`: {`one`}, `b`: {`New lines! %0D %0D%0A`}}
	if pass, err := signAndTest(newLines, `cf50d`); !pass {
		t.Error(err)
	}

	strangeSymbols := url.Values{`a`: {`one`}, `b`: {`strange "'?& symbols `}}
	if pass, err := signAndTest(strangeSymbols, `7c952`); !pass {
		t.Error(err)
	}

	fields := url.Values{`a`: {`one`}, `b`: {`a £ sign`}}
	if pass, err := signAndTest(fields, `13637`); !pass {
		t.Error(err)
	}

	fields = url.Values{`a`: {`one`}, `b`: {"newline \n characater"}}
	if pass, err := signAndTest(fields, `19582`); !pass {
		t.Error(err)
	}

	fields = url.Values{`a[aa]`: {`12`}, `a[bb]`: {`13`}, `a1`: {`0`}, `aa`: {`1`}, `aZ`: {`2`}}
	if pass, err := signAndTest(fields, `4aeaa`); !pass {
		t.Error(err)
	}
}

func signAndTest(fields url.Values, expected string) (bool, string) {
	g := NewGateway("http://test.3ds-pit.com/direct/", "100856", "Threeds2Test60System")
	sig := g.Sign(fields, `pass`)

	if !strings.HasPrefix(sig, expected) {
		return false, fmt.Sprintf("Signature Error: %s", sig)
	}

	return true, ``
}

func TestFieldToHTML(t *testing.T) {

	fields := map[string]string{
		`name`:        `John Smith`,
		`age`:         `42`,
		`address`:     `Somewhere over the rainbow`,
		`randomChars`: `A&nd st*r qu\'ote double""quote excla!m close}brace d-ash pl+us s/ash and backs\ash and semi; finally h#ash`,
	}

	var res = fieldsToHTML("main", fields)

	pass := strings.Contains(res, `name="main[age]" value="42"`)
	if !pass {
		t.Error("unable to validate fieldsToHTML")
	}

	pass = strings.Contains(res, `name="main[address]" value="Somewhere over the rainbow"`)
	if !pass {
		t.Error("unable to validate fieldsToHTML")
	}

	pass = strings.Contains(res, `"main[randomChars]" value="A&amp;nd st*r qu\&#39;ote double&#34;&#34;quote excla!m close}brace d-ash pl+us s/ash and backs\ash and semi; finally h#ash"`)
	if !pass {
		t.Error("unable to validate fieldsToHTML")
	}

	pass = strings.Contains(res, `name="main[name]" value="John Smith"`)
	if !pass {
		t.Error("unable to validate fieldsToHTML")
	}
}
