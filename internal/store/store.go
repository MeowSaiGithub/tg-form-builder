package store

import (
	"fmt"
	"go-tg-support-ticket/form"
	"go-tg-support-ticket/internal/database"
)

func init() {
	Store = store{}
	Tickets = ticketObj{}
}

var adp database.Adaptor
var availableAdapters = make(map[string]database.Adaptor)
var enabled bool

type PersistenceStorageInterface interface {
	Open(cfg *database.Config) error
	Migrate(schema *form.Form) error
}

var Store PersistenceStorageInterface

type store struct {
}

func (store) Open(cfg *database.Config) error {
	enabled = cfg.Enable
	if enabled {
		return openAdaptor(cfg)
	}
	return nil
}

func (store) Migrate(schema *form.Form) error {
	if enabled {
		return adp.Migrate(schema)
	}
	return fmt.Errorf("database persistence is not enabled")
}

func openAdaptor(cfg *database.Config) error {
	if ad, ok := availableAdapters[cfg.UseAdaptor]; ok {
		adp = ad
	} else {
		return fmt.Errorf("store: %s adapter is not available in this binary", cfg.UseAdaptor)
	}

	dsn, err := database.ParseConfig(cfg)
	if err != nil {
		return fmt.Errorf("store: failed to parse %s adaptor config: %w", cfg.UseAdaptor, err)
	}

	return adp.Open(dsn)
}

func RegisterAdaptor(a database.Adaptor) {
	if a == nil {
		panic("store: Register adaptor is nil")
	}
	adapterName := a.GetName()
	if _, ok := availableAdapters[adapterName]; ok {
		panic("store: adaptor " + adapterName + " is already registered")
	}
	availableAdapters[adapterName] = a
}

type TicketPersistence interface {
	Create(tableName string, fields []form.Field) error
}

var Tickets TicketPersistence

type ticketObj struct{}

func (ticketObj) Create(tableName string, fields []form.Field) error {
	if enabled {
		return adp.InsertUserInputs(tableName, fields)
	}
	return nil
}
