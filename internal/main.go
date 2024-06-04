package internal

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

// func main() {
// 	log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
//
// 	out, err := run(os.Stdin)
// 	if err != nil {
// 		fmt.Fprintln(os.Stderr, err)
// 		os.Exit(exitFail)
// 	}
// 	fmt.Fprint(os.Stdout, string(out))
// }

// func runEnv(stdinRaw io.Reader) (string, error) {
// 	if err := checkStat(); err != nil {
// 		return "", err
// 	}
//   return run(stdinRaw)
// }

// func run(stdinRaw io.Reader) (string, error) {
// 	stdin := read(stdinRaw)

func Run(stdin []byte) (string, error) {

	isjson := isJSON(stdin)
	if !isjson {
		injsonbytes, err := yaml.YAMLToJSON(stdin)
		if err != nil {
			return "", fmt.Errorf("error converting from yaml to json : %w", err)
		}
		stdin = injsonbytes
	}

	var s secret
	if err := unmarshal(stdin, &s, isjson); err != nil {
		return "", err
	}

	// Check if the object is a list.
	if s["items"] != nil {
		for k, i := range s["items"].([]interface{}) {
			ii := i.(map[string]interface{})
			ss, err := stringData(ii)
			if err != nil {
				return "", err
			}
			s["items"].([]interface{})[k] = ss
		}
	} else {
		var err error
		s, err = stringData(s)
		if err != nil {
			return "", fmt.Errorf("error unmarshaling secret: %w", err)
		}
	}

	var bs []byte
	var err error
	if isjson {
		if bs, err = marshal(s); err != nil {
			return "", fmt.Errorf("can not marshal secret to JSON: %w", err)
		}
	} else {
		if bs, err = marshalYAML(s); err != nil {
			return "", fmt.Errorf("can not marshal secret to YAML: %w", err)
		}
	}

	return string(bs), nil
}

func checkStat() error {
	info, err := os.Stdin.Stat()
	if err != nil {
		return err
	}
	if (info.Mode()&os.ModeCharDevice) != 0 || info.Size() < 0 {
		return fmt.Errorf(notPipeError)
	}
	return nil
}

func isJSON(s []byte) bool {
	return json.Unmarshal(s, &json.RawMessage{}) == nil
}

func cast(data interface{}) (map[string]interface{}, bool) {
	d, ok := data.(map[string]interface{})
	return d, ok
}

func stringData(s secret) (secret, error) {
	data, ok := cast(s["data"])
	if !ok || len(data) == 0 {
		return s, nil
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

func unmarshal(in []byte, out interface{}, asJSON bool) error {
	if asJSON {
		return json.Unmarshal(in, out)
	}
	return yaml.Unmarshal(in, out)
}

func marshal(d interface{}) ([]byte, error) {
	return json.MarshalIndent(d, "", "    ")
}

func marshalYAML(d interface{}) ([]byte, error) {
	return yaml.Marshal(d)
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
