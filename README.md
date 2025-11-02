# AutoHttps

AutoHttps solves the problem of easily deploying web applications with HTTPS in development environments without the need to manage certificates or domains. Perfect for developers who want to test their applications with HTTPS locally or in quick prototypes and home labs, without the hassle of certificate management or domain configuration.

No domain name is required: AutoHttps will automatically generate a domain based on your server's public IP address using the [sslip.io](https://sslip.io) service. However, you can also use your own domain if you have one.

## Features

- 🔒 Automatic HTTPS for your applications without manual certificate management
- 🌐 No domain required - automatic domain generation based on your IP
- 🚀 Simple configuration through environment variables
- 🐳 Docker-ready with docker-compose support
- ⚡ Zero-configuration setup

## Table of Contents

- [Basic Usage](#basic-usage)
- [Advanced Usage](#advanced-usage)
  - [Using Custom Domains](#using-custom-domains)
  - [Multiple Applications](#multiple-applications)
- [Configuration](#configuration)
  - [Environment Variables](#environment-variables)
  - [Volumes](#volumes)
- [Limitations](#limitations)
  - [Let's Encrypt Rate Limits](#lets-encrypt-rate-limits)
  - [SSLIP.io Considerations](#sslipio-considerations)
  - [Security Notes](#security-notes)
- [Examples](#examples)
- [Project Warning](#project-warning)

## Basic Usage

Let's walk through securing a simple web application with HTTPS:

1. Create a VM or server with a public IP address and open ports:
   - Port 80 (HTTP): Required for initial Let's Encrypt verification
   - Port 443 (HTTPS): For secure traffic to your application

2. Install Docker and Docker Compose on your server.

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
      - SERVICES=myapp:3000  # Point to your app's service name and port
    depends_on:
      - myapp

  # Your application service
  myapp:
    image: your-app-image
    # IMPORTANT: Your app should:
    # 1. Serve plain HTTP (not HTTPS)
    # 2. Listen on port 3000 (any port is fine, just match it with
    #    the port used in SERVICES)
    # 3. No need to expose ports - AutoHttps will handle that
```

4. Start everything:
```bash
docker compose up -d
```

Your application will be available via HTTPS at `https://myapp-YOUR-IP.sslip.io`, where `YOUR-IP` is your server's public IP address formatted with dashes instead of dots.

For example, if your server's IP is `1.2.3.4`, the URL will be `https://myapp-1-2-3-4.sslip.io`

## Advanced Usage

### Using Custom Domains

If you have your own domain names, you can use them instead of the auto-generated ones. Make sure your domain's DNS records point to your server's IP address:

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
      - SERVICES=myapp:3000=myapp.yourdomain.com
    depends_on:
      - myapp

  myapp:
    image: your-app-image
```

### Multiple Applications

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
    image: your-first-app-image

  app2:
    image: your-second-app-image
```

## Configuration

### Environment Variables

- `SERVICES`: A comma-separated list of services in the format:
  ```
  serviceName:port[=domain.com][,anotherService:port]
  ```
  If the domain is omitted, a domain will be automatically generated using sslip.io with your server's public IP in the format `service-name-PUBLIC-IP.sslip.io`.

### Volumes

AutoHttps uses two possible volume mounts:

1. **`./caddy_data:/data` (Required)**
   - Stores the SSL/TLS certificates and other Caddy data
   - Must be persistent to avoid hitting Let's Encrypt rate limits
   - Without this volume, certificates will be regenerated on every restart

2. **`./caddy_config:/config` (Optional)**
   - Stores the generated Caddyfile configuration
   - Only mount this if you need to customize the Caddy configuration
   - **Important:** When this volume is mounted:
     - The Caddyfile is generated only if the directory is empty
     - Changes to the `SERVICES` environment variable won't update the Caddyfile
     - Manual updates to the Caddyfile are required if you modify the `SERVICES` environment variable

## Limitations

### Let's Encrypt Rate Limits

   - Do not abuse of sslip.io service, they actually can handle up to  [10k domains](https://github.com/cunnie/sslip.io/issues/57#issuecomment-2439742710), but it is a free service maintained by volunteers.

### SSLIP.io Considerations

1. **Domain availability:**
   - sslip.io domains are public and shared
   - If someone has misused your IP-based domain, it might be temporarily blocked
   - Consider using custom domains for production environments

2. **DNS resolution:**
   - sslip.io service might experience occasional downtime
   - DNS resolution depends on the sslip.io service availability

### Security Notes

1. AutoHttps is designed for development and testing environments
2. For production use consider using custom domains

## Examples

You can find a complete working examples in the `example` directory. To test it:

1. Create a VM with public IP and open ports 80 and 443.
2. Install Docker and Docker Compose.
3. Execute the following commands:
```
git clone https://github.com/cruizba/AutoHttps
cd example
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
