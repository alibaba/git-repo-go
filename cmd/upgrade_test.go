package cmd

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVerify(t *testing.T) {
	var (
		data     = "Hello, world.\n"
		checksum = "1ab1a2bb8502820a83881a5b66910b819121bafe336d76374637aa4ea7ba2616  hello.txt\n"
		sig      = `
-----BEGIN PGP SIGNATURE-----

iQJOBAABCgA4FiEEh/2pehIUeddqpt2hs/sBsKSIL5IFAlz3FUQaHHpoaXlvdS5q
eEBhbGliYWJhLWluYy5jb20ACgkQs/sBsKSIL5L6MRAAn8ncLGis6tAvOMwGS4R6
5vfCxtoWb8d7Fv9At6ANzo1M0vAeAfTfbXFqYV6oqzS89In1E9Km4Wi/i9Qr5yu8
u2L0gaDrK2n04K+NSGG61DcSTb2vyldpY5+I3Grbvw56G8WpnIKLmCwlF1GrhJWE
89G9p/WPATnkgaCtSbIaMSW0tW3PTl31XTUFqdQ2lutn+BDxY6ItjU5znESyCTC+
+3zFlIqqOo0iM4BCFLTwZk375EyxPj5WUJWWdbTMo7Bh5oJgNC7saaLEvCUQ9mi9
0ffQy3Zl1lykudxYlHJrNL+ym5VaB3l6rMe94v3cD6ObyPyy1jWh9r1QPOnQxCTq
lU3Y/MGynJfO1GBKE/bM8hY/t/Rt8unP9F/A7gMkvNDCjjVZ5qv6eKgvzkEMusjJ
XhRtsynEo4ybVzeHgBXopKqSR6A7kRemcq6DR7v/5iEoPzjIQgpd/P9VsYV+jTvU
V2jHrRxjRo1L7HGmpkwg1cdU0ewNjlrf+qtRFUPGGAW9ObSmsww6ggGCn/GXVZnG
43f0Tuo0BGrRR8s37P1TV0N0kCFPCqH/AxWA9UUr4cjTSiaOC2M1vNDAYyLWNDx9
qUDzwGflENTCedt+7lz7oLhAk3Oor0Gxk95u43ki9REJMIkS68CXDfe3t4GKUrKa
9GO21W1p3viN2eJhj76/oZU=
=K8cW
-----END PGP SIGNATURE-----
`

		err    error
		assert = assert.New(t)
	)

	binFile, err := ioutil.TempFile("", "test-git-repo.-")
	if err != nil {
		t.Fatalf("fail to create binfile: %s", err)
	}
	defer os.Remove(binFile.Name())
	binFile.WriteString(data)
	binFile.Close()

	checksumFile, err := os.OpenFile(binFile.Name()+".sha256", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		t.Fatalf("fail to create checksum file: %s", err)
	}
	defer os.Remove(checksumFile.Name())
	checksumFile.WriteString(checksum)
	checksumFile.Close()

	sigFile, err := os.OpenFile(binFile.Name()+".sha256.gpg", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		t.Fatalf("fail to create sig file: %s", err)
	}
	defer os.Remove(sigFile.Name())
	sigFile.WriteString(sig)
	sigFile.Close()

	cmd := upgradeCommand{}
	err = cmd.verifyChecksum(binFile.Name(), checksumFile.Name())
	assert.Nil(err)

	err = cmd.verifySignature(checksumFile.Name(), sigFile.Name())
	assert.Nil(err)
}
