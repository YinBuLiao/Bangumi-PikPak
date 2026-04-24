# pikpak-go local patch

This directory vendors `github.com/kanghengliu/pikpak-go v0.1.0` through a `go.mod` replace directive.

Local patch:

- Fix `RequestLogin.ClientSecret` JSON tag from malformed `json:client_secret,omitempty` to `json:"client_secret,omitempty"`.

Without this patch, login JSON is encoded as `ClientSecret`, which can make PikPak reject auth with `invalid_argument` / `currently not supported`.
