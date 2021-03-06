// Code generated by entc, DO NOT EDIT.

package migrate

import (
	"entgo.io/ent/dialect/sql/schema"
	"entgo.io/ent/schema/field"
)

var (
	// WireguardClientsColumns holds the columns for the "wireguard_clients" table.
	WireguardClientsColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "created_at", Type: field.TypeTime},
		{Name: "updated_at", Type: field.TypeTime},
		{Name: "destroyed_at", Type: field.TypeTime},
		{Name: "expired", Type: field.TypeBool, Default: false},
		{Name: "identify", Type: field.TypeString, Default: ""},
		{Name: "server_private_key", Type: field.TypeString, Unique: true},
		{Name: "client_private_key", Type: field.TypeString, Unique: true},
		{Name: "netif_name", Type: field.TypeString, Unique: true, Size: 14},
		{Name: "peer_addr", Type: field.TypeString, Default: ""},
		{Name: "listen_addr", Type: field.TypeString},
		{Name: "alloc_cidr", Type: field.TypeString},
		{Name: "server_address", Type: field.TypeString},
		{Name: "client_address", Type: field.TypeString},
		{Name: "receive_bytes", Type: field.TypeUint64, Default: 0},
		{Name: "transmit_bytes", Type: field.TypeUint64, Default: 0},
	}
	// WireguardClientsTable holds the schema information for the "wireguard_clients" table.
	WireguardClientsTable = &schema.Table{
		Name:       "wireguard_clients",
		Columns:    WireguardClientsColumns,
		PrimaryKey: []*schema.Column{WireguardClientsColumns[0]},
	}
	// Tables holds all the tables in the schema.
	Tables = []*schema.Table{
		WireguardClientsTable,
	}
)

func init() {
}
