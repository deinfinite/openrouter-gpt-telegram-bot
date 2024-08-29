package lang

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var translations map[string]map[string]interface{}

func LoadTranslations(langDir string) error {
	translations = make(map[string]map[string]interface{})

	languages := []string{"EN", "RU"}

	for _, lang := range languages {
		filePath := filepath.Join(langDir, lang+".json")
		data, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		var langMap map[string]interface{}
		err = json.Unmarshal(data, &langMap)
		if err != nil {
			return err
		}

		translations[lang] = langMap
	}

	for _, lang := range languages {
		filePath := filepath.Join(langDir, lang+".json")
		log.Printf("Loading translations from: %s", filePath)
	}

	//log.Printf("Loaded translations: %+v", translations)
	return nil
}

func Translate(key string, lang string) string {
	//log.Printf("Translating key: %s, language: %s", key, lang)
	if translations == nil {
		log.Println("Translations not loaded. Did you call LoadTranslations?")
		return key
	}
	keys := strings.Split(key, ".")
	value := interface{}(translations[lang])

	for _, k := range keys {
		if m, ok := value.(map[string]interface{}); ok {
			value = m[k]
		} else {
			return key
		}
	}

	if str, ok := value.(string); ok {
		return str
	}

	return key
}
