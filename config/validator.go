package config

import (
	"fmt"
	urllib "net/url"
	"strings"
	"time"
)

type InvalidValue struct {
	key      string
	given    string
	accepted string
}

func (e *InvalidValue) Error() string {
	return fmt.Sprintf(
		"[%s] Invalid value for \"%s\": \"%s\". Accepted: %s",
		DEFAULT_CONFIG_FILE, e.key, e.given, e.accepted)
}

func (c *Config) validateConfig() error {

	if _, err := urllib.ParseRequestURI(c.DefaultInstance); err != nil {
		return &InvalidValue{
			key:      "default_instance",
			given:    c.DefaultInstance,
			accepted: "Any valid URL.",
		}
	}

	if c.Proxy.Port < 0 || c.Proxy.Port > 65535 {
		return &InvalidValue{
			key:      "proxy.port",
			given:    fmt.Sprint(c.Proxy.Port),
			accepted: "Any valid TCP port number.",
		}
	}

	minTime := int64(0)
	maxTime := int64(^uint64(0)>>1) / int64(time.Minute)
	if c.Updater.UpdateInterval < minTime || c.Updater.UpdateInterval > maxTime {
		return &InvalidValue{
			key:      "updater.update_interval",
			given:    fmt.Sprint(c.Proxy.Port),
			accepted: fmt.Sprintf("Any number (in minutes) from %d-%d.", minTime, maxTime),
		}
	}

	respWeightHelper := func(k string, v float64) (bool, InvalidValue) {
		if v <= 0 || v >= 2 {
			return false, InvalidValue{
				key:      k,
				given:    fmt.Sprint(v),
				accepted: "Any number n: 0 < n < 2. Check the README for more information.",
			}
		} else {
			return true, InvalidValue{}
		}
	}

	if ok, ret := respWeightHelper(
		"updater.advanced.initial_resp_weight",
		c.Updater.Advanced.InitialRespWeight); !ok {

		return &ret
	}
	if ok, ret := respWeightHelper(
		"updater.advanced.search_resp_weight",
		c.Updater.Advanced.SearchRespWeight); !ok {

		return &ret
	}
	if ok, ret := respWeightHelper(
		"updater.advanced.google_search_resp_weight",
		c.Updater.Advanced.GoogleSearchRespWeight); !ok {

		return &ret
	}
	if ok, ret := respWeightHelper(
		"updater.advanced.wikipedia_search_resp_weight",
		c.Updater.Advanced.WikipediaSearchRespWeight); !ok {

		return &ret
	}

	isLetterGrade := func(grade string) bool {
		switch grade {
		case "A+":
			return true
		case "A":
			return true
		case "A-":
			return true
		case "B+":
			return true
		case "B":
			return true
		case "B-":
			return true
		case "C+":
			return true
		case "C":
			return true
		case "C-":
			return true
		case "D+":
			return true
		case "D":
			return true
		case "D-":
			return true
		case "F":
			return true
		default:
			return false
		}
	}

	if !isLetterGrade(c.Updater.Criteria.MinimumCspGrade) {
		return &InvalidValue{
			key:      "updater.criteria.minimum_csp_grade",
			given:    c.Updater.Criteria.MinimumCspGrade,
			accepted: "A+, A, A-, B+, B, B-, C+, C, C-, D+, D, D-, F",
		}
	}
	if !isLetterGrade(c.Updater.Criteria.MinimumTlsGrade) {
		return &InvalidValue{
			key:      "updater.criteria.minimum_tls_grade",
			given:    c.Updater.Criteria.MinimumTlsGrade,
			accepted: "A+, A, A-, B+, B, B-, C+, C, C-, D+, D, D-, F",
		}
	}

	for _, grade := range c.Updater.Criteria.AllowedHttpGrades {
		switch strings.ToLower(grade) {
		case "v":
			continue
		case "f":
			continue
		case "c":
			continue
		case "cjs":
			continue
		case "e":
			continue
		case "👁️":
			continue
		default:
			return &InvalidValue{
				key:      "updater.criteria.allowed_http_grades",
				given:    strings.Join(c.Updater.Criteria.AllowedHttpGrades, ", "),
				accepted: "Check the README.",
			}
		}
	}

	switch strings.ToLower(c.Updater.Criteria.SearxngPreference) {
	case "required":
		break
	case "forbidden":
		break
	case "impartial":
		break
	default:
		return &InvalidValue{
			key:      "updater.criteria.searxng_preference",
			given:    c.Updater.Criteria.SearxngPreference,
			accepted: "required, forbidden, impartial. Check the README for more information.",
		}
	}

	return nil
}
