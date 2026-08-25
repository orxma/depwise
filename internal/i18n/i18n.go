package i18n

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/orxma/depwise/internal/db"
)

var translations = map[string]map[string]string{
	"es": esStrings,
	"en": enStrings,
}

var (
	langCache   = make(map[int64]string)
	langCacheMu sync.RWMutex
	cacheOnce   sync.Once
)

func loadCacheFromDB() {
	data, err := db.Load()
	if err != nil || data.UserLanguages == nil {
		return
	}
	langCacheMu.Lock()
	defer langCacheMu.Unlock()
	for idStr, lang := range data.UserLanguages {
		id, _ := strconv.ParseInt(idStr, 10, 64)
		if id != 0 {
			langCache[id] = lang
		}
	}
}

// GetLang returns the preferred language for a user ("es" or "en")
func GetLang(chatID int64) string {
	cacheOnce.Do(loadCacheFromDB)
	langCacheMu.RLock()
	lang, ok := langCache[chatID]
	langCacheMu.RUnlock()
	if ok {
		return lang
	}
	return "es"
}

// HasLang checks if a user has already selected a language
func HasLang(chatID int64) bool {
	cacheOnce.Do(loadCacheFromDB)
	langCacheMu.RLock()
	_, ok := langCache[chatID]
	langCacheMu.RUnlock()
	return ok
}

// SetLang saves the language preference for a user (in cache and DB)
func SetLang(chatID int64, lang string) {
	cacheOnce.Do(loadCacheFromDB)

	// Update cache
	langCacheMu.Lock()
	langCache[chatID] = lang
	langCacheMu.Unlock()

	// Persist to DB
	db.Update(func(data *db.ConfigData) error {
		if data.UserLanguages == nil {
			data.UserLanguages = make(map[string]string)
		}
		data.UserLanguages[fmt.Sprintf("%d", chatID)] = lang
		return nil
	})
}

// T returns a translated string for a user's preferred language
func T(chatID int64, key string) string {
	lang := GetLang(chatID)
	if strs, ok := translations[lang]; ok {
		if val, ok := strs[key]; ok {
			return val
		}
	}
	// Fallback to Spanish
	if val, ok := esStrings[key]; ok {
		return val
	}
	return key
}

// Tf returns a formatted translated string (uses fmt.Sprintf)
func Tf(chatID int64, key string, args ...interface{}) string {
	return fmt.Sprintf(T(chatID, key), args...)
}

// TLang returns a translated string for a specific language code (useful for background tasks)
func TLang(lang string, key string) string {
	if strs, ok := translations[lang]; ok {
		if val, ok := strs[key]; ok {
			return val
		}
	}
	if val, ok := esStrings[key]; ok {
		return val
	}
	return key
}

// TfLang returns a formatted translated string for a specific language code
func TfLang(lang string, key string, args ...interface{}) string {
	return fmt.Sprintf(TLang(lang, key), args...)
}
