# Legacy

Collection of legacy plugins for backward compatibility.

Decoder now doesn't have hard dependency of *vector packages, so need to provide support of old callbacks/modifiers for
parsing various formats.

### Usage

Just add package to import section like that
```go
import (
	_ "github.com/koykov/decoder/legacy"
)

```
