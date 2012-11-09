package i18n

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"strings"
	"sync"
)

type I18nManager struct {
	I18nName        string
	localePath      string
	defaultLanguage string
	rmutex          sync.RWMutex
	mutex           sync.Mutex
	languages       map[string]map[string]string
	lastModTime     map[string]int64
}

func New(name, path, lang string) *I18nManager {
	if name == "" {
		name = "GoLangerI18n"
	}

	if path == "" {
		path = "./locale/"
	}
	if lang == "" {
		lang = "zh-cn"
	}

	i := &I18nManager{
		I18nName:        name,
		localePath:      path,
		defaultLanguage: lang,
		languages:       map[string]map[string]string{},
		lastModTime:     map[string]int64{},
	}

	return i
}

func loadFile(filepath string) (data []byte, err error) {
	if filepath == "" {
		return nil, errors.New("Filepath Error")
	}
	data, err = ioutil.ReadFile(filepath)
	return
}

func (i *I18nManager) loadLanguageFile(lang string) error {
	if lang == "" {
		lang = i.defaultLanguage
	}

	file := i.localePath + lang
	dataFi, err := os.Stat(file)

	i.rmutex.RLock()
	if err == nil {
		if dataFi.ModTime().Unix() <= i.lastModTime[lang] {
			_, ok := i.languages[lang]
			if ok {
				i.rmutex.RUnlock()
				return nil
			}
		}
	}

	i.rmutex.RUnlock()

	data, err := loadFile(file)
	if err != nil {
		i.rmutex.RLock()
		_, ok := i.languages[i.defaultLanguage]
		i.rmutex.RUnlock()

		if ok {
			return nil
		} else {
			return errors.New("Error: Loading Language File " + file)
		}
	}

	m := map[string]string{}
	err = json.Unmarshal(data, &m)

	i.mutex.Lock()
	if err == nil {
		i.languages[lang] = m
		i.lastModTime[lang] = dataFi.ModTime().Unix()
	}
	i.mutex.Unlock()

	return err
}

func (i *I18nManager) Lang(l string) map[string]string {
	l = strings.ToLower(l)
	i.rmutex.RLock()
	defer i.rmutex.RUnlock()

	file := i.localePath + l
	dataFi, err := os.Stat(file)
	if err == nil {
		if dataFi.ModTime().Unix() > i.lastModTime[l] {
			err := i.loadLanguageFile(l)
			if err != nil {
				l = i.defaultLanguage
			}
		}
	}

	msgs, found := i.languages[l]
	if !found {
		// Load The Language File
		err := i.loadLanguageFile(l)
		if err != nil {
			l = i.defaultLanguage
		}

		msgs = i.languages[l]
	}

	return msgs
}

func (i *I18nManager) Get(lang, key string) string {
	lang = strings.ToLower(lang)

	targetLang := ""
	i.rmutex.RLock()

	file := i.localePath + lang
	dataFi, err := os.Stat(file)
	if err == nil {
		if dataFi.ModTime().Unix() > i.lastModTime[lang] {
			err := i.loadLanguageFile(lang)
			if err != nil {
				lang = i.defaultLanguage
			}
		}
	}

	msgs, found := i.languages[lang]
	if !found {
		// Load The Language File
		err := i.loadLanguageFile(lang)
		if err != nil {
			var ok bool
			targetLang, ok = i.languages[i.defaultLanguage][key]
			i.rmutex.RUnlock()
			if ok {
				return targetLang
			} else {
				return key
			}
		}

		msgs = i.languages[lang]
	}

	targetLang, ok := msgs[key]
	i.rmutex.RUnlock()

	if !ok {
		targetLang = key
	}

	return targetLang
}
