package links

import "testing"

func TestRoundTrip(t *testing.T) {
	in := Payload{
		Name:     "iran-server",
		Host:     "198.105.115.89",
		Port:     2222,
		User:     "u_test",
		Password: "fOY6FR/UcxzQwdw1tmnDj8k9",
	}
	link, err := Build(in)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	out, err := Parse(link)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if *out != (Payload{
		V: Version, Name: in.Name, Host: in.Host, Port: in.Port, User: in.User, Password: in.Password,
	}) {
		t.Fatalf("mismatch:\n in:  %+v\n out: %+v", in, *out)
	}
}

func TestParseRejectsUnknownVersion(t *testing.T) {
	link, _ := Build(Payload{V: 2, Host: "x", Port: 22, User: "u", Password: "p"})
	if _, err := Parse(link); err == nil {
		t.Fatal("expected error for unknown version")
	}
}

func TestParseRejectsNonUnic(t *testing.T) {
	if _, err := Parse("vless://abc"); err == nil {
		t.Fatal("expected error for non-unic link")
	}
}

func TestParseAcceptsPaddedBase64(t *testing.T) {
	link, _ := Build(Payload{Host: "x", Port: 22, User: "u", Password: "p"})
	if _, err := Parse(link + "=="); err != nil {
		t.Fatalf("parse padded: %v", err)
	}
}

func TestParseRejectsGarbage(t *testing.T) {
	cases := []string{"unic://", "unic://!!!", "unic://" + "Zm9v" /* "foo" — not JSON */}
	for _, c := range cases {
		if _, err := Parse(c); err == nil {
			t.Fatalf("expected error parsing %q", c)
		}
	}
}
