// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"testing"
)

func TestMainHelp(t *testing.T) {
	origArgs := os.Args
	t.Cleanup(func() {
		os.Args = origArgs
	})

	os.Args = []string{"gemara-tools", "--help"}
	main()
}
