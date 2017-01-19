package consul

import (
	"context"
	"strings"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/hashicorp/terraform/backend"
	"github.com/hashicorp/terraform/helper/schema"
)

func New() backend.Backend {
	return &schema.Backend{
		Schema: map[string]*schema.Schema{
			"path": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Path to store state in Consul",
			},

			"access_token": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Access token for a Consul ACL",
			},

			"address": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Address to Consul",
			},

			"scheme": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Scheme to communicate to Consul with",
			},

			"datacenter": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Datacenter to communicate with",
			},

			"http_auth": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "HTTP Auth in the format of 'username:password'",
			},
		},

		ConfigureFunc: configureFunc,
	}
}

type backend struct {
	*schema.Backend

	client *remote.State
}

func (b *backend) State() (state.State, error) {
	return b.client, nil
}

func (b *backend) configure(ctx context.Context) error {
	// Grab the resource data
	data := schema.FromContextBackendConfig(ctx)

	// Configure the client
	config := consulapi.DefaultConfig()
	if v, ok := data.GetOk("access_token"); ok {
		config.Token = v.(string)
	}
	if v, ok := data.GetOk("address"); ok {
		config.Address = v.(string)
	}
	if v, ok := data.GetOk("scheme"); ok && v.(string) != "" {
		config.Scheme = v.(string)
	}
	if v, ok := data.GetOk("datacenter"); ok && v.(string) != "" {
		config.Datacenter = v.(string)
	}
	if v, ok := data.GetOk("http_auth"); ok && v.(string) != "" {
		auth := v.(string)

		var username, password string
		if strings.Contains(auth, ":") {
			split := strings.SplitN(auth, ":", 2)
			username = split[0]
			password = split[1]
		} else {
			username = auth
		}

		config.HttpAuth = &consulapi.HttpBasicAuth{
			Username: username,
			Password: password,
		}
	}

	client, err := consulapi.NewClient(config)
	if err != nil {
		return err
	}

	b.client = &remote.ConsulClient{
		Client: client,
		Path:   path,
	}

	return nil
}
