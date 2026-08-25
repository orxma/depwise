package db

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// ConfigData representa el archivo bot_data.json
type ConfigData struct {
	SuperAdmin      string           `json:"super_admin"`
	SuperAdminID    string           `json:"super_admin_id"`
	Admins          map[string]AdminInfo `json:"admins"`
	ExtraInfo        string               `json:"extra_info"`
	UserHistory      []int64              `json:"user_history"`
	PublicAccess     bool                 `json:"public_access"`
	PublicScanner    bool                 `json:"public_scanner"`
	SSHOwners        map[string]string    `json:"ssh_owners"`
	SSHTimeUsers     map[string]string    `json:"ssh_time_users"` // user -> expire date
	CloudflareDomain string               `json:"cloudflare_domain"`
	CloudfrontDomain string               `json:"cloudfront_domain"`
	ProxyDT          ProxyDTConfig        `json:"proxydt"`
	SlowDNS          SlowDNSConfig        `json:"slowdns"`
	VayDNS           VayDNSConfig         `json:"vaydns"`
	Slipstream       SlipstreamInfo       `json:"slipstream"`
	Zivpn            bool                 `json:"zivpn"`
	ZivpnUsers       map[string]string    `json:"zivpn_users"`  // password -> expire
	ZivpnOwners      map[string]string    `json:"zivpn_owners"` // password -> owner chat ID
	BadVPN           bool                 `json:"badvpn"`
	UDPCustom        bool                 `json:"udp_custom"`
	Falcon           string               `json:"falcon"`     // Port as string for compatibility
	Dropbear         string               `json:"dropbear"`   // Port as string for compatibility
	SSLTunnel        string               `json:"ssl_tunnel"` // Port as string for compatibility
	SSHBanner        string               `json:"ssh_banner"`
	SSHLastActive    map[string]string    `json:"ssh_last_active"`   // user -> last active RFC3339
	ZivpnLastActive  map[string]string    `json:"zivpn_last_active"` // pass -> last active RFC3339
	SSHHandles       map[string]string    `json:"ssh_handles"`       // user -> @handle
	ZivpnHandles     map[string]string    `json:"zivpn_handles"`     // pass -> @handle
	SSHWebSocket     bool                 `json:"ssh_websocket"`     // SSH WebSocket proxy WS/WSS
	SSHBannerTitles    map[string]string    `json:"ssh_banner_titles"` // user -> banner title
	BannerPromoText    string               `json:"banner_promo_text"`
	BannerPromoChannel string               `json:"banner_promo_channel"`
	BannerPromoSupport string               `json:"banner_promo_support"`
	BannerPromoBotName string               `json:"banner_promo_botname"`
	MaxDaysPublic      int                  `json:"max_days_public"`   // Max days for public user creation
	MaxLimitPublic   int                  `json:"max_limit_public"`  // Max device limit for public
	MaxDaysAdmin     int                  `json:"max_days_admin"`    // Max days for admin user creation
	MaxLimitAdmin    int                  `json:"max_limit_admin"`   // Max device limit for admins
	MaxXrayPublic    int                  `json:"max_xray_public"`   // Max VMess accounts for public
	MaxXrayAdmin     int                  `json:"max_xray_admin"`    // Max VMess accounts for admins
	MaxSSHPublic     int                  `json:"max_ssh_public"`    // Max SSH accounts for public
	MaxSSHAdmin      int                  `json:"max_ssh_admin"`     // Max SSH accounts for admins
	MaxZivpnPublic   int                  `json:"max_zivpn_public"`  // Max ZiVPN accounts for public
	MaxZivpnAdmin    int                  `json:"max_zivpn_admin"`   // Max ZiVPN accounts for admins
	BannedUsers      map[string]BannedUserInfo `json:"banned_users"` // chatID -> BannedUserInfo
	SysRXLast        uint64               `json:"sys_rx_last"`
	SysTXLast        uint64               `json:"sys_tx_last"`
	SysRXTotal       uint64               `json:"sys_rx_total"`
	SysTXTotal       uint64               `json:"sys_tx_total"`
	Xray             XrayConfig           `json:"xray"`
	XrayUsers        map[string]XrayUser  `json:"xray_users"` // uuid -> XrayUser data
	AutoReboot       bool                 `json:"auto_reboot"`
	AutoUpdate       bool                 `json:"auto_update"`
	BackupIntervalDays int                  `json:"backup_interval_days"`
	BackupChatID       int64                `json:"backup_chat_id"`
	LocalLastBackup    string               `json:"local_last_backup"`
	Alerts1DaySent     map[string]bool      `json:"alerts_1day_sent"`
	Alerts1HourSent    map[string]bool      `json:"alerts_1hour_sent"`
	Monetization       bool                 `json:"monetization"`
	WebAppURL          string               `json:"webapp_url"`
	ReferredBy         map[int64]int64      `json:"referred_by"`
	ReferralCompleted  map[int64]bool       `json:"referral_completed"`
	ReferralPoints     map[int64]int        `json:"referral_points"`
}

