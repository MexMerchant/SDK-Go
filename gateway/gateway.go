package gateway

import (
	"crypto/sha512"
	"errors"
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

// Gateway - container for the key information
type Gateway struct {
	DirectURL      string
	MerchantID     string
	MerchantSecret string
}

// NewGateway - simple constructor
func NewGateway(directURL, merchantID, merchantSecret string) *Gateway {
	g := new(Gateway)
	g.DirectURL = directURL
	g.MerchantID = merchantID
	g.MerchantSecret = merchantSecret

	return g
}

// DirectRequest - Perform a direct request (as opposed to Hosted) to the gateway server.
func (g *Gateway) DirectRequest(fields url.Values) (url.Values, error) {
	fields.Add("signature", g.Sign(fields, g.MerchantSecret))

	resp, err := http.PostForm(g.DirectURL, fields)
	responsebody, _ := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	responseFields, err := url.ParseQuery(string(responsebody))

	if err != nil {
		return nil, err
	}

	if !g.VerifyResponse(responseFields, g.MerchantSecret) {
		return nil, errors.New("Response from server failed verification")
	}

	return responseFields, nil
}

func fieldsToHTML(name string, values map[string]string) string {

	var builder strings.Builder
	for n, v := range values {
		builder.WriteString(fieldToHTML(name+"["+n+"]", v))
	}

	return builder.String()
}

func fieldToHTML(name string, value string) string {

	// Convert all applicable characters or none printable characters to HTML entities
	return fmt.Sprintf(`<input type="hidden" name="%s" value="%s" />%s`, name, html.EscapeString(value), "\n")
}

// Exclusively used by the sign function below.
func getSortedKeys(m url.Values) []string {
	var keys []string = make([]string, 0, len(m))

	for k := range m {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		return strings.Replace(keys[i], "[", "00", 1) < strings.Replace(keys[j], "[", "00", 1)
	})

	return keys
}

// Returns a signature calculated from the fields using the merchant's key.
// This must be consistent with the server's implementation
func (g *Gateway) Sign(fields url.Values, key string) string {
	// Get an array of the keys sorted with [ preceding all alphanumeric characters
	sortedKeys := getSortedKeys(fields)

	var encodedFields []string = make([]string, 0, len(sortedKeys))
	// There is a url.Value.encode method, however this will apply its own sort
	// to the result which unfortunately invalidates the key
	for _, k := range sortedKeys {
		encodedFields = append(encodedFields, url.QueryEscape(k)+"="+url.QueryEscape(fields[k][0]))
	}
	hashbody := strings.Join(encodedFields, "&")
	hashbody = strings.ReplaceAll(hashbody, "*", "%2A")

	hashThis := []byte(hashbody + key)

	hashResult := sha512.Sum512(hashThis)
	return fmt.Sprintf("%x", hashResult)
}

func (g *Gateway) VerifyResponse(responseFields url.Values, key string) bool {
	signature := responseFields["signature"][0]
	responseFields.Del("signature")
	ourSignature := g.Sign(responseFields, key)

	return ourSignature == signature
}
