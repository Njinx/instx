package proxy

import (
	"errors"
	"testing"
)

func TestIsDDGBang(t *testing.T) {
	if err := initBangMap(); err != nil && !errors.Is(err, &ErrBangMapInitialized{}) {
		t.Errorf("Could not initialize the DDG Bang map: %s\n", err.Error())
	}

	tests := map[string]bool{
		"!!g":                              true,
		"!a":                               false,
		"! ddg":                            false,
		"!! gh":                            false,
		"!!yt   ":                          true,
		"  !!w":                            true,
		"!bing":                            false,
		"!!google this should be valid   ": true,
		"!!gfdjlkgdf this probably isn't":  false,
	}

	for input, expected := range tests {
		if given := isDDGBang(input); given != expected {
			t.Errorf(
				"isDDGBang(\"%s\"): expected = %t, given = %t\n",
				input,
				expected,
				given)
		}
	}
}

func TestInitBangMap(t *testing.T) {
	if err := initBangMap(); err != nil && !errors.Is(err, &ErrBangMapInitialized{}) {
		t.Errorf("Could not initialize the DDG Bang map: %s\n", err.Error())
	}

	tests := map[string]bool{
		"g":                  true,
		"gh":                 true,
		"google":             true,
		"g oogle":            false,
		"fsjfsfld":           false,
		"\u0448\u0440\u0443": true,
		"\u044f":             true,
		"\u044f\u044f\u044f\u044f\u044f\u044f\u044f\u044f\u044f\u044f": false,
		"ab-er": true,
	}

	for input, expected := range tests {
		if _, given := bangMap[input]; given != expected {
			t.Errorf(
				"bangMap[\"%s\"]: expected = %t, given = %t\n",
				input,
				expected,
				given)
		}
	}
}

func TestResolveDDGBang(t *testing.T) {
	if err := initBangMap(); err != nil && !errors.Is(err, &ErrBangMapInitialized{}) {
		t.Errorf("Could not initialize the DDG Bang map: %s\n", err.Error())
	}

	type expectedT struct {
		url   string
		isErr bool
	}

	tests := map[string]expectedT{
		"!!g":                  {"https://www.google.com/search?q=", false},
		"!!g  ":                {"https://www.google.com/search?q=", false},
		"!!rsub test":          {"https://reddit.com/r/test", false},
		"!!rsub test   ":       {"https://reddit.com/r/test", false},
		"!!a /\\bedding/\\/\\": {"https://www.amazon.com/s/?tag=duc0c-20&url=search-alias%3Daps&field-keywords=%2F%5Cbedding%2F%5C%2F%5C", false},
		"!!-_notvalid_-":       {"", true},
		"  !!yt  ":             {"https://www.youtube.com/results?search_query=", false},
		"  !!yt   funny      ": {"https://www.youtube.com/results?search_query=funny", false},
	}

	for input, expected := range tests {
		givenURL, givenErr := resolveDDGBang(input)
		if ((givenErr != nil) != expected.isErr) || (givenURL != expected.url) {
			t.Errorf(
				"resolveDDGBang(\"%s\"): expected = \"%s\", given = \"%s\"\n",
				input,
				expected.url,
				givenURL)
		}
	}
}
