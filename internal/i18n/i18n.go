package i18n

import (
	"fmt"
	"sync"

	"github.com/orxma/depwise/internal/db"
)

var translations = map[string]map[string]string{
	"en": enStrings,
}

var (
	langCache   = make(map[int64]string)
	langCacheMu sync.RWMutex
	cacheOnce   sync.Once
)

func loadCacheFromDB() {
	_, err := db.Load()
	if err != nil {
		return
	}
}

func GetLang(chatID int64) string {
	return "en"
}

func SetLang(chatID int64, lang string) {
}

func T(chatID int64, key string) string {
	if strs, ok := translations["en"]; ok {
		if val, ok := strs[key]; ok {
			return val
		}
	}
	return key
}

func Tf(chatID int64, key string, args ...interface{}) string {
	return fmt.Sprintf(T(chatID, key), args...)
}

func TLang(lang string, key string) string {
	if strs, ok := translations["en"]; ok {
		if val, ok := strs[key]; ok {
			return val
		}
	}
	return key
}

func TfLang(lang string, key string, args ...interface{}) string {
	return fmt.Sprintf(TLang(lang, key), args...)
}

