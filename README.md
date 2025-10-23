# AutoHttps

AutoHttps is a minimalistic Caddy Proxy Container Image that enables easy deployment of web applications with automatic HTTPS using Let's Encrypt certificates and [sslip.io](https://sslip.io/). It provides a zero-configuration approach to setup SSL/TLS for your services making it ideal for development and quick prototyping or home labs.

## Features

- 🔒 Automatic HTTPS with Let's Encrypt certificates
- 🌐 Automatic domain generation using sslip.io.
- 🚀 Simple configuration through environment variables
- 🐳 Docker-ready with docker-compose support
- ⚡ Zero-configuration SSL/TLS setup

## Example Quick Start

1. Create a VM or server with a public IP address and ports 80 and 443 open with Linux.
2. Install Docker and Docker Compose on your server if not already installed.
3. Execute the following commands to see how AutoHttps works with example applications:

```bash
git clone https://github.com/cruizba/AutoHttps
cd AutoHttps
docker compose up -d
docker compose logs -f autohttps
```

Check the logs of the `autohttps` service and navigate to the generated domains for the example applications:

- Random Cats: `http://random-cats-YOUR_IP.sslip.io`
- Random Dogs: `http://random-dogs-YOUR_IP.sslip.io`

For example, if your public IP is `1.2.3.4` the URLs would be:

- Random Cats: `http://random-cats-1-2-3-4.sslip.io`
- Random Dogs: `http://random-dogs-1-2-3-4.sslip.io`

## Real Scenarios

### Without custom domains

If you have a VM or server with a public IP address, you can use AutoHttps to automatically generate domains using sslip.io.

1. Create a VM or server with a public IP address and ports 80 and 443 open.
2. Create a `docker-compose.yaml` file with autohttps service and your applications:

```yaml
services:
  autohttps:
    image: cruizba/autohttps:latest
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./caddy_data:/data
      - ./caddy_config:/config
    environment:
      - SERVICES=myapp1:3000,myapp2:3000
    depends_on:
      - myapp
      - api

    myapp1:
      image: your-app-image
      ... your app configuration ...

    myapp2:
      image: your-app-image-2
      ... your app configuration ...
```

2. Run your services:

```bash
docker-compose up -d
```

3. Check the logs of the `autohttps` service to see the generated domains for your applications.

### With custom domains

If you have your own domain names, you can configure AutoHttps to use them. Make sure your domain's DNS records point to the public IP address of your server. You need to create a domain for each service and specify it in the `SERVICES` environment variable.

1. Create a VM or server with a public IP address and ports 80 and 443 open.
2. Create a `docker-compose.yaml` file with autohttps service and your applications:

```yaml
services:
  autohttps:
    image: cruizba/autohttps:latest
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./caddy_data:/data
      - ./caddy_config:/config
    environment:
      - SERVICES=myapp1:3000=myapp1.domain.com,myapp2:3000=myapp2.domain.com
    depends_on:
      - myapp
      - api

    myapp1:
      image: your-app-image
      ... your app configuration ...

    myapp2:
      image: your-app-image-2
      ... your app configuration ...
```

3. Run your services:

```bash
docker-compose up -d
```

4. Access your applications using your custom domain names.

## Configuration

### Environment Variables

- `SERVICES`: A comma-separated list of services in the format:

  ```
  serviceName:port=domain.com,anotherService:port
  ```
  If the domain is omitted, sslip.io will be used with the autodiscovered public IP.

### Caddy Configuration

After the first run, you can customize the Caddy configuration by modifying the files in the `caddy_config` directory. This allows you to add custom Caddy directives or modify existing ones as needed.

Take into account that if you want to add or modify sites, you will need to remove the existing Caddy configuration for those sites in order to let AutoHttps generate the Caddyfile again.
