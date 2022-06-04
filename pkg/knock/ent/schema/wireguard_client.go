package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"time"
)

// WireguardClient holds the schema definition for the WireguardClient entity.
type WireguardClient struct {
	ent.Schema
}

// Fields of the User.
func (WireguardClient) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
		field.Time("destroyed_at").Default(func() time.Time { return time.Unix(0, 0) }),
		field.Bool("expired").Default(false),
		field.String("identify").Default("").NotEmpty(),
		field.String("server_private_key").NotEmpty().Unique(),
		field.String("client_private_key").NotEmpty().Unique(),
		field.String("netif_name").NotEmpty().MaxLen(14).Unique(),
		field.String("peer_addr"),
		field.String("listen_addr").NotEmpty(),
		field.String("alloc_cidr").NotEmpty(),
		field.String("server_address").NotEmpty(),
		field.String("client_address").NotEmpty(),
		field.Uint64("receive_bytes"),
		field.Uint64("transmit_bytes"),
	}
}

// Edges of the User.
func (WireguardClient) Edges() []ent.Edge {
	return nil
}
