package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"sync"
)

type Config struct {
	Env    string `yaml:"env" env-default:"local" env-required:"true"`
	Listen struct {
		BindIP string `yaml:"bind_ip" env-default:"127.0.0.1"`
		Port   string `yaml:"port" env-default:"9800"`
		ApiKey string `yaml:"key" env-default:""`
	} `yaml:"listen"`
	SQL struct {
		Enabled  bool   `yaml:"enabled" env-default:"false"`
		Driver   string `yaml:"driver" env-default:"mysql"`
		HostName string `yaml:"hostname" env-default:"localhost"`
		UserName string `yaml:"username" env-default:"root"`
		Password string `yaml:"password" env-default:""`
		Database string `yaml:"database" env-default:""`
		Port     string `yaml:"port" env-default:"8080"`
		Prefix   string `yaml:"prefix" env-default:""`
	} `yaml:"sql"`
	Images struct {
		Path string `yaml:"path" env-default:""`
		Url  string `yaml:"url" env-default:""`
	} `yaml:"images"`
	Cache struct {
		Path string `yaml:"path" env-default:""` // OpenCart file cache directory (DIR_CACHE)
	} `yaml:"cache"`
	Product struct {
		CustomFields []string `yaml:"custom_fields"` // additional allowed custom field names
	} `yaml:"product"`
	// OrderPayment maps payment facts that a payment integration stores in custom
	// columns of the order table. OpenCart has no such columns, so the whole block is
	// optional and every column name is configurable; see docs/config.md.
	OrderPayment OrderPayment `yaml:"order_payment"`
	Telegram     struct {
		Enabled bool   `yaml:"enabled" env-default:"false"`
		ApiKey  string `yaml:"api_key" env-default:""`
	} `yaml:"telegram"`
}

var instance *Config
var once sync.Once

func MustLoad(path string) *Config {
	var err error
	once.Do(func() {
		instance = &Config{}
		if err = cleanenv.ReadConfig(path, instance); err != nil {
			desc, _ := cleanenv.GetDescription(instance, nil)
			err = fmt.Errorf("%s; %s", err, desc)
			instance = nil
			log.Fatal(err)
		}
	})
	return instance
}

// OrderPayment names the order-table columns holding payment data. An empty name means
// the field is not available; an entirely empty block disables payment data.
type OrderPayment struct {
	StatusField      string `yaml:"status_field"`       // payment status, e.g. "paid"
	IdField          string `yaml:"id_field"`           // payment id at the provider
	AmountField      string `yaml:"amount_field"`       // paid amount
	AmountMinorUnits bool   `yaml:"amount_minor_units"` // amount column holds cents
	CurrencyField    string `yaml:"currency_field"`     // ISO currency code of the amount
	TimeField        string `yaml:"time_field"`         // moment the status was reached
	TimeLocation     string `yaml:"time_location"`      // IANA zone of the time column, default UTC
}

// Enabled reports whether at least one payment column is configured.
func (p OrderPayment) Enabled() bool {
	return p.StatusField != "" || p.IdField != "" || p.AmountField != "" ||
		p.CurrencyField != "" || p.TimeField != ""
}
