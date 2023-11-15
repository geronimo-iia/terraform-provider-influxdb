package influxdb

import (
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/influxdata/influxdb/client"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"influxdb_database":         resourceDatabase(),
			"influxdb_user":             resourceUser(),
			"influxdb_continuous_query": resourceContinuousQuery(),
		},

		Schema: map[string]*schema.Schema{
			"url": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Influxdb connection url",
				DefaultFunc: schema.EnvDefaultFunc(
					"INFLUXDB_URL", "http://localhost:8086/",
				),
			},
			"username": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Influxdb user name",
				DefaultFunc: schema.EnvDefaultFunc("INFLUXDB_USERNAME", ""),
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Influxdb password",
				Sensitive:   true,
				StateFunc:   hashSum,
				DefaultFunc: schema.EnvDefaultFunc("INFLUXDB_PASSWORD", ""),
			},
			"skip_ssl_verify": {
				Type:        schema.TypeBool,
				Description: "skip ssl verify on connection",
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("INFLUXDB_SKIP_SSL_VERIFY", "0"),
			},
		},

		ConfigureFunc: configure,
	}
}

func configure(d *schema.ResourceData) (interface{}, error) {
	url, err := url.Parse(d.Get("url").(string))
	if err != nil {
		return nil, fmt.Errorf("invalid InfluxDB URL: %w", err)
	}

	config := client.Config{
		URL:       *url,
		Username:  d.Get("username").(string),
		Password:  d.Get("password").(string),
		UnsafeSsl: d.Get("skip_ssl_verify").(bool),
	}

	conn, err := client.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("error connecting server: %w", err)
	}

	// assume that an InfluxBD is already provision when using the InfluxDB provider.
	// you have to manage dependency between your modules
	_, version, err := conn.Ping()
	if err != nil {
		return nil, fmt.Errorf("error connecting server: %w", err)
	}

	if version == "" {
		return nil, fmt.Errorf("error connecting server: no version information %s", version)
	}

	return conn, nil
}

func exec(conn *client.Client, query string) error {
	resp, err := conn.Query(client.Query{
		Command: query,
	})
	if err != nil {
		return err
	}
	if resp.Err != nil {
		return resp.Err
	}
	return nil
}
