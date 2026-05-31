# AutoHttps

AutoHttps is a proxy in a Docker image to easily allow https access to a web application or REST API. 

* 🔒 Generates valid SSL certificates using [Let's Encrypt](https://letsencrypt.org/) and keep it rotated when needed. 
* 🌐 Can be used with a domain name or just with the public IP (thaks to [sslip.io](https://sslip.io)).
* ⚡ Is designed to be very easy to configure in a docker-compose deployment.
* 🚀 Can be used in production or in development.

## Table of Contents

- [How to use](#how-to-use)
  - [Without domain](#without-domain)
  - [With domain](#with-domain)
  - [With multiple web applications](#with-multiple-web-applications)
- [Configuration](#configuration)
  - [Environment Variables](#environment-variables)
  - [Volumes](#volumes)
- [Limitations](#limitations)
  - [SSLIP.io Considerations](#sslipio-considerations)  
- [Examples](#examples)
- [Project Warning](#project-warning)

## How to use

### Without domain

1. SSH connect to a server with a public IP address and open ports:
   - Port 80 (HTTP): Required for initial Let's Encrypt verification
   - Port 443 (HTTPS): For secure traffic to your application

2. Install Docker and Docker Compose.

3. Create a `docker-compose.yaml` file with two services:
   - Your application (serving plain HTTP)
   - The AutoHttps proxy (handling HTTPS)

```yaml
services:
  # The AutoHttps proxy service
  autohttps:
    image: cruizba/autohttps:latest
    ports:
      - "80:80"    # Required for Let's Encrypt verification
      - "443:443"  # Your users will connect here
    volumes:
      - ./caddy_data:/data  # Store certificates persistently
    environment:
      - SERVICES=web:3000  # Point to your app's service name and port
    depends_on:
      - web

  # Your web application service
  web:
    image: your-web-image
    # IMPORTANT: Your app should:
    # 1. Serve plain HTTP (not HTTPS)
    # 2. Listen on port 3000 (any port is fine, just match it with
    #    the port used in SERVICES environment variable)
    # 3. No need to expose ports - AutoHttps will handle that
```

4. Start everything:
```bash
docker compose up -d
```
Your application will be available via HTTPS at `https://web-YOUR-IP.sslip.io`, where `YOUR-IP` is your server's public IP address formatted with dashes instead of dots.

For example, if your server's IP is `1.2.3.4`, the URL will be `https://web-1-2-3-4.sslip.io`

### With domain

If you have a domain name pointing to the public IP, set it with the dedicated
`DOMAIN` environment variable:

```yaml
services:
  autohttps:
    image: cruizba/autohttps:latest
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./caddy_data:/data
    environment:
      - SERVICES=web:3000
      - DOMAIN=www.yourdomain.com
    depends_on:
      - web

  web:
    image: your-web-image
```
Your application will be available via HTTPS at `https://www.yourdomain.com`.

Keeping the domain in its own variable is handy when you publish a
`docker-compose.yml` (e.g. in a registry) and want users to set the domain at
deploy time **without editing the compose file**. Use a substitution with an
empty default and let whoever deploys provide a `.env` file:

```yaml
    environment:
      - SERVICES=web:3000
      - DOMAIN=${DOMAIN:-}   # falls back to sslip.io when unset
```

```bash
# .env (provided at deploy time)
DOMAIN=www.yourdomain.com
```

If `DOMAIN` is left empty (or the `.env` file is omitted), AutoHttps falls back
to a sslip.io domain automatically, so the same template works with or without a
custom domain. `DOMAIN` is only valid with a single service; for multiple
applications set each domain inline in `SERVICES` (see below).

### With multiple web applications

You can secure multiple applications by listing them in the `SERVICES` variable:

```yaml
services:
  autohttps:
    image: cruizba/autohttps:latest
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./caddy_data:/data
    environment:
      # Format: service1:port,service2:port
      # Or with custom domains: service1:port=domain1.com,service2:port=domain2.com
      - SERVICES=app1:3000,app2:8080
    depends_on:
      - app1
      - app2

  app1:
    image: your-app1-image

  app2:
    image: your-app2-image
```
## Configuration

### Environment Variables

- `SERVICES`: A comma-separated list of services in the format:
  ```
  serviceName:port[=domain.com][,anotherService:port]
  ```
  If the domain is omitted (or left empty, e.g. `web:3000=`), a domain will be
  automatically generated using sslip.io with your server's public IP in the
  format `service-name-PUBLIC-IP.sslip.io`.

- `DOMAIN` (optional): Sets the domain for a **single-service** deployment
  without having to edit `SERVICES`. Useful for publishable docker-compose
  templates: keep `SERVICES=web:3000` in the compose file and let users pass
  `DOMAIN` via a `.env` file. When `DOMAIN` is unset or empty, AutoHttps falls
  back to a sslip.io domain. It cannot be combined with multiple services or
  with an inline domain in `SERVICES` (set each domain inline instead).

### Volumes

AutoHttps uses two possible volume mounts:

1. **`./caddy_data:/data` (Required)**
   - Stores the SSL/TLS certificates and other Caddy data
   - Without this volume, certificates will be regenerated on every restart (hitting Let's Encrypt rate limits)   

2. **`./caddy_config:/config` (Optional)**
   - Stores the generated Caddyfile configuration
   - Only mount this if you need to customize the Caddy configuration
   - **Important:** When this volume is mounted:
     - The Caddyfile is generated only if the directory is empty
     - Changes to the `SERVICES` environment variable won't update the Caddyfile
     - Manual updates to the Caddyfile are required if you modify the `SERVICES` environment variable after initial creation

## Limitations

### SSLIP.io Considerations

1. **Rate Limits**
   - Do not abuse of sslip.io service, they actually can handle up to  [10k domains](https://github.com/cunnie/sslip.io/issues/57#issuecomment-2439742710), but it is a free service maintained by volunteers.

2. **Domain availability:**
   - sslip.io domains are public and shared
   - If someone has misused your IP-based domain, it might be temporarily blocked
   - It is recommended to use domains for production environments

2. **DNS resolution:**
   - sslip.io service might experience occasional downtime
   - DNS resolution depends on the sslip.io service availability

## Examples

You can find a complete working examples in the `example` directory. To test it:

1. Create a VM with public IP and open ports 80 and 443.
2. Install Docker and Docker Compose.
3. Execute the following commands:
```
git clone https://github.com/cruizba/AutoHttps
cd AutoHttps/example
docker compose up -d
```

You will have two applications available at:

- `https://random-cats-YOUR-IP.sslip.io`
- `https://random-dogs-YOUR-IP.sslip.io`

For example, if your server's IP is `1.2.3.4`, the URLs will be:

- `https://random-cats-1-2-3-4.sslip.io`
- `https://random-dogs-1-2-3-4.sslip.io`

## Project Warning

> [!WARNING]
> I am not responsible for any misuse of this tool. Do not use AutoHttps for bad purposes, as misuse can lead to domain blacklisting and other issues, deteriorating the service for everyone. Always use this tool responsibly and ethically.
