package main

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
)

var testConfig01 = &Configuration{
	Obfuscations: []TargetedObfuscation{
		TargetedObfuscation{
			Target{Table: "auth_user", Column: "email"},
			ScrambleEmail,
		},
		TargetedObfuscation{
			Target{Table: "auth_user", Column: "password"},
			GenScrambleBytes(7),
		},
		TargetedObfuscation{
			Target{Table: "accounts_profile", Column: "phone"},
			ScrambleDigits,
		},
	},
}

const testInput01 = `
--

SELECT pg_catalog.setval('auth_user_id_seq', 123111, true);


--
-- Data for Name: auth_user; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY auth_user (id, username, first_name, last_name, email, password, is_staff, is_active, is_superuser, last_login, date_joined) FROM stdin;
123123111	499964777.sdsad	testname	testname	499964777.sdsad.com@dev.bing	!	f	t	f	2011-02-07 12:08:30+00	2010-11-22 19:27:12.31832+00
333114441	testT1@bing.com			testT1@bing.com	!	f	t	f	2011-06-08 12:57:36+00	2011-06-08 12:50:25.206298+00
515131311	whoisthere			noreply781134796251@bing.com	pbkdf2_sha256$10000$qweqweqweqwe$cThxOHE4	f	t	f	2012-11-16 18:27:43.673889+00	2012-11-16 18:27:43.229281+00
\.

COPY accounts_profile (id, user_id, opted_in, next_break, status, phone, last_visited, come_from, cs_letter, city_id, budget_range_id, prefs_opt) FROM stdin;
6161	12113	f	\N	0	+74991002000	2011-07-04 12:28:33.895325+00	\N	f	\N	\N	\N
1223	1321	f	\N	0	666666666	2011-09-28 09:37:20.83051+00	\N	f	\N	\N	\N
4423	55512	f	\N	0		\N	\N	f	\N	\N	\N
\.
`

func TestProcess01(t *testing.T) {
	input := bufio.NewReader(bytes.NewBufferString(testInput01))
	output := new(bytes.Buffer)
	process(testConfig01, input, output)
	outString := output.String()
	if outString == testInput01 {
		t.Fatal("Outputs are equal")
	}
	if !strings.Contains(outString, "COPY auth_user (id, username, first_name, last_name, email, password, is_staff, is_active, is_superuser, last_login, date_joined) FROM stdin;") ||
		!strings.Contains(outString, "COPY accounts_profile (id, user_id, opted_in, next_break, status, phone, last_visited, come_from, cs_letter, city_id, budget_range_id, prefs_opt) FROM stdin;") {
		t.Fatal("Changed SQL")
	}
	if strings.Contains(outString, "499964777.sdsad.com@dev.bing") ||
		strings.Contains(outString, "pbkdf2_sha256$10000$qweqweqweqwe$cThxOHE4") ||
		strings.Contains(outString, "+3801445223001") {
		t.Fatal("Did not scramble sensitive data")
	}
	if !strings.Contains(outString, "515131311	whoisthere") ||
		!strings.Contains(outString, `	2011-07-04 12:28:33.895325+00	\N	f	\N	\N	\N`) ||
		!strings.Contains(outString, `1223	1321	f	\N	0	`) {
		t.Fatal("Changed other data")
	}
}

func TestScrambleDigits01(t *testing.T) {
	Salt = []byte("test-salt")
	out := string(ScrambleDigits([]byte("+7(876) 123-0011 или 99999999999;")))
	if out != "+1(584) 047-9250 или 22280031035;" {
		t.Fatal("ScrambleDigits: invalid result Digits:", out)
	}
}

func BenchmarkScrambleDigits01(b *testing.B) {
	Salt = []byte("test-salt")
	Digits := []byte("+7(876) 123-0011")
	for i := 0; i < b.N; i++ {
		if ScrambleDigits(Digits) == nil {
			b.Fatal("Result is nil")
		}
	}
}
