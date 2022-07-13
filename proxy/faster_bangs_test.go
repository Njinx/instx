package proxy

import (
	"testing"
)

func TestIsDDGBang(t *testing.T) {
	if err := initBangMap(); err != nil {
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
	if err := initBangMap(); err != nil {
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
