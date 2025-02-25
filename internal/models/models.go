package models

import (
	"time"

	"gorm.io/gorm"
)

// Invoice repräsentiert eine Rechnung
type Invoice struct {
	gorm.Model
	Date         time.Time
	DueDate      time.Time
	TotalAmount  float64
	VatAmount    float64
	Currency     string
	SupplierID   uint
	Supplier     Supplier
	Category     string
	Status       string
	InvoiceItems []InvoiceItem
	FileRefs     []FileReference
}

// InvoiceItem repräsentiert eine Position auf der Rechnung
type InvoiceItem struct {
	gorm.Model
	InvoiceID    uint
	Description  string
	Quantity     float64
	SinglePrice  float64
	TotalPrice   float64
}

// Supplier repräsentiert einen Lieferanten
type Supplier struct {
	gorm.Model
	Name         string
	Address      string
	TaxNumber    string
	ContactInfo  string
	Invoices     []Invoice
}

// FileReference speichert Informationen zur originalen Rechnungsdatei
type FileReference struct {
	gorm.Model
	InvoiceID    uint
	OriginalPath string
	Filename     string
	Hash         string
}