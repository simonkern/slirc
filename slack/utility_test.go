package slack

import (
	"html"
	"testing"
)

func TestUnslackify(t *testing.T) {
	sc := setup(t)
	// all of these should be known
	raw1 := "<@U11A2B8C1> foobars &lt; &gt; <http://www.test.com> or <http://www.test.com|test.com> in <#C03JAPEHJ> for <mailto:test@example.com|test@example.com>"
	want1 := "tester1 foobars < > http://www.test.com or http://www.test.com in #dev for test@example.com"

	// contains unknown IDs
	raw2 := "<@U11A2B8C3> foobars &lt; &gt; <http://www.test.com> or <http://www.test.com|test.com> in <#C03JAPAAA> for <mailto:test@example.com|test@example.com>"
	want2 := "@U11A2B8C3 foobars < > http://www.test.com or http://www.test.com in #C03JAPAAA for test@example.com"

	// contains angle brackets but nothing special
	raw3 := "<fooBarBaz &lt; &gt;>"
	want3 := "<fooBarBaz < >>"

	got1 := html.UnescapeString(bracketRe.ReplaceAllStringFunc(raw1, sc.unSlackify))
	got2 := html.UnescapeString(bracketRe.ReplaceAllStringFunc(raw2, sc.unSlackify))
	got3 := html.UnescapeString(bracketRe.ReplaceAllStringFunc(raw3, sc.unSlackify))

	if got1 != want1 {
		t.Log("Unslackify failed:")
		t.Logf("Got: %v", got1)
		t.Logf("Want: %v", want1)
		t.Fail()
	}

	if got2 != want2 {
		t.Log("Unslackify failed:")
		t.Logf("Got: %v", got2)
		t.Logf("Want: %v", want2)
		t.Fail()
	}
	if got3 != want3 {
		t.Log("Unslackify failed:")
		t.Logf("Got: %v", got2)
		t.Logf("Want: %v", want2)
		t.Fail()
	}

}
