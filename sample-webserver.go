package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"./gateway"
)

// NOTICE, THIS MUST BE CHANGED TO AN APPROPRIATE SESSION VARIABLE
var threeDSRef string

func handler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	mainDispatch(w, r)
}

// This is the primary entry point for 3DSv2 transaction flows.
func mainDispatch(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		panic(err)
	}

	var body string
	var g = gateway.NewGateway("https://example.com/direct/", "100856", "Threeds2Test60System")

	// Just to make life a little easier
	for k := range r.PostForm {
		if index := strings.Index(k, "["); index != -1 {
			r.PostForm.Add(k[:index], "Array")
		}
	}

	if r.Form.Get("acs") != "" {
		// This is the post back from the 3DS server, the following posts it back
		// again using a target of _parent to remove the frame

		fields := map[string]string{}
		for k, v := range r.PostForm {
			fields["threeDSResponse["+k+"]"] = v[0]
		}
		body = silentPost(getPageURL(r), fields, "_parent")

	} else {
		if r.PostForm.Get("browserInfo[deviceChannel]") == "" && r.PostForm.Get("threeDSResponse") == "" {
			// The first thing we need to do is capture browser information,
			// which will then get further passed onto the Gateway
			body = collectBrowserInfo(r)

		} else if r.PostForm.Get("threeDSResponse") != "" {
			// Response from the 3DS Server, so the browser is returning from
			// the 3DS server (via the acs juggling to remove the frame)

			reqFields := url.Values{
				"threeDSRef": {threeDSRef},
			}
			for k, v := range r.PostForm {
				if strings.HasPrefix(k, "threeDSResponse[") {
					reqFields[k] = v
				}
			}

			// Send the 3DS response to the gateway (which may either then
			// approve the transaction, or require further verification)
			response, err := g.DirectRequest(reqFields)
			if err != nil {
				panic(err)
			}
			body = processResponse(response)

		} else {
			// Browser info present, but no threeDSResponse, this means it's the
			// initial request to the gateway (not 3DS) server. // RemoteAddr should be ipv4:port
			// Please note that IPv6 format addresses aren't supported
			reqFields := getInitialRequestFields(getPageURL(r), strings.Split(r.RemoteAddr, ":")[0])

			for k, v := range r.PostForm {
				if strings.HasPrefix(k, "browserInfo[") {
					reqFields[k[12:len(k)-1]] = v
				}
			}

			// Send the initial request to the gateway. This must contain the
			// basic transaction fields and the browser info.
			response, err := g.DirectRequest(reqFields)
			if err != nil {
				panic(err)
			}
			body = processResponse(response)
		}
	}

	w.Write(getWrapHTML(body))
}

// NOTICE, this is likely to depend on the deployment configuration.
// This was developed using Apache as a reverse proxy, since:
// NOTICE, the gateway requires HTTPS, it will reject HTTP.
func getPageURL(r *http.Request) string {
	if r.Header.Get("X-Forwarded-Server") != "" {
		return "https://" + r.Header["X-Forwarded-Server"][0] + strings.Replace(r.RequestURI, "acs=1", "", 1)
	}

	var protocol string
	if r.TLS != nil {
		protocol = "https://"
	} else {
		protocol = "http://"
	}

	return protocol + r.Host + strings.Replace(r.RequestURI, "acs=1", "", 1)
}

func silentPost(url string, fields map[string]string, target string) string {
	var builder strings.Builder

	for n, v := range fields {
		builder.WriteString(fmt.Sprintf(`<input type="hidden" name="%s" value="%s" />`+"\n", n, v))
	}

	formtag := fmt.Sprintf(`<form id="silentPost" action="%s" method="post" target="%s"> %s`, url, target, "\n")

	return formtag + builder.String() + `
	<noscript><input type="submit" value="Continue"></noscript>
	</form>
	<script>
		 window.setTimeout('document.forms.silentPost.submit()', 0);
	</script>
	`
}

// Determines the next step following the response from the server.
// Either contact the 3DS server, or the payment has been accepted or rejected.
func processResponse(responseFields url.Values) string {
	if responseFields["responseCode"][0] == "65802" {
		return showFrameForThreeDS(responseFields)

	} else if responseFields["responseCode"][0] == "0" {
		return "<p>Thank you for your payment.</p>"
	}

	return "<p>Failed to take payment: " + responseFields["responseMessage"][0] + "</p>"
}

