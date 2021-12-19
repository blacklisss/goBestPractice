package configuration

import (
	"fmt"
	"os"
	"testing"
)

type testFile struct {
	fileName string
	content  string
}

var testFiles = []testFile{
	{
		"config_test.json", "{\n \"startUrl\": \"https://gb.ru\"\n}",
	},
	{
		"config_test.yaml", "startUrl: \"https://google.com\"",
	},
	{
		"config_test.env", "EXTERNAL_URL=http://yandex.ru",
	},
}
var testFilesErrContentFormat = []testFile{
	{
		"config_test.json", "\n \"startUrl\": \"https://gb.ru\"\n",
	},
	{
		"config_test.yaml", "{startUrl: \"https://google.com\"}",
	},
	{
		"config_test.env", "{EXTERNAL_URL=http://yandex.ru}",
	},
}
var testFilesErrContent = []testFile{
	{
		"config_test.json", "{\n  \"startUrl\": \"//gb.ru\"\n}",
	},
	{
		"config_test.yaml", "startUrl: \"//google.com\"",
	},
}

func TestLoadOk(t *testing.T) {
	var err error
	var config *Config
	createTestFiles(testFiles)

	defer deleteTestFiles(testFiles)

	var tests = []struct {
		envFile string
		want    *Config
		error   error
	}{
		{"config_test.json", &Config{StartUrl: "https://gb.ru"}, nil},
		{"config_test.yaml", &Config{StartUrl: "https://google.com"}, nil},
		{"config_test.env", &Config{StartUrl: "http://yandex.ru"}, nil},
	}

	for _, test := range tests {
		config, err = Load(test.envFile)
		if err != nil {
			t.Errorf("произошла ошибка: «%v»", err)
		}

		if *config != *test.want {
			t.Errorf("Требуется: «%v». Пришло: «%v»", test.want, config)
		}
	}
}

func TestLoadErrFiles(t *testing.T) {
	var err error

	var tests = []struct {
		envFile string
		want    *Config
		error   error
	}{
		{"configuration1.env", nil, fmt.Errorf("произошла ошибка парсинга файла окружения")},
		{"configuration.envs", nil, fmt.Errorf("invalid format of configuration file")},
	}

	for _, test := range tests {
		_, err = Load(test.envFile)
		if err != nil {
			if err.Error() != test.error.Error() {
				t.Errorf("Ожидалось: «%v». Пришло: «%v»", test.error, err)
			}
		}
	}
}

/*func TestLoadErrFilesContentFormat(t *testing.T) {
	var err error
	createTestFiles(testFilesErrContentFormat)

	defer deleteTestFiles(testFilesErrContentFormat)

	var tests = []struct {
		envFile string
		want    *Config
		error   error
	}{
		{"config_test.json", nil, nil},
		{"config_test.yaml", nil, nil},
		{"config_test.env", nil, nil},
	}

	for _, test := range tests {
		_, err = Load(test.envFile)
		if err == nil {
			t.Errorf("Ожидалась ошибка: «%v» Файл: %s", err, test.envFile)
		} else {
			t.Logf("Пришла ошибка: %v", err)
		}
	}
}*/

func TestLoadErrFilesContent(t *testing.T) {
	var err error
	createTestFiles(testFilesErrContent)

	defer deleteTestFiles(testFilesErrContent)

	var tests = []struct {
		envFile string
		want    *Config
		error   error
	}{
		{"config_test.json", nil, nil},
		{"config_test.yaml", nil, nil},
	}

	for _, test := range tests {
		_, err = Load(test.envFile)
		if err == nil {
			t.Errorf("Ожидалась ошибка: «%v» Файл: %s", err, test.envFile)
		} else {
			t.Logf("Пришла ошибка: %v", err)
		}
	}
}

func TestValidateUrlOk(t *testing.T) {
	url := "https://www.gb.ru"

	if ok := validateUrl(&url); !ok {
		t.Errorf("Ожидалось: «true». Пришло: «%v»", ok)
	}
}

func TestValidateUrlNotOk(t *testing.T) {
	url := "www.gb.ru"

	if ok := validateUrl(&url); ok {
		t.Errorf("Ожидалось: «false». Пришло: «%v»", ok)
	}
}

func createTestFiles(testFiles []testFile) {

	for _, tf := range testFiles {
		f, err := os.OpenFile(tf.fileName, os.O_CREATE|os.O_WRONLY, os.ModePerm)
		if err != nil {
			panic(err)
		}
		if _, er := f.WriteString(tf.content); er != nil {
			panic(er)
		}

		if er := f.Close(); er != nil {
			panic(er)
		}
	}
}

func deleteTestFiles(testFiles []testFile) {
	for _, tf := range testFiles {
		err := os.Remove(tf.fileName)

		if err != nil {
			fmt.Println(err)
		}
	}
}
