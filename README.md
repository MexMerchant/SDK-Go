Disclaimer: Please note that we no longer support older versions of SDKs and Modules. We recommend that the latest versions are used.

# README

# Contents
- Introduction
- Prerequisites
- Using the Gateway SDK
- License

# Introduction
This Go SDK provides an easy method to integrate with the payment gateway.
 - The gateway/gateway.go file contains the main body of the SDK.
 - The sample-webserver.go file is intended as a minimal guide to demonstrate a complete 3DSv2 authentication process.

# Prerequisites
- The SDK requires the following prerequisites to be met in order to function correctly:
    - Go v1.10+

> <span style="color: red">Please note that we can only offer support for the SDK itself. While every effort has been made to ensure the sample code is complete and bug free, it is only a guide and should not be used in a production environment.</span>

# Using the Gateway SDK

Require the gateway SDK into your project

```
import (
	"./gateway"
)
```

Instantiate the Gateway object

```
g := gateway.NewGateway("https://test.3ds-pit.com/direct/", "100856", "Threeds2Test60System")
```

Once your SDK has been required. You create your request array, for example:

```
	url.Values{
		"merchantID":        {"100856"},
		"action":            {"SALE"},
		"type":              {"1"},
		"transactionUnique": {randStringBytes(10)},
		"currencyCode":      {"826"},
		"countryCode":       {"826"},
		"amount":            {"1001"},
		"cardNumber":        {"4012001037141112"},
		"cardExpiryMonth":   {"12"},
		"cardExpiryYear":    {"20"},
		"cardCVV":           {"356"},
		"customerName":      {"Test Customer"},
		"customerEmail":     {"test@testcustomer.com"},
		"customerAddress":   {"16 Test Street"},
		"customerPostCode":  {"XX15 5XX"},
		"orderRef":          {"Test purchase"},

		// The following fields are mandatory for direct 3DS v2
		"threeDSVersion":       {"2"},
		"threeDSRedirectURL":   {pageURL + "acs=1"}, // Go's understanding of URIs seems to always include a ?
	}

```
> NB: This is a sample request. The gateway features many more options. Please see our integration guides for more details.

Then, depending on your integration method, you'd either call (as a promise):

```
response, err := g.DirectRequest(reqFields)
```

OR

```
response, err := g.HostedRequest(reqFields)
```

And then handle the response received from the gateway.

License
----
MIT