func collectBrowserInfo(r *http.Request) string {
	return fmt.Sprintf(`
	<form id="collectBrowserInfo" method="post" action="?">
	<input type="hidden" name="browserInfo[deviceChannel]" value="browser" />
	<input type="hidden" name="browserInfo[deviceIdentity]" value="%s" />
	<input type="hidden" name="browserInfo[deviceTimeZone]" value="0" />
	<input type="hidden" name="browserInfo[deviceCapabilities]" value="" />
	<input type="hidden" name="browserInfo[deviceScreenResolution]" value="1x1x1" />
	<input type="hidden" name="browserInfo[deviceAcceptContent]" value="%s" />
	<input type="hidden" name="browserInfo[deviceAcceptEncoding]" value="%s" />
	<input type="hidden" name="browserInfo[deviceAcceptLanguage]" value="%s" />

	</form>
	<script>
	  var screen_width = (window && window.screen ? window.screen.width : '0');
	  var screen_height = (window && window.screen ? window.screen.height : '0');
	  var screen_depth = (window && window.screen ? window.screen.colorDepth : '0');
	  var identity = (window && window.navigator ? window.navigator.userAgent : '');
	  var language = (window && window.navigator ? (window.navigator.language ? window.navigator.language : window.navigator.browserLanguage) : '');
	  var timezone = (new Date()).getTimezoneOffset();
	  var java = (window && window.navigator ? navigator.javaEnabled() : false);
	  var fields = document.forms.collectBrowserInfo.elements;
	  fields['browserInfo[deviceIdentity]'].value = identity;
	  fields['browserInfo[deviceTimeZone]'].value = timezone;
	  fields['browserInfo[deviceCapabilities]'].value = 'javascript' + (java ? ',java' : '');
	  fields['browserInfo[deviceAcceptLanguage]'].value = language;
	  fields['browserInfo[deviceScreenResolution]'].value = screen_width + 'x' + screen_height + 'x' + screen_depth;
	  window.setTimeout('document.forms.collectBrowserInfo.submit()', 0);
	</script>

	`, string(r.Header.Get("User-Agent")[0]),
		string(r.Header.Get("Accept")[0]),
		string(r.Header.Get("Accept-Encoding")[0]),
		string(r.Header.Get("Accept-Language")[0]))
}

func getWrapHTML(content string) []byte {
	return []byte(strings.TrimSpace(`
<!DOCTYPE html>
<html>
  <head>
    <meta charset="UTF-8" />
  </head>
  <body>`) + "\n\n" + content +
		`  </body>
</html>`)
}

func showFrameForThreeDS(responseFields url.Values) string {
	//Send a request to the ACS server by POSTing a form with the target set as the IFrame.

	//The form is hidden for threeDSMethodData requests (frictionless) and visible when the ACS
	//server may show a challenge to the user.

	var style string
	if responseFields.Get("threeDSRequest[threeDSMethodData]") != "" {
		style = "display: none;"
	} else {
		style = ""
	}

	iframeHTML := fmt.Sprintf(`<iframe name="threeds_acs" style="height:420px; width:420px; %s"></iframe>`+"\n\n", style)

	// We could extract each key by name, however in the interests of
	// facilitating forwards compatibility, we pass through every field in the
	//  threeDSRequest array.
	formField := map[string]string{}

	for k, v := range responseFields {
		if strings.HasPrefix(k, "threeDSRequest[") && strings.HasSuffix(k, "]") {
			formKey := k[15 : len(k)-1]
			formField[formKey] = v[0]
		}
	}

	// Silently POST the 3DS request to the ACS in the IFRAME
	silentPostHTML := silentPost(responseFields["threeDSURL"][0], formField, "threeds_acs")

	// Remember the threeDSRef, it's required when the ACS server responds.
	threeDSRef = responseFields["threeDSRef"][0]

	return iframeHTML + silentPostHTML
}

func getInitialRequestFields(pageURL string, remoteAddress string) url.Values {
	return url.Values{
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

		// The following fields are mandatory for 3DS v2
		"remoteAddress":        {remoteAddress},
		"merchantCategoryCode": {"5411"},
		"threeDSVersion":       {"2"},
		"threeDSRedirectURL":   {pageURL + "acs=1"}, // Go's understanding of URIs seems to always include a ?
	}
}

// Hat tip; https://stackoverflow.com/q/22892120
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randStringBytes(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func main() {
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8012", nil))
}