type XrayConfig struct {
	Installed bool   `json:"installed"`
	Port      int    `json:"port"` // usually 10002
}

type XrayUser struct {
	Alias  string `json:"alias"`
	Expire string `json:"expire"` // YYYY-MM-DD
	Owner  string `json:"owner"`  // Chat ID
	Handle string `json:"handle"`
}

type AdminInfo struct {
	Alias      string `json:"alias"`
	Expire     string `json:"expire"`
	FullAccess bool   `json:"full_access"`
}

type BannedUserInfo struct {
	Name   string `json:"name"`
	Reason string `json:"reason"`
	Date   string `json:"date"`
}

type ProxyDTConfig struct {
	Ports map[string]string `json:"ports"`
	Token string            `json:"token"`
}

type SlowDNSConfig struct {
	NS   string `json:"ns"`
	Port string `json:"port"`
	Key  string `json:"key"`
}

type VayDNSConfig struct {
	NS   string `json:"ns"`
	Port string `json:"port"`
	Key  string `json:"key"`
}

type SlipstreamInfo struct {
	NS   string `json:"ns"`
	Port string `json:"port"`
}

var (
	mutex sync.Mutex
	dir   = "/opt/orxtunnel_bot"
)

// SetDir permite cambiar el directorio del DB (util para testing local)
func SetDir(newDir string) {
	dir = newDir
}

// GetDataPath retorna la ruta absoluta del bot_data.json
func GetDataPath() string {
	return filepath.Join(dir, "bot_data.json")
}

// Load lee el archivo bot_data.json o retorna una data por defecto
func Load() (*ConfigData, error) {
	mutex.Lock()
	defer mutex.Unlock()
	return loadUnlocked()
}

func loadUnlocked() (*ConfigData, error) {
	path := GetDataPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return defaultData(), nil
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return defaultData(), err
	}

	var data ConfigData
	err = json.Unmarshal(raw, &data)
	if err != nil {
		return defaultData(), err // Archivo corrupto, reset fallback (en un caso real, haríamos backup)
	}

	// Inicializaciones de seguridad para mapas nulos
	if data.Admins == nil {
		data.Admins = make(map[string]AdminInfo)
	}
	if data.BannedUsers == nil {
		data.BannedUsers = make(map[string]BannedUserInfo)
	}
	if data.SSHOwners == nil {
		data.SSHOwners = make(map[string]string)
	}
	if data.SSHTimeUsers == nil {
		data.SSHTimeUsers = make(map[string]string)
	}
	if data.ZivpnUsers == nil {
		data.ZivpnUsers = make(map[string]string)
	}
	if data.ZivpnOwners == nil {
		data.ZivpnOwners = make(map[string]string)
	}
	if data.ProxyDT.Ports == nil {
		data.ProxyDT.Ports = make(map[string]string)
	}
	if data.SSHLastActive == nil {
		data.SSHLastActive = make(map[string]string)
	}
	if data.ZivpnLastActive == nil {
		data.ZivpnLastActive = make(map[string]string)
	}
	if data.SSHHandles == nil {
		data.SSHHandles = make(map[string]string)
	}
	if data.XrayUsers == nil {
		data.XrayUsers = make(map[string]XrayUser)
	}
	if data.ZivpnHandles == nil {
		data.ZivpnHandles = make(map[string]string)
	}
	if data.SSHBannerTitles == nil {
		data.SSHBannerTitles = make(map[string]string)
	}
	if data.Alerts1DaySent == nil {
		data.Alerts1DaySent = make(map[string]bool)
	}
	if data.Alerts1HourSent == nil {
		data.Alerts1HourSent = make(map[string]bool)
	}
	if data.ReferredBy == nil {
		data.ReferredBy = make(map[int64]int64)
	}
	if data.ReferralCompleted == nil {
		data.ReferralCompleted = make(map[int64]bool)
	}
	if data.ReferralPoints == nil {
		data.ReferralPoints = make(map[int64]int)
	}
	return &data, nil
}

// IsNameTaken revisa si un nombre, password o alias ya está en uso
// en cualquier protocolo para evitar colisiones al eliminar.
func (d *ConfigData) IsNameTaken(target string) bool {
	target = strings.TrimSpace(target)
	
	if _, exists := d.SSHOwners[target]; exists {
		return true
	}
	if _, exists := d.ZivpnOwners[target]; exists {
		return true
	}
	for uid, user := range d.XrayUsers {
		if strings.EqualFold(user.Alias, target) || uid == target {
			return true
		}
	}
	return false
}

