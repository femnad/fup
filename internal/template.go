package internal

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

var UtilFns = template.FuncMap{
	"cut":     cut,
	"head":    head,
	"iter":    iterItems,
	"revCut":  reverseCut,
	"split":   split,
	"splitBy": splitBy,
}

func absIndex(s string, i int) (int, error) {
	sLen := len(s)
	if i < 0 {
		i = sLen + i
	}
	if i < 0 || i >= sLen {
		return 0, fmt.Errorf("invalid index %d for string %s", i, s)
	}

	return i, nil
}

func cut(i int, s string) (string, error) {
	i, err := absIndex(s, i)
	if err != nil {
		return "", err
	}

	return s[i:], nil
}

func head(i int, s string) (string, error) {
	return splitBy("\n", i, s)
}

func iterItems(item ...string) []string {
	return item
}

func reverseCut(i int, s string) (string, error) {
	i, err := absIndex(s, i)
	if err != nil {
		return "", err
	}

	return s[:i], nil
}

func split(i int, s string) (string, error) {
	return splitBy(" ", i, s)
}

func splitBy(delimiter string, i int, s string) (string, error) {
	s = strings.Trim(s, delimiter)
	fields := strings.Split(s, delimiter)
	numFields := len(fields)
	if i < 0 {
		i = numFields + i
	}
	if i >= numFields {
		return "", fmt.Errorf("input %s has not have field with index %d when split by %s", s, i, delimiter)
	}
	return fields[i], nil
}

func RunTemplateFn(input, tmplFn string) (string, error) {
	tmpl := template.New("post-proc").Funcs(UtilFns)

	ctx := struct {
		Args string
	}{Args: input}

	tmplTxt := fmt.Sprintf("{{ print .Args | %s }}", tmplFn)
	parsed, err := tmpl.Parse(tmplTxt)
	if err != nil {
		return "", err
	}

	var out bytes.Buffer
	err = parsed.Execute(&out, ctx)
	if err != nil {
		Log.Errorf("error executing function %s on input %s: %v", tmplFn, input, err)
		return "", err
	}

	return out.String(), nil
}
