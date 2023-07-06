package settings

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"sort"

	"github.com/femnad/fup/internal"
)

type FactMap map[string]map[string]string

type Settings struct {
	CloneDir      string            `yaml:"clone_dir"`
	ExtractDir    string            `yaml:"extract_dir"`
	HostFacts     FactMap           `yaml:"host_facts"`
	SelfClonePath string            `yaml:"self_clone_path"`
	TemplateDir   string            `yaml:"template_dir"`
	Versions      map[string]string `yaml:"versions"`
	VirtualEnvDir string            `yaml:"virtualenv_dir"`
}

func expand(s string, lookup map[string]string) string {
	var cur bytes.Buffer
	var out bytes.Buffer
	var backspace bool
	var consuming bool
	var dollar bool

	for _, c := range s {
		if backspace {
			if c != '$' {
				out.WriteRune('\\')
			}
		} else if c == '$' {
			backspace = false
			dollar = true
			continue
		}

		backspace = c == '\\'
		if backspace {
			continue
		}

		if dollar {
			if c == '{' {
				dollar = false
				consuming = true
				continue
			} else {
				out.WriteRune('$')
				dollar = false
			}
		}

		if c == '}' && consuming {
			consuming = false
			curStr := cur.String()
			val, ok := lookup[curStr]
			envLookup := os.Getenv(curStr)
			if ok {
				out.WriteString(val)
			} else if envLookup != "" {
				out.WriteString(envLookup)
			} else {
				orig := fmt.Sprintf("${%s}", curStr)
				out.WriteString(orig)
			}
			cur.Reset()
			continue
		}

		if consuming {
			cur.WriteRune(c)
		} else {
			out.WriteRune(c)
		}
	}

	return out.String()
}

func addHostFacts(lookup map[string]string, factMap FactMap) map[string]string {
	hostName, err := os.Hostname()
	if err != nil {
		internal.Log.Errorf("error determining hostname: %v", err)
		return lookup
	}

	for fact, hostFacts := range factMap {
		facts := make([]string, 0, len(hostFacts))
		for key := range hostFacts {
			facts = append(facts, key)
		}

		// Sort by length in reverse to start with more specific regexes (probably).
		sort.Slice(facts, func(i, j int) bool {
			return len(facts[i]) > len(facts[j])
		})

		for _, regex := range facts {
			value := hostFacts[regex]
			cmp, err := regexp.Compile(regex)
			if err != nil {
				internal.Log.Errorf("ignoring regexp in host fact: %v", err)
				continue
			}
			if cmp.MatchString(hostName) {
				lookup[fact] = value
				break
			}
		}
	}

	return lookup
}

func ExpandStringWithLookup(settings Settings, s string, lookup map[string]string) string {
	lookup["clone_dir"] = settings.CloneDir
	lookup["extract_dir"] = settings.ExtractDir
	lookup = addHostFacts(lookup, settings.HostFacts)

	expanded := expand(s, lookup)
	return internal.ExpandUser(expanded)
}

func ExpandString(settings Settings, s string) string {
	return ExpandStringWithLookup(settings, s, map[string]string{})
}
