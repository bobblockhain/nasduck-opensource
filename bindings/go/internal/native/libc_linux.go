// Derived from tui-test (https://github.com/microsoft/tui-test), MIT, Microsoft Corporation.
package native

import "github.com/ebitengine/purego"

func nativeLibc() string {
	// Inspect the running process, not the build host or the distribution name.
	// Of the two supported libcs, only glibc exports this symbol.
	if _, err := purego.Dlsym(purego.RTLD_DEFAULT, "gnu_get_libc_version"); err == nil {
		return "gnu"
	}
	return "musl"
}
