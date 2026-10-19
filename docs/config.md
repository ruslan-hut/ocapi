## Command-line flags
- `-conf` — path to the config file (default `config.yml`)
- `-log` — full log file name (default `/var/log/ocapi.log`); ignored when `env: local`, which logs to stdout

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
## OpenCart file cache
cache:
  path: /var/www/site/system/storage/cache/ # DIR_CACHE of the site; empty disables cache clearing
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

### Startup migrations
On start, the service prepares the OpenCart schema:
- adds the `*_uid` columns used to address products, categories, attributes and images;
- converts `product_special.date_start`/`date_end` from `DATE` to `DATETIME NOT NULL DEFAULT '0000-00-00 00:00:00'` (existing values become midnight);
- creates the `customer_product_special` table for personal customer prices.

Database sessions use `sql_mode = 'NO_ZERO_IN_DATE,NO_ENGINE_SUBSTITUTION'`, as OpenCart does, so zero dates are accepted.

### Cache
After `product/special/sync` the service removes OpenCart file cache entries with product prices (`cache.product.*`, `cache.catproductany.*`, `cache.upstore_productcategorytabs*`, `cache.theme*`) from `cache.path`.
