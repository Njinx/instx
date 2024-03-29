package config

import (
	"fmt"
	"net"
	urllib "net/url"
	"regexp"
	"strings"
	"time"

	"gitlab.com/Njinx/instx/util"
)

type ErrInvalidValue struct {
	key      string
	given    string
	accepted string
}

func (e *ErrInvalidValue) Error() string {
	return fmt.Sprintf(
		"[%s] Invalid value for \"%s\": \"%s\". Accepted: %s",
		DEFAULT_CONFIG_FILE, e.key, e.given, e.accepted)
}

type ErrCouldNotBindPort struct {
	key   string
	given string
	err   error
}

func (e *ErrCouldNotBindPort) Error() string {
	return fmt.Sprintf(
		"[%s] Unable to bind to port \"%s\" specified by \"%s\": %s",
		DEFAULT_CONFIG_FILE, e.given, e.key, e.err.Error())
}

// Validate instx.yaml
func (c *Config) validateConfig() []error {
	errorArray := make([]error, 0, 64)

	if _, err := urllib.ParseRequestURI(c.DefaultInstance); err != nil {
		errorArray = append(errorArray, &ErrInvalidValue{
			key:      "default_instance",
			given:    c.DefaultInstance,
			accepted: "Any valid URL (accepted by net.url.Parse)",
		})
	}

	if !util.IsInstxCtlMode() {
		if err := tryBindToPort(c.Proxy.Port); err != nil {
			errorArray = append(errorArray, &ErrCouldNotBindPort{
				key:   "proxy.port",
				given: fmt.Sprint(c.Proxy.Port),
				err:   err,
			})
		}
	}

	if c.Proxy.Port < 0 || c.Proxy.Port > 65535 {
		errorArray = append(errorArray, &ErrInvalidValue{
			key:      "proxy.port",
			given:    fmt.Sprint(c.Proxy.Port),
			accepted: "Any valid TCP port number.",
		})
	}

	minTime := int64(0)
	maxTime := int64(^uint64(0)>>1) / int64(time.Minute)
	if c.Updater.UpdateInterval < minTime || c.Updater.UpdateInterval > maxTime {
		errorArray = append(errorArray, &ErrInvalidValue{
			key:      "updater.update_interval",
			given:    fmt.Sprint(c.Proxy.Port),
			accepted: fmt.Sprintf("Any number (in minutes) from %d-%d.", minTime, maxTime),
		})
	}

	respWeightHelper := func(k string, v float64) {
		if v <= 0 || v >= 2 {
			errorArray = append(errorArray, &ErrInvalidValue{
				key:      k,
				given:    fmt.Sprint(v),
				accepted: "Any number n: 0 < n < 2. Check the README for more information.",
			})
		}
	}

	respWeightHelper(
		"updater.advanced.initial_resp_weight",
		c.Updater.Advanced.InitialRespWeight)
	respWeightHelper(
		"updater.advanced.search_resp_weight",
		c.Updater.Advanced.SearchRespWeight)
	respWeightHelper(
		"updater.advanced.google_search_resp_weight",
		c.Updater.Advanced.GoogleSearchRespWeight)
	respWeightHelper(
		"updater.advanced.wikipedia_search_resp_weight",
		c.Updater.Advanced.WikipediaSearchRespWeight)

	// Checks whether grade is a valid letter grade.
	// Valid: A+, A, A-, B+, B, B-, C+, C, C-, D+, D, D-, F
	// Case-insensitive and does not care about surrounding whitespace.
	//   Ex: "  A+ " matches but "  A + " does not.
	isLetterGrade := func(grade string) bool {
		re, _ := regexp.Compile(`^\s*(?:[a-dA-D][-+]?|[fF])\s*$`)
		return re.Match([]byte(grade))
	}

	if !isLetterGrade(c.Updater.Criteria.MinimumCspGrade) {
		errorArray = append(errorArray, &ErrInvalidValue{
			key:      "updater.criteria.minimum_csp_grade",
			given:    c.Updater.Criteria.MinimumCspGrade,
			accepted: "A+, A, A-, B+, B, B-, C+, C, C-, D+, D, D-, F",
		})
	}
	if !isLetterGrade(c.Updater.Criteria.MinimumTlsGrade) {
		errorArray = append(errorArray, &ErrInvalidValue{
			key:      "updater.criteria.minimum_tls_grade",
			given:    c.Updater.Criteria.MinimumTlsGrade,
			accepted: "A+, A, A-, B+, B, B-, C+, C, C-, D+, D, D-, F",
		})
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
			errorArray = append(errorArray, &ErrInvalidValue{
				key:      "updater.criteria.allowed_http_grades",
				given:    strings.Join(c.Updater.Criteria.AllowedHttpGrades, ", "),
				accepted: "Check the README.",
			})
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
		errorArray = append(errorArray, &ErrInvalidValue{
			key:      "updater.criteria.searxng_preference",
			given:    c.Updater.Criteria.SearxngPreference,
			accepted: "required, forbidden, impartial. Check the README for more information.",
		})
	}

	for i, inst := range c.Updater.InstanceBlacklist {
		url, err := urllib.Parse(inst)
		if err != nil || len(url.Host) == 0 {
			errorArray = append(errorArray, &ErrInvalidValue{
				key:      fmt.Sprintf("updater.instance_blacklist[%d]", i),
				given:    inst,
				accepted: "Any valid URL (net.url.Parse)",
			})
		}
	}

	return errorArray
}

// Attempt to bind to port. Successful if return is nil.
func tryBindToPort(port int) error {
	conn, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	err = conn.Close()
	return err
}
