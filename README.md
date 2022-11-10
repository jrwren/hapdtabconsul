# HAPDTABCONSUL
**Proxy DTAB CONSUL

Yes, the name is not creative.

Goal: replace linkerd+namerd with haproxy configuration built from dtab and consul data.

Actual: not tied to haproxy at all, but the examples generate haproxy config.
This could work with any proxy.

This program is a plugin to consul-template so that consul-template is still used to render the haproxy configuration. See https://github.com/hashicorp/consul-template/blob/main/docs/plugins.md

This program expects a dtab to be passed as a string as the first argument and a json encoded list of consul services to be passed as the second argument.

The output of this program is a JSON array ready to be consumed by consul-template.

Very few namerd+dtab finagle expressions are supported;
Only enough for our environment.

```
{{ $dtab := key "namerd/dtabs/default" }}
{{ $h := services | toJSON | plugin "hapdtabconsul" ($dtab) | parseJSON}}
```

## haproxy

A more real example to generate a partial haproxy config:

```
{{ $dtab := key "namerd/dtabs/default" }}
{{ $c := (plugin "hapdtabconsul" ($dtab) (services | toJSON) "https"|parseJSON) }}
{{ range $c.canary_services }}
    use_backend public-{{.name}} if { hdr_beg(host) -i {{.name}} }
{{ end }}
# Backends
{{ range $c.canary_services }}
backend public-{{.name}}
    option forwardfor
    http-request set-header X-Forwarded-Port %[dst_port]
    http-request add-header X-Forwarded-Proto https if { ssl_fc }
    option httpchk GET /
    http-check expect status 401
    retry-on conn-failure
    retries 1
    timeout connect 4000
    {{ $canary_weight := .canary_weight -}}
    {{- $non_canary_weight := .non_canary_weight -}}
    {{- $name := .name -}}
    {{- range $i, $s := service .name -}}
      {{- if .Tags | contains "canary" }}
    server {{.name}}{{$i}} {{$s.Address}}:{{$s.Port}} weight {{$canary_weight}}  # canary
      {{ end -}}
      {{- if .Tags | contains "noncanary" }}
    server {{$name}}{{$i}} {{$s.Address}}:{{$s.Port}} weight {{$non_canary_weight}}  # noncanary
      {{ end -}}
    {{- end -}}
{{ end }}

```