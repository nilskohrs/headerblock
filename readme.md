# Header Block

Header Block is a middleware plugin for [Traefik](https://github.com/traefik/traefik) to block request and response headers which regex matched by their name and/or value

## Configuration

### Static

```yaml
pilot:
  token: "xxxxx"

experimental:
  plugins:
    headerblock:
      moduleName: "github.com/nilskohrs/headerblock"
      version: "v0.0.1"
```

### Dynamic

```yaml
http:
  middlewares:
    headerblock-foo:
      headerblock:
        requestHeaders:
          - name: "header"
            value: "value"
        responseHeaders:
          - name: "header"
            value: "value"
```