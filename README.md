# passlib
passlib is a Go-based harm reduction program for credentials stored in the
local [development] environment.

It uses the `pass` utility and the address scheme for that application.

It requires GUI passphrase management software such as GNOME Keyring for GPG
decryption.

## Usage
```go
import "github.com/richlj/passlib"

password, err := pass.Get("^http/.*some.*/.*$")

/* Returns the *pass.Item corresponding to the path "http/someAPI/username"
together with a nil error value, if this is the only match */
```
