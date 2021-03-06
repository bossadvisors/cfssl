package bundle

import (
	"fmt"

	"github.com/cloudflare/cfssl/bundler"
	"github.com/cloudflare/cfssl/cli"
	"github.com/cloudflare/cfssl/ubiquity"
)

// Usage text of 'cfssl bundle'
var bundlerUsageText = `cfssl bundle -- create a certificate bundle that contains the client cert

Usage of bundle:
	- Bundle local certificate files
        cfssl bundle [-ca-bundle file] [-int-bundle file] [-key keyfile] [-flavor int] [-metadata file] CERT
	- Bundle certificate from remote server.
        cfssl bundle -domain domain_name [-ip ip_address] [-ca-bundle file] [-int-bundle file] [-metadata file]

Arguments:
	CERT:    Client certificate, possible followed by intermediates to form a (partial) chain, use '-' to read from stdin.

Note:
	CERT can be specified as flag value. But flag value will take precedence, overwriting the argument.

Flags:
`

// flags used by 'cfssl bundle'
var bundlerFlags = []string{"cert", "key", "ca-bundle", "int-bundle", "flavor", "metadata", "domain", "ip", "config"}

// bundlerMain is the main CLI of bundler functionality.
// TODO(zi): Decide whether to drop the argument list and only use flags to specify all the inputs.
// There are debates on whether flag or arg is more appropriate for required parameters.
func bundlerMain(args []string, c cli.Config) (err error) {
	// Grab cert file through args only if flag values for cert and domain are absent
	if c.CertFile == "" && c.Domain == "" {
		c.CertFile, args, err = cli.PopFirstArgument(args)
		if err != nil {
			return
		}
	}

	ubiquity.LoadPlatforms(c.Metadata)
	flavor := bundler.BundleFlavor(c.Flavor)
	// Initialize a bundler with CA bundle and intermediate bundle.
	b, err := bundler.NewBundler(c.CABundleFile, c.IntBundleFile)
	if err != nil {
		return
	}

	var bundle *bundler.Bundle
	if c.CertFile != "" {
		if c.CertFile == "-" {
			var certPEM, keyPEM []byte
			certPEM, err = cli.ReadStdin(c.CertFile)
			if err != nil {
				return
			}
			if c.KeyFile != "" {
				keyPEM, err = cli.ReadStdin(c.KeyFile)
				if err != nil {
					return
				}
			}
			bundle, err = b.BundleFromPEM(certPEM, keyPEM, flavor)
			if err != nil {
				return
			}
		} else {
			// Bundle the client cert
			bundle, err = b.BundleFromFile(c.CertFile, c.KeyFile, flavor)
			if err != nil {
				return
			}
		}
	} else if c.Domain != "" {
		bundle, err = b.BundleFromRemote(c.Domain, c.IP, flavor)
		if err != nil {
			return
		}
	}
	marshaled, err := bundle.MarshalJSON()
	if err != nil {
		return
	}
	fmt.Printf("%s", marshaled)
	return
}

// Command assembles the definition of Command 'bundle'
var Command = &cli.Command{UsageText: bundlerUsageText, Flags: bundlerFlags, Main: bundlerMain}