// GetMaxDaysPublic returns max days for public users (default 3)
func (d *ConfigData) GetMaxDaysPublic() int {
	if d.MaxDaysPublic <= 0 {
		return 3
	}
	return d.MaxDaysPublic
}

// GetMaxLimitPublic returns max device limit for public users (default 1)
func (d *ConfigData) GetMaxLimitPublic() int {
	if d.MaxLimitPublic <= 0 {
		return 1
	}
	return d.MaxLimitPublic
}

// GetMaxDaysAdmin returns max days for admins (default 7)
func (d *ConfigData) GetMaxDaysAdmin() int {
	if d.MaxDaysAdmin <= 0 {
		return 7
	}
	return d.MaxDaysAdmin
}

// GetMaxLimitAdmin returns max device limit for admins (default 20)
func (d *ConfigData) GetMaxLimitAdmin() int {
	if d.MaxLimitAdmin <= 0 {
		return 20
	}
	return d.MaxLimitAdmin
}

// GetMaxXrayPublic returns max VMess accounts for public users (default 1)
func (d *ConfigData) GetMaxXrayPublic() int {
	if d.MaxXrayPublic <= 0 {
		return 1
	}
	return d.MaxXrayPublic
}

// GetMaxXrayAdmin returns max VMess accounts for admins (default 5)
func (d *ConfigData) GetMaxXrayAdmin() int {
	if d.MaxXrayAdmin <= 0 {
		return 5
	}
	return d.MaxXrayAdmin
}

// GetMaxSSHPublic returns max SSH accounts for public users (default 1)
func (d *ConfigData) GetMaxSSHPublic() int {
	if d.MaxSSHPublic <= 0 {
		return 1
	}
	return d.MaxSSHPublic
}

// GetMaxSSHAdmin returns max SSH accounts for admins (default 5)
func (d *ConfigData) GetMaxSSHAdmin() int {
	if d.MaxSSHAdmin <= 0 {
		return 5
	}
	return d.MaxSSHAdmin
}

// GetMaxZivpnPublic returns max ZiVPN accounts for public users (default 1)
func (d *ConfigData) GetMaxZivpnPublic() int {
	if d.MaxZivpnPublic <= 0 {
		return 1
	}
	return d.MaxZivpnPublic
}

// GetMaxZivpnAdmin returns max ZiVPN accounts for admins (default 5)
func (d *ConfigData) GetMaxZivpnAdmin() int {
	if d.MaxZivpnAdmin <= 0 {
		return 5
	}
	return d.MaxZivpnAdmin
}

// Save guarda la memoria en el archivo bot_data.json
func Save(data *ConfigData) error {
	mutex.Lock()
	defer mutex.Unlock()
	return saveUnlocked(data)
}

// Update encierra una operacion de lectura y escritura en un solo bloqueo concurrente
func Update(fn func(*ConfigData) error) error {
	mutex.Lock()
	defer mutex.Unlock()

	data, err := loadUnlocked()
	if err != nil {
		return err
	}

	if err := fn(data); err != nil {
		return err
	}

	return saveUnlocked(data)
}

func saveUnlocked(data *ConfigData) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	raw, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile(GetDataPath(), raw, 0644)
}

func defaultData() *ConfigData {
	return &ConfigData{
		Admins:        make(map[string]AdminInfo),
		BannedUsers:   make(map[string]BannedUserInfo),
		ExtraInfo:     "Puertos: 22, 80, 443",
		PublicAccess:  true,
		SSHOwners:     make(map[string]string),
		SSHTimeUsers:  make(map[string]string),
		ZivpnUsers:    make(map[string]string),
		ZivpnOwners:   make(map[string]string),
		ProxyDT: ProxyDTConfig{
			Ports: make(map[string]string),
			Token: "dummy",
		},
		SSHLastActive:   make(map[string]string),
		ZivpnLastActive: make(map[string]string),
		SSHHandles:      make(map[string]string),
		ZivpnHandles:    make(map[string]string),
		SSHBannerTitles: make(map[string]string),
		Alerts1DaySent:  make(map[string]bool),
		Alerts1HourSent: make(map[string]bool),
		PublicScanner:   true,
		XrayUsers:       make(map[string]XrayUser),
		AutoReboot:      false,
		AutoUpdate:      false,
		BackupIntervalDays: 0,
		BackupChatID:       0,
		LocalLastBackup:    "",
		Monetization:       false,
		WebAppURL:          "",
		ReferredBy:         make(map[int64]int64),
		ReferralCompleted:  make(map[int64]bool),
		ReferralPoints:     make(map[int64]int),
	}
}

