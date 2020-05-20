vcl 4.0;

import directors;

{{range $key, $value := . -}}
{{if $value -}}
backend backend_{{$value.ShortID}} {
    .host = "{{$value.GetIPAddressFromNetwork}}";
    .port = "80";
}

{{end -}}
{{end -}}

sub vcl_init {
    new vdir = directors.round_robin();
    {{- range $key, $value := .}}
    {{- if $value }}
    vdir.add_backend(backend_{{$value.ShortID}});
    {{- end}}
    {{- end}}
}

sub vcl_recv {
  set req.backend_hint = vdir.backend();
}

sub vcl_deliver {
  unset resp.http.Server;
  unset resp.http.Via;
}
