package hcat

import (
	"net/http"
	"os"
	"time"

	dep "github.com/hashicorp/hcat/internal/dependency"
)

// Looker is an interface for looking up data from Consul, Vault and the
// Environment.
type Looker interface {
	dep.Clients
	Env() []string
	Stop()
}

// ClientSet focuses only on external (consul/vault) dependencies
// at this point so we extend it here to include environment variables to meet
// the looker interface.
type ClientSet struct {
	*dep.ClientSet
	injectedEnv []string
}

// NewClientSet is used to create the clients used.
// Fulfills the Looker interface.
func NewClientSet() *ClientSet {
	return &ClientSet{
		ClientSet:   dep.NewClientSet(),
		injectedEnv: []string{},
	}
}

// AddConsul creates a Consul client and adds to the client set
func (cs *ClientSet) AddConsul(i ConsulInput) *ClientSet {
	cs.CreateConsulClient(i.toInternal())
	return cs
}

// AddVault creates a Vault client and adds to the client set
func (cs *ClientSet) AddVault(i VaultInput) *ClientSet {
	cs.CreateVaultClient(i.toInternal())
	return cs
}

// Stop closes all idle connections for any attached clients and clears
// the list of injected environment variables.
func (cs *ClientSet) Stop() {
	if cs.ClientSet != nil {
		cs.ClientSet.Stop()
	}
	cs.injectedEnv = []string{}
}

// InjectEnv adds "key=value" pairs to the environment used for template
// evaluations and child process runs. Note that this is in addition to the
// environment running consul template and in the case of duplicates, the
// last entry wins.
func (cs *ClientSet) InjectEnv(env ...string) {
	cs.injectedEnv = append(cs.injectedEnv, env...)
}

// You should do any messaging of the Environment variables during startup
// As this will just use the raw Environment.
func (cs *ClientSet) Env() []string {
	return append(os.Environ(), cs.injectedEnv...)
}

// Input wrappers around internal structure. Going to rework the internal
// structure, so this abstracts that away to make that workable.

// VaultInput defines the inputs needed to configure the Vault client.
type VaultInput struct {
	Address     string
	Namespace   string
	Token       string
	UnwrapToken bool
	Transport   TransportInput
	// optional, principally for testing
	HttpClient *http.Client
}

func (i VaultInput) toInternal() *dep.CreateClientInput {
	cci := &dep.CreateClientInput{
		Address:     i.Address,
		Namespace:   i.Namespace,
		Token:       i.Token,
		UnwrapToken: i.UnwrapToken,
	}
	return i.Transport.toInternal(cci)
}

// ConsulInput defines the inputs needed to configure the Consul client.
type ConsulInput struct {
	Address      string
	Namespace    string
	Token        string
	AuthEnabled  bool
	AuthUsername string
	AuthPassword string
	Transport    TransportInput
	// optional, principally for testing
	HttpClient *http.Client
}

func (i ConsulInput) toInternal() *dep.CreateClientInput {
	cci := &dep.CreateClientInput{
		Address:      i.Address,
		Namespace:    i.Namespace,
		Token:        i.Token,
		AuthEnabled:  i.AuthEnabled,
		AuthUsername: i.AuthUsername,
		AuthPassword: i.AuthPassword,
	}
	return i.Transport.toInternal(cci)
}

type TransportInput struct {
	// Transport/TLS
	SSLEnabled bool
	SSLVerify  bool
	SSLCert    string
	SSLKey     string
	SSLCACert  string
	SSLCAPath  string
	ServerName string

	DialKeepAlive       time.Duration
	DialTimeout         time.Duration
	DisableKeepAlives   bool
	IdleConnTimeout     time.Duration
	MaxIdleConns        int
	MaxIdleConnsPerHost int
	TLSHandshakeTimeout time.Duration
}

func (i TransportInput) toInternal(cci *dep.CreateClientInput) *dep.CreateClientInput {
	cci.SSLEnabled = i.SSLEnabled
	cci.SSLVerify = i.SSLVerify
	cci.SSLCert = i.SSLCert
	cci.SSLKey = i.SSLKey
	cci.SSLCACert = i.SSLCACert
	cci.SSLCAPath = i.SSLCAPath
	cci.ServerName = i.ServerName
	cci.TransportDialKeepAlive = i.DialKeepAlive
	cci.TransportDialTimeout = i.DialTimeout
	cci.TransportDisableKeepAlives = i.DisableKeepAlives
	cci.TransportIdleConnTimeout = i.IdleConnTimeout
	cci.TransportMaxIdleConns = i.MaxIdleConns
	cci.TransportMaxIdleConnsPerHost = i.MaxIdleConnsPerHost
	cci.TransportTLSHandshakeTimeout = i.TLSHandshakeTimeout
	return cci
}
