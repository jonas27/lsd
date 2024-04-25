package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/ghodss/yaml"
)

const (
	exitFail     = 1
	notPipeError = `lsd is intended to work with pipes in either YAML or JSON format.")
usage: kubectl get secret -o yaml | lsd
usage: lsd secret.yaml`
)

type secret map[string]interface{}

type decodedSecret struct {
	Key   string
	Value string
}

var log *slog.Logger

func main() {
	info, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}

	if (info.Mode()&os.ModeCharDevice) != 0 || info.Size() < 0 {
		fmt.Println(notPipeError)
		os.Exit(exitFail)
	}

	log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	out, err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(exitFail)
	}
	fmt.Fprint(os.Stdout, string(out))
}

func run() (string, error) {
	stdin := read(os.Stdin)

	var injson string
	isjson := isJSON(stdin)
	if !isjson {
		injsonbytes, err := yaml.YAMLToJSON(stdin)
		if err != nil {
			return "", fmt.Errorf("error converting from yaml to json : %w", err)
		}
		injson = string(injsonbytes)
	} else {
		injson = string(stdin)
	}
	_ = injson

	b, err := parse(stdin)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func isJSON(s []byte) bool {
	return json.Unmarshal(s, &json.RawMessage{}) == nil
}

func cast(data interface{}, isJSON bool) (map[string]interface{}, bool) {
	if isJSON {
		d, ok := data.(map[string]interface{})
		return d, ok
	}

	parsed, ok := data.(map[interface{}]interface{})
	if !ok {
		return nil, false
	}
	d := make(map[string]interface{}, len(parsed))
	for key, value := range parsed {
		d[key.(string)] = value
	}
	return d, true
}

func parse(in []byte) (secret, error) {
	isJSON := isJSON(in)

	var s secret
	if err := unmarshal(in, &s, isJSON); err != nil {
		return nil, err
	}

	if isList(s) {
		for k, i := range s["items"].([]interface{}) {
			item, err := json.Marshal(i)
			if err != nil {
				return nil, err
			}
			parsed, err := parse(item)
			if err != nil {
				return nil, err
			}
			_ = k
			s["items"].([]interface{})[k] = parsed
		}
		fmt.Println(s)
		os.Exit(1)
		return s, nil
	}

	data, ok := cast(s["data"], isJSON)
	if !ok || len(data) == 0 {
		return nil, fmt.Errorf("could not read secret")
	}
	s["stringData"] = decode(data)
	delete(s, "data")
  return s, nil
}

func read(rd io.Reader) []byte {
	var output []byte
	reader := bufio.NewReader(rd)
	for {
		input, err := reader.ReadByte()
		if err != nil && err == io.EOF {
			break
		}
		output = append(output, input)
	}
	return output
}

func isList(s secret) bool {
	return s["items"] != nil
}

func unmarshal(in []byte, out interface{}, asJSON bool) error {
	if asJSON {
		return json.Unmarshal(in, out)
	}
	return yaml.Unmarshal(in, out)
}

func marshal(d interface{}) ([]byte, error) {
	return json.MarshalIndent(d, "", "    ")
}

func decodeSecret(key, secret string, secrets chan decodedSecret) {
	var value string
	// avoid wrong encoded secrets
	if decoded, err := base64.StdEncoding.DecodeString(secret); err == nil {
		value = string(decoded)
	} else {
		value = secret
	}
	secrets <- decodedSecret{Key: key, Value: value}
}

func decode(data map[string]interface{}) map[string]string {
	length := len(data)
	secrets := make(chan decodedSecret, length)
	decoded := make(map[string]string, length)
	for key, encoded := range data {
		go decodeSecret(key, encoded.(string), secrets)
	}
	for i := 0; i < length; i++ {
		secret := <-secrets
		decoded[secret.Key] = secret.Value
	}
	return decoded
}
