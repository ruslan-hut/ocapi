## Config file structure
The configuration file is a YAML file that contains the following sections:
```yaml
---

env: local               # Environment for logging
## OCAPI service configuration
listen:
  bind_ip: 127.0.0.1     # IP address to bind the service
  port: 9800             # Port to listen
  key: api-key           # API key for the OCAPI service
## OpenCart database connection
sql:                     
  enabled: false         # Enable or disable SQL connection
  driver: mysqli         # Database driver
  hostname: localhost    # Database hostname
  username: username     # Database username
  password: password     # Database password
  database: db           # Database name
  port: 8080             # Database port
  prefix: prefix_        # Database table prefix
## Product images
images:
  path: /path/to/images/ # Path to the images directory on the server
  url: catalog/product/  # URL to the images directory
## Product settings
product:
  custom_fields:         # Additional allowed custom field names (beyond defaults)
    - points             # Example: allow updating 'points' column
    - sort_order         # Example: allow updating 'sort_order' column
```

### Custom Fields
By default, the following product columns can be updated via the `custom_fields` API parameter:
- `sku`, `upc`, `ean`, `jan`, `isbn`, `mpn`, `location`

To allow additional columns, add them to `product.custom_fields` in the config file.