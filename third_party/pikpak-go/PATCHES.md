# pikpak-go local patch

This directory vendors `github.com/kanghengliu/pikpak-go v0.1.0` through a `go.mod` replace directive.

Local patches:

- Switch login from the old web `/v1/auth/token` password grant to the current `/v1/auth/signin` flow used by `PikPakAPI==0.1.11`.
- Use Android client constants: `client_id=YNxT9w7GMdWvEOKa`, package `com.pikcloud.pikpak`, version `1.47.1`, and updated captcha salts.
- Add captcha initialization before signin and include `captcha_token` in the signin request.
- Keep `RequestLogin.ClientSecret` correctly tagged as `json:"client_secret,omitempty"` for signin.

The old web auth flow can fail with:

`permission_denied: [Danger], Please Do Not Save Your client_secret in browser, it is NOT SAFE`
