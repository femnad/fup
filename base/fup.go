package base

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"os"
	"text/template"

	"gopkg.in/yaml.v3"

	"github.com/femnad/fup/entity"
	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/precheck"
	"github.com/femnad/fup/remote"
)

type configReader struct {
	reader   io.Reader
	isRemote bool
}

type configOut struct {
	content  []byte
	filename string
	isRemote bool
}

func readLocalConfigFile(config string) (io.Reader, error) {
	f, err := os.Open(config)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func readRemoteConfigFile(config string) (io.Reader, error) {
	response, err := remote.ReadResponseBody(config)
	if err != nil {
		return nil, err
	}

	return response.Body, nil
}

func getConfigReader(config string) (configReader, error) {
	parsed, err := url.Parse(config)
	if err != nil {
		return configReader{}, err
	}

	var readerFn func(string) (io.Reader, error)
	var isRemote bool
	if parsed.Scheme == "" {
		readerFn = readLocalConfigFile
		isRemote = false
	} else {
		readerFn = readRemoteConfigFile
		isRemote = true
	}

	reader, err := readerFn(config)
	if err != nil {
		return configReader{}, err
	}

	return configReader{reader: reader, isRemote: isRemote}, nil
}

func evalConfig(data []byte) ([]byte, error) {
	tmpl := template.New("config").Funcs(precheck.FactFns).Funcs(internal.UtilFns)
	parsed, err := tmpl.Parse(string(data))
	if err != nil {
		return data, err
	}

	var out bytes.Buffer
	err = parsed.Execute(&out, nil)
	if err != nil {
		return data, err
	}

	return out.Bytes(), nil
}

func finalizeConfig(filename string) (configOut, error) {
	cfgReader, err := getConfigReader(filename)
	if err != nil {
		return configOut{}, err
	}

	data, err := io.ReadAll(cfgReader.reader)
	if err != nil {
		return configOut{}, err
	}

	data, err = evalConfig(data)
	if err != nil {
		return configOut{}, err
	}

	return configOut{
		content:  data,
		filename: filename,
		isRemote: cfgReader.isRemote,
	}, nil
}

func FinalizeConfig(filename string) (string, error) {
	filename = internal.ExpandUser(filename)
	out, err := finalizeConfig(filename)
	if err != nil {
		return "", err
	}

	return string(out.content), nil
}

func unmarshalConfig(filename string) (config entity.Config, err error) {
	finalConfig, err := finalizeConfig(filename)
	if err != nil {
		return config, err
	}

	err = yaml.Unmarshal(finalConfig.content, &config)
	if err != nil {
		return config, fmt.Errorf("error deserializing config from %s: %v", filename, err)
	}

	config.Filename = finalConfig.filename
	config.Remote = finalConfig.isRemote
	return
}

func ReadConfig(filename string) (entity.Config, error) {
	filename = internal.ExpandUser(filename)
	return unmarshalConfig(filename)
}
