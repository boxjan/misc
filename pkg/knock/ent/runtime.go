//go:build tools
// +build tools

// Code generated by entc, DO NOT EDIT.

package ent

import (
	"time"

	"github.com/boxjan/misc/pkg/knock/ent/schema"
	"github.com/boxjan/misc/pkg/knock/ent/wireguardclient"
)

// The init function reads all schema descriptors with runtime code
// (default values, validators, hooks and policies) and stitches it
// to their package variables.
func init() {
	wireguardclientFields := schema.WireguardClient{}.Fields()
	_ = wireguardclientFields
	// wireguardclientDescCreatedAt is the schema descriptor for created_at field.
	wireguardclientDescCreatedAt := wireguardclientFields[0].Descriptor()
	// wireguardclient.DefaultCreatedAt holds the default value on creation for the created_at field.
	wireguardclient.DefaultCreatedAt = wireguardclientDescCreatedAt.Default.(func() time.Time)
	// wireguardclientDescUpdatedAt is the schema descriptor for updated_at field.
	wireguardclientDescUpdatedAt := wireguardclientFields[1].Descriptor()
	// wireguardclient.DefaultUpdatedAt holds the default value on creation for the updated_at field.
	wireguardclient.DefaultUpdatedAt = wireguardclientDescUpdatedAt.Default.(func() time.Time)
	// wireguardclient.UpdateDefaultUpdatedAt holds the default value on update for the updated_at field.
	wireguardclient.UpdateDefaultUpdatedAt = wireguardclientDescUpdatedAt.UpdateDefault.(func() time.Time)
	// wireguardclientDescDestroyedAt is the schema descriptor for destroyed_at field.
	wireguardclientDescDestroyedAt := wireguardclientFields[2].Descriptor()
	// wireguardclient.DefaultDestroyedAt holds the default value on creation for the destroyed_at field.
	wireguardclient.DefaultDestroyedAt = wireguardclientDescDestroyedAt.Default.(func() time.Time)
	// wireguardclient.UpdateDefaultDestroyedAt holds the default value on update for the destroyed_at field.
	wireguardclient.UpdateDefaultDestroyedAt = wireguardclientDescDestroyedAt.UpdateDefault.(func() time.Time)
	// wireguardclientDescExpired is the schema descriptor for expired field.
	wireguardclientDescExpired := wireguardclientFields[3].Descriptor()
	// wireguardclient.DefaultExpired holds the default value on creation for the expired field.
	wireguardclient.DefaultExpired = wireguardclientDescExpired.Default.(bool)
	// wireguardclientDescIdentify is the schema descriptor for identify field.
	wireguardclientDescIdentify := wireguardclientFields[4].Descriptor()
	// wireguardclient.DefaultIdentify holds the default value on creation for the identify field.
	wireguardclient.DefaultIdentify = wireguardclientDescIdentify.Default.(string)
	// wireguardclient.IdentifyValidator is a validator for the "identify" field. It is called by the builders before save.
	wireguardclient.IdentifyValidator = wireguardclientDescIdentify.Validators[0].(func(string) error)
	// wireguardclientDescServerPrivateKey is the schema descriptor for server_private_key field.
	wireguardclientDescServerPrivateKey := wireguardclientFields[5].Descriptor()
	// wireguardclient.ServerPrivateKeyValidator is a validator for the "server_private_key" field. It is called by the builders before save.
	wireguardclient.ServerPrivateKeyValidator = wireguardclientDescServerPrivateKey.Validators[0].(func(string) error)
	// wireguardclientDescClientPrivateKey is the schema descriptor for client_private_key field.
	wireguardclientDescClientPrivateKey := wireguardclientFields[6].Descriptor()
	// wireguardclient.ClientPrivateKeyValidator is a validator for the "client_private_key" field. It is called by the builders before save.
	wireguardclient.ClientPrivateKeyValidator = wireguardclientDescClientPrivateKey.Validators[0].(func(string) error)
	// wireguardclientDescNetifName is the schema descriptor for netif_name field.
	wireguardclientDescNetifName := wireguardclientFields[7].Descriptor()
	// wireguardclient.NetifNameValidator is a validator for the "netif_name" field. It is called by the builders before save.
	wireguardclient.NetifNameValidator = func() func(string) error {
		validators := wireguardclientDescNetifName.Validators
		fns := [...]func(string) error{
			validators[0].(func(string) error),
			validators[1].(func(string) error),
		}
		return func(netif_name string) error {
			for _, fn := range fns {
				if err := fn(netif_name); err != nil {
					return err
				}
			}
			return nil
		}
	}()
	// wireguardclientDescListenAddr is the schema descriptor for listen_addr field.
	wireguardclientDescListenAddr := wireguardclientFields[9].Descriptor()
	// wireguardclient.ListenAddrValidator is a validator for the "listen_addr" field. It is called by the builders before save.
	wireguardclient.ListenAddrValidator = wireguardclientDescListenAddr.Validators[0].(func(string) error)
	// wireguardclientDescAllocCidr is the schema descriptor for alloc_cidr field.
	wireguardclientDescAllocCidr := wireguardclientFields[10].Descriptor()
	// wireguardclient.AllocCidrValidator is a validator for the "alloc_cidr" field. It is called by the builders before save.
	wireguardclient.AllocCidrValidator = wireguardclientDescAllocCidr.Validators[0].(func(string) error)
	// wireguardclientDescServerAddress is the schema descriptor for server_address field.
	wireguardclientDescServerAddress := wireguardclientFields[11].Descriptor()
	// wireguardclient.ServerAddressValidator is a validator for the "server_address" field. It is called by the builders before save.
	wireguardclient.ServerAddressValidator = wireguardclientDescServerAddress.Validators[0].(func(string) error)
	// wireguardclientDescClientAddress is the schema descriptor for client_address field.
	wireguardclientDescClientAddress := wireguardclientFields[12].Descriptor()
	// wireguardclient.ClientAddressValidator is a validator for the "client_address" field. It is called by the builders before save.
	wireguardclient.ClientAddressValidator = wireguardclientDescClientAddress.Validators[0].(func(string) error)
}
