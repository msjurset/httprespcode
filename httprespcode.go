
package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// ANSI color constants
const (
	reset   = "\033[0m"
	bold    = "\033[1m"
	dim     = "\033[2m"
	red     = "\033[31m"
	green   = "\033[32m"
	yellow  = "\033[33m"
	blue    = "\033[34m"
	magenta = "\033[35m"
	cyan    = "\033[36m"
)

// statusColor returns the ANSI color for a given HTTP status code class.
func statusColor(code int) string {
	switch {
	case code >= 100 && code < 200:
		return cyan
	case code >= 200 && code < 300:
		return green
	case code >= 300 && code < 400:
		return yellow
	case code >= 400 && code < 500:
		return red
	case code >= 500 && code < 600:
		return magenta
	default:
		return ""
	}
}

var verbose bool

func main() {
	flag.BoolVar(&verbose, "v", false, "show extended details for status codes")
	flag.Parse()

	args := flag.Args()
	if len(args) > 0 {
		printStatus(args[0], verbose)
	} else {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Print(bold + "Enter HTTP status code (or 'q' to quit): " + reset)
		for scanner.Scan() {
			codeStr := scanner.Text()
			if codeStr == "q" {
				break
			}
			printStatus(codeStr, verbose)
			fmt.Print(bold + "Enter HTTP status code (or 'q' to quit): " + reset)
		}
	}
}

var statusDescriptions = map[int]string{
	100: "Continue: The server has received the request headers and the client should proceed to send the request body (in the case of a request for which a body needs to be sent; for example, a POST request). Sending a large request body to a server after a request has been rejected for inappropriate headers would be inefficient. To have a server check the request's headers, a client must send Expect: 100-continue as a header in its initial request and receive a 100 Continue status code in response before sending the body. If the client receives an error code such as 403 (Forbidden) or 405 (Method Not Allowed) then it should not send the request's body. The response 417 Expectation Failed indicates that the request should be repeated without the Expect header as it indicates that the server does not support expectations (this is the case, for example, of HTTP/1.0 servers).",
	101: "Switching Protocols: The requester has asked the server to switch protocols and the server has agreed to do so.",
	102: "Processing (WebDAV; RFC 2518): A WebDAV request may contain many sub-requests involving file operations, requiring a long time to complete the request. This code indicates that the server has received and is processing the request, but no response is available yet. This prevents the client from timing out and assuming the request was lost. The status code is deprecated.",
	103: "Early Hints (RFC 8297): Used to return some response headers before final HTTP message.",
	200: "OK: Standard response for successful HTTP requests. The actual response will depend on the request method used. In a GET request, the response will contain an entity corresponding to the requested resource. In a POST request, the response will contain an entity describing or containing the result of the action.",
	201: "Created: The request has been fulfilled, resulting in the creation of a new resource.",
	202: "Accepted: The request has been accepted for processing, but the processing has not been completed. The request might or might not be eventually acted upon, and may be disallowed when processing occurs.",
	203: "Non-Authoritative Information (since HTTP/1.1): The server is a transforming proxy (e.g. a Web accelerator) that received a 200 OK from its origin, but is returning a modified version of the origin's response.",
	204: "No Content: The server successfully processed the request, and is not returning any content.",
	205: "Reset Content: The server successfully processed the request, asks that the requester reset its document view, and is not returning any content.",
	206: "Partial Content: The server is delivering only part of the resource (byte serving) due to a range header sent by the client. The range header is used by HTTP clients to enable resuming of interrupted downloads, or split a download into multiple simultaneous streams.",
	207: "Multi-Status (WebDAV; RFC 4918): The message body that follows is by default an XML message and can contain a number of separate response codes, depending on how many sub-requests were made.",
	208: "Already Reported (WebDAV; RFC 5842): The members of a DAV binding have already been enumerated in a preceding part of the (multistatus) response, and are not being included again.",
	226: "IM Used (RFC 3229): The server has fulfilled a request for the resource, and the response is a representation of the result of one or more instance-manipulations applied to the current instance.",
	300: "Multiple Choices: Indicates multiple options for the resource from which the client may choose (via agent-driven content negotiation). For example, this code could be used to present multiple video format options, to list files with different filename extensions, or to suggest word-sense disambiguation.",
	301: "Moved Permanently: This and all future requests should be directed to the given URI.",
	302: "Found (Previously \"Moved temporarily\"): Tells the client to look at (browse to) another URL. The HTTP/1.0 specification required the client to perform a temporary redirect with the same method (the original describing phrase was \"Moved Temporarily\"), but popular browsers implemented 302 redirects by changing the method to GET. Therefore, HTTP/1.1 added status codes 303 and 307 to distinguish between the two behaviours.",
	303: "See Other (since HTTP/1.1): The response to the request can be found under another URI using the GET method. When received in response to a POST (or PUT/DELETE), the client should presume that the server has received the data and should issue a new GET request to the given URI.",
	304: "Not Modified: Indicates that the resource has not been modified since the version specified by the request headers If-Modified-Since or If-None-Match. In such case, there is no need to retransmit the resource since the client still has a previously-downloaded copy.",
	305: "Use Proxy (since HTTP/1.1): The requested resource is available only through a proxy, the address for which is provided in the response. For security reasons, many HTTP clients (such as Mozilla Firefox and Internet Explorer) do not obey this status code.",
	306: "Switch Proxy: No longer used. Originally meant \"Subsequent requests should use the specified proxy.\"",
	307: "Temporary Redirect (since HTTP/1.1): In this case, the request should be repeated with another URI; however, future requests should still use the original URI. In contrast to how 302 was historically implemented, the request method is not allowed to be changed when reissuing the original request. For example, a POST request should be repeated using another POST request.",
	308: "Permanent Redirect: This and all future requests should be directed to the given URI. 308 parallels the behavior of 301, but does not allow the HTTP method to change. So, for example, submitting a form to a permanently redirected resource may continue smoothly.",
	400: "Bad Request: The server cannot or will not process the request due to an apparent client error (e.g., malformed request syntax, size too large, invalid request message framing, or deceptive request routing).",
	401: "Unauthorized: Similar to 403 Forbidden, but specifically for use when authentication is required and has failed or has not yet been provided. The response must include a WWW-Authenticate header field containing a challenge applicable to the requested resource. See Basic access authentication and Digest access authentication. 401 semantically means \"unauthenticated\", the user does not have valid authentication credentials for the target resource.",
	402: "Payment Required: Reserved for future use. The original intention was that this code might be used as part of some form of digital cash or micropayment scheme, as proposed, for example, by GNU Taler, but that has not yet happened, and this code is not widely used. Google Developers API uses this status if a particular developer has exceeded the daily limit on requests. Sipgate uses this code if an account does not have sufficient funds to start a call. Shopify uses this code when the store has not paid their fees and is temporarily disabled. Stripe uses this code for failed payments where parameters were correct, for example blocked fraudulent payments.",
	403: "Forbidden: The request contained valid data and was understood by the server, but the server is refusing action. This may be due to the user not having the necessary permissions for a resource or needing an account of some sort, or attempting a prohibited action (e.g. creating a duplicate record where only one is allowed). This code is also typically used if the request provided authentication by answering the WWW-Authenticate header field challenge, but the server did not accept that authentication. The request should not be repeated.",
	404: "Not Found: The requested resource could not be found but may be available in the future. Subsequent requests by the client are permissible.",
	405: "Method Not Allowed: A request method is not supported for the requested resource; for example, a GET request on a form that requires data to be presented via POST, or a PUT request on a read-only resource.",
	406: "Not Acceptable: The requested resource is capable of generating only content not acceptable according to the Accept headers sent in the request. See Content negotiation.",
	407: "Proxy Authentication Required: The client must first authenticate itself with the proxy.",
	408: "Request Timeout: The server timed out waiting for the request. According to HTTP specifications: \"The client did not produce a request within the time that the server was prepared to wait. The client MAY repeat the request without modifications at any later time.\"",
	409: "Conflict: Indicates that the request could not be processed because of conflict in the current state of the resource, such as an edit conflict between multiple simultaneous updates.",
	410: "Gone: Indicates that the resource requested was previously in use but is no longer available and will not be available again. This should be used when a resource has been intentionally removed and the resource should be purged. Upon receiving a 410 status code, the client should not request the resource in the future. Clients such as search engines should remove the resource from their indices. Most use cases do not require clients and search engines to purge the resource, and a \"404 Not Found\" may be used instead.",
	411: "Length Required: The request did not specify the length of its content, which is required by the requested resource.",
	412: "Precondition Failed: The server does not meet one of the preconditions that the requester put on the request header fields.",
	413: "Payload Too Large: The request is larger than the server is willing or able to process. Previously called \"Request Entity Too Large\".",
	414: "URI Too Long: The URI provided was too long for the server to process. Often the result of too much data being encoded as a query-string of a GET request, in which case it should be converted to a POST request. Called \"Request-URI Too Long\" previously.",
	415: "Unsupported Media Type: The request entity has a media type which the server or resource does not support. For example, the client uploads an image as image/svg+xml, but the server requires that images use a different format.",
	416: "Range Not Satisfiable: The client has asked for a portion of the file (byte serving), but the server cannot supply that portion. For example, if the client asked for a part of the file that lies beyond the end of the file. Called \"Requested Range Not Satisfiable\" previously.",
	417: "Expectation Failed: The server cannot meet the requirements of the Expect request-header field.",
	418: "I'm a teapot (RFC 2324, RFC 7168): This code was defined in 1998 as one of the traditional IETF April Fools' jokes, in RFC 2324, Hyper Text Coffee Pot Control Protocol, and is not expected to be implemented by actual HTTP servers. The RFC specifies this code should be returned by teapots requested to brew coffee. This HTTP status is used as an Easter egg in some websites, such as Google.com's \"I'm a teapot\" easter egg. Sometimes, this status code is also used as a response to a blocked request, instead of the more appropriate 403 Forbidden.",
	421: "Misdirected Request: The request was directed at a server that is not able to produce a response (for example because of connection reuse).",
	422: "Unprocessable Content: The request was well-formed (i.e., syntactically correct) but could not be processed.",
	423: "Locked (WebDAV; RFC 4918): The resource that is being accessed is locked.",
	424: "Failed Dependency (WebDAV; RFC 4918): The request failed because it depended on another request and that request failed (e.g., a PROPPATCH).",
	425: "Too Early (RFC 8470): Indicates that the server is unwilling to risk processing a request that might be replayed.",
	426: "Upgrade Required: The client should switch to a different protocol such as TLS/1.3, given in the Upgrade header field.",
	428: "Precondition Required (RFC 6585): The origin server requires the request to be conditional. Intended to prevent the 'lost update' problem, where a client GETs a resource's state, modifies it, and PUTs it back to the server, when meanwhile a third party has modified the state on the server, leading to a conflict.",
	429: "Too Many Requests (RFC 6585): The user has sent too many requests in a given amount of time. Intended for use with rate-limiting schemes.",
	431: "Request Header Fields Too Large (RFC 6585): The server is unwilling to process the request because either an individual header field, or all the header fields collectively, are too large.",
	451: "Unavailable For Legal Reasons (RFC 7725): A server operator has received a legal demand to deny access to a resource or to a set of resources that includes the requested resource. The code 451 was chosen as a reference to the novel Fahrenheit 451.",
	500: "Internal Server Error: A generic error message, given when an unexpected condition was encountered and no more specific message is suitable.",
	501: "Not Implemented: The server either does not recognize the request method, or it lacks the ability to fulfil the request. Usually this implies future availability (e.g., a new feature of a web-service API).",
	502: "Bad Gateway: The server was acting as a gateway or proxy and received an invalid response from the upstream server.",
	503: "Service Unavailable: The server cannot handle the request (because it is overloaded or down for maintenance). Generally, this is a temporary state.",
	504: "Gateway Timeout: The server was acting as a gateway or proxy and did not receive a timely response from the upstream server.",
	505: "HTTP Version Not Supported: The server does not support the HTTP version used in the request.",
	506: "Variant Also Negotiates (RFC 2295): Transparent content negotiation for the request results in a circular reference.",
	507: "Insufficient Storage (WebDAV; RFC 4918): The server is unable to store the representation needed to complete the request.",
	508: "Loop Detected (WebDAV; RFC 5842): The server detected an infinite loop while processing the request (sent instead of 208 Already Reported).",
	510: "Not Extended (RFC 2774): Further extensions to the request are required for the server to fulfil it.",
	511: "Network Authentication Required (RFC 6585): The client needs to authenticate to gain network access. Intended for use by intercepting proxies used to control access to the network (e.g., \"captive portals\" used to require agreement to Terms of Service before granting full Internet access via a Wi-Fi hotspot).",
}

var verboseDescriptions = map[int]string{
	100: `Common causes: The client sent an "Expect: 100-continue" header with a large POST/PUT request. The server responds with 100 to signal the client to proceed with the body.
Real-world usage: Large file uploads where the client wants to confirm the server will accept the request before transmitting the body. Commonly seen with HTTP/1.1 clients like curl when uploading files.
Related codes: 417 Expectation Failed (server does not support Expect header), 200 OK (final success after body is sent).
RFC: RFC 7231, Section 6.2.1.`,

	101: `Common causes: The client sent an Upgrade header requesting a protocol change (e.g., from HTTP/1.1 to WebSocket or HTTP/2).
Real-world usage: WebSocket connections begin with an HTTP request that includes "Upgrade: websocket". The server responds with 101 to confirm the switch. Also used in HTTP/2 upgrade from HTTP/1.1 (though most HTTP/2 uses ALPN during TLS instead).
Related codes: 426 Upgrade Required (server insists the client must upgrade).
RFC: RFC 7231, Section 6.2.2; RFC 6455 (WebSocket).`,

	102: `Common causes: A WebDAV PROPFIND or COPY operation on a large directory tree that takes significant time.
Real-world usage: Rarely seen in modern systems. Was used to prevent client timeouts during long WebDAV operations. Deprecated in RFC 4918 — modern implementations use asynchronous patterns instead.
Related codes: 207 Multi-Status (the eventual response to a multi-resource WebDAV request).
RFC: RFC 2518 (deprecated).`,

	103: `Common causes: The server knows certain resources (CSS, JS) will be needed and sends Link headers early so the browser can start preloading before the final response is ready.
Real-world usage: Performance optimization for web pages. The server sends 103 with "Link: </style.css>; rel=preload" headers while still computing the main 200 response. Supported by modern browsers and CDNs like Cloudflare.
Related codes: 200 OK (the final response that follows the early hints).
RFC: RFC 8297.`,

	200: `Common causes: The request was successful. This is the most common HTTP response code.
Real-world usage: Every successful web page load, API call, or resource fetch typically returns 200. In REST APIs, used for successful GET, PUT, and PATCH requests. POST requests that return the created resource may also use 200 (though 201 is more precise).
Related codes: 201 Created (specifically for new resource creation), 204 No Content (success but no body), 304 Not Modified (cached version is current).
RFC: RFC 7231, Section 6.3.1.`,

	201: `Common causes: A POST request successfully created a new resource on the server.
Real-world usage: REST APIs return 201 after creating a new record (e.g., POST /users creates a user). The response typically includes a Location header pointing to the new resource and may include the created entity in the body.
Related codes: 200 OK (general success), 202 Accepted (creation queued but not yet complete), 409 Conflict (creation failed due to duplicate).
RFC: RFC 7231, Section 6.3.2.`,

	202: `Common causes: The server accepted the request for processing but hasn't completed it yet. Common with asynchronous operations.
Real-world usage: Job queues, batch processing, and long-running operations. For example, submitting a video for transcoding, initiating a report generation, or triggering a CI/CD pipeline. The response often includes a URL to poll for status.
Related codes: 200 OK (synchronous success), 201 Created (synchronous creation), 303 See Other (redirect to status endpoint).
RFC: RFC 7231, Section 6.3.3.`,

	203: `Common causes: A proxy or CDN modified the response from the origin server (e.g., added headers, transformed content).
Real-world usage: Rarely used explicitly. Could be returned by a transforming proxy that compresses images, strips headers, or modifies the response body. Most proxies simply return 200 even when they modify responses.
Related codes: 200 OK (unmodified response from origin), 214 Warning (non-standard, used for transformation warnings).
RFC: RFC 7231, Section 6.3.4.`,

	204: `Common causes: A successful request that intentionally returns no body. Common for DELETE operations and updates that don't return the modified resource.
Real-world usage: REST APIs use 204 for successful DELETE requests, PUT/PATCH when the client doesn't need the updated resource back, and for preflight CORS responses. Also used for "fire and forget" endpoints like analytics event collection.
Related codes: 200 OK (success with body), 205 Reset Content (success, client should reset form).
RFC: RFC 7231, Section 6.3.5.`,

	205: `Common causes: The server successfully processed the request and wants the client to reset the document view (e.g., clear a form).
Real-world usage: Rarely used in practice. Intended for browser forms where after a successful submission, the form should be cleared. Most web applications handle form reset client-side with JavaScript instead.
Related codes: 204 No Content (success, no body, no reset needed), 200 OK (success with body).
RFC: RFC 7231, Section 6.3.6.`,

	206: `Common causes: The client sent a Range header requesting only part of the resource, and the server is delivering that portion.
Real-world usage: Video/audio streaming (seeking to a specific timestamp), resuming interrupted file downloads, and parallel download managers that split files into chunks. The response includes Content-Range headers indicating which bytes are being delivered.
Related codes: 200 OK (full resource), 416 Range Not Satisfiable (requested range is invalid).
RFC: RFC 7233, Section 4.1.`,

	207: `Common causes: A WebDAV request that operated on multiple resources, each with its own status.
Real-world usage: WebDAV file operations (PROPFIND, PROPPATCH) that query or modify multiple files/folders at once. The XML body contains individual status codes for each sub-resource. Also used by some non-WebDAV APIs (e.g., Microsoft Graph batch requests).
Related codes: 200 OK (single-resource success), 208 Already Reported (avoids repeating members in a binding).
RFC: RFC 4918, Section 11.1.`,

	208: `Common causes: In a WebDAV multistatus response, members of a DAV binding have already been listed in an earlier part of the response.
Real-world usage: Very rare outside WebDAV. Prevents duplicate enumeration of the same resources when a collection has multiple bindings (essentially, hard links for WebDAV resources).
Related codes: 207 Multi-Status (the response format that contains 208 entries), 508 Loop Detected (infinite loop in bindings).
RFC: RFC 5842, Section 7.1.`,

	226: `Common causes: The server applied one or more instance-manipulations (delta encoding) to the resource.
Real-world usage: Extremely rare. Was designed for efficient bandwidth usage where the server sends only the differences (delta) from a previous version of the resource. Never gained wide adoption in browsers or servers.
Related codes: 200 OK (full resource delivery), 304 Not Modified (resource unchanged).
RFC: RFC 3229, Section 10.4.1.`,

	300: `Common causes: The requested resource has multiple representations, and the server is presenting the options for the client to choose.
Real-world usage: Rare in practice. Could theoretically be used for content negotiation where a resource exists in multiple formats (JSON, XML, HTML) or languages. Most servers perform server-driven negotiation and return the best match directly as 200.
Related codes: 301/302 (server-chosen redirect), 406 Not Acceptable (no suitable representation found).
RFC: RFC 7231, Section 6.4.1.`,

	301: `Common causes: The resource has permanently moved to a new URL. All future requests should use the new URL.
Real-world usage: Domain migrations (http to https), URL restructuring, and permanent content moves. Search engines transfer SEO rank to the new URL. Browsers and clients cache this redirect permanently. Warning: some clients change POST to GET on redirect (use 308 to preserve method).
Related codes: 302 Found (temporary redirect), 307 Temporary Redirect (preserves method, temporary), 308 Permanent Redirect (preserves method, permanent).
Troubleshooting: If a 301 redirect seems "stuck," clear the browser cache — browsers aggressively cache permanent redirects.
RFC: RFC 7231, Section 6.4.2.`,

	302: `Common causes: The resource temporarily resides at a different URL. The client should continue using the original URL for future requests.
Real-world usage: Post-login redirects, temporary maintenance redirects, and A/B testing. Very commonly used on the web, though its historical behavior is inconsistent — older clients changed POST to GET on redirect.
Related codes: 301 Moved Permanently (permanent redirect), 303 See Other (always changes to GET), 307 Temporary Redirect (preserves method).
Troubleshooting: If your POST data is lost after a 302 redirect, the client is likely changing the method to GET. Use 307 instead.
RFC: RFC 7231, Section 6.4.3.`,

	303: `Common causes: After processing a POST (or PUT/DELETE), the server redirects the client to a GET endpoint to retrieve the result. This is the POST/Redirect/GET pattern.
Real-world usage: Web forms that redirect to a confirmation page after submission. This prevents duplicate form submissions if the user refreshes the page. The client always uses GET to follow a 303 redirect, regardless of the original method.
Related codes: 302 Found (ambiguous method change), 307 Temporary Redirect (preserves method).
RFC: RFC 7231, Section 6.4.4.`,

	304: `Common causes: The client's cached copy of the resource is still valid. The server checked the If-Modified-Since or If-None-Match headers and determined no update is needed.
Real-world usage: Browser caching — almost every cached resource on a web page is validated with conditional requests. CDNs use 304 extensively to avoid retransmitting unchanged assets. Reduces bandwidth and improves page load times.
Related codes: 200 OK (full response when resource has changed), 412 Precondition Failed (different conditional semantics).
Troubleshooting: If you're not getting 304s when expected, check that ETag or Last-Modified headers are being sent by the server.
RFC: RFC 7232, Section 4.1.`,

	305: `Common causes: The server requires the client to access the resource through a specific proxy.
Real-world usage: Deprecated and almost never used. Was intended for servers to redirect clients to a proxy, but this was a security concern because it could be used to redirect traffic through malicious proxies. Most browsers ignore this status code.
Related codes: 407 Proxy Authentication Required (proxy needs credentials).
RFC: RFC 7231, Section 6.4.5 (deprecated).`,

	306: `Common causes: No longer used. Was defined in an earlier HTTP specification but subsequently removed.
Real-world usage: This code is reserved and should not be used. It was originally intended to mean "Subsequent requests should use the specified proxy." No modern HTTP client or server uses this code.
Related codes: 305 Use Proxy (also deprecated), 307 Temporary Redirect (the modern alternative for temporary redirects).
RFC: RFC 7231, Section 6.4.6 (reserved).`,

	307: `Common causes: The resource temporarily resides at a different URL. Unlike 302, the client must not change the HTTP method when following the redirect.
Real-world usage: HTTP-to-HTTPS redirects where the original request method must be preserved (e.g., a POST with a body). HSTS (HTTP Strict Transport Security) uses internal 307 redirects. Also used for temporary API endpoint migrations.
Related codes: 302 Found (may change method), 308 Permanent Redirect (permanent, preserves method), 301 Moved Permanently (permanent, may change method).
RFC: RFC 7231, Section 6.4.7.`,

	308: `Common causes: The resource has permanently moved and the client must use the new URL with the same HTTP method.
Real-world usage: Permanent URL migrations where POST, PUT, and other methods must be preserved. Used by APIs that permanently relocate endpoints. Similar to 301 but guarantees the method and body are not changed.
Related codes: 301 Moved Permanently (permanent, may change method), 307 Temporary Redirect (temporary, preserves method).
RFC: RFC 7538, Section 3.`,

	400: `Common causes: Malformed JSON/XML body, missing required parameters, invalid query strings, request body exceeding limits, or incorrect Content-Type header.
Real-world usage: The most common client error in APIs. Typically returned when input validation fails. Many APIs include a JSON body with specific field-level error details. Also returned by web servers for genuinely malformed HTTP syntax.
Related codes: 422 Unprocessable Content (syntactically valid but semantically wrong), 415 Unsupported Media Type (wrong Content-Type).
Troubleshooting: Check the request body format, Content-Type header, required fields, and encoding. Compare your request with the API documentation.
RFC: RFC 7231, Section 6.5.1.`,

	401: `Common causes: Missing Authorization header, expired token, invalid credentials, or revoked API key.
Real-world usage: Returned when an API or web resource requires authentication and the request either lacks credentials or the credentials are invalid. The response must include a WWW-Authenticate header indicating the expected auth scheme (Bearer, Basic, etc.).
Related codes: 403 Forbidden (authenticated but not authorized — you're logged in but don't have permission), 407 Proxy Authentication Required (proxy needs auth).
Troubleshooting: Check that the Authorization header is present and correctly formatted. Verify the token hasn't expired. For Basic auth, ensure the credentials are base64-encoded.
RFC: RFC 7235, Section 3.1.`,

	402: `Common causes: Reserved code, but used by some services when payment is required or a billing limit is reached.
Real-world usage: Not standardized, but some services use it creatively: Stripe returns 402 for failed payments, Shopify for unpaid store fees, and Google APIs for quota exceeded. Could see wider use with web monetization and micropayment standards.
Related codes: 403 Forbidden (general access denial), 429 Too Many Requests (rate limit hit).
RFC: RFC 7231, Section 6.5.2 (reserved for future use).`,

	403: `Common causes: Valid authentication but insufficient permissions, IP-based restrictions, accessing a resource that requires a specific role or subscription, or server-side access rules blocking the request.
Real-world usage: User is logged in but lacks the right role/permission (e.g., a regular user trying to access an admin endpoint). Also used for IP blocklists, geographic restrictions, and WAF (Web Application Firewall) blocks.
Related codes: 401 Unauthorized (not authenticated — need to log in), 404 Not Found (sometimes used instead of 403 to hide resource existence).
Troubleshooting: Verify user permissions, check IP restrictions, review server access rules. Some APIs return 403 with a JSON body explaining the specific permission that's missing.
RFC: RFC 7231, Section 6.5.3.`,

	404: `Common causes: Typo in the URL, deleted resource, incorrect route configuration, undeployed endpoint, or missing file on disk.
Real-world usage: The most recognized HTTP error. Returned when a URL doesn't map to any resource. In REST APIs, returned for GET/PUT/DELETE on a resource ID that doesn't exist. Some security-conscious APIs return 404 instead of 403 to hide the existence of resources from unauthorized users.
Related codes: 410 Gone (resource existed but was permanently removed — use this when you know it's intentional), 405 Method Not Allowed (URL exists but method is wrong).
Troubleshooting: Check for URL typos, verify the resource exists, check route configuration, and ensure the server/application is properly deployed.
RFC: RFC 7231, Section 6.5.4.`,

	405: `Common causes: Sending a GET to a POST-only endpoint, PUT to a read-only resource, or DELETE to a resource that doesn't support deletion.
Real-world usage: REST APIs return this when the URL exists but the HTTP method isn't supported for that endpoint. The response must include an Allow header listing the valid methods (e.g., "Allow: GET, POST").
Related codes: 404 Not Found (URL doesn't exist at all), 501 Not Implemented (method not recognized by server).
Troubleshooting: Check the Allow header in the response to see which methods are valid. Verify you're using the correct HTTP method per the API documentation.
RFC: RFC 7231, Section 6.5.5.`,

	406: `Common causes: The client's Accept header requests a content type the server can't produce (e.g., Accept: application/xml when the server only serves JSON).
Real-world usage: Content negotiation failures. Less common in practice because many APIs only support one content type. More relevant for APIs that support multiple formats. Some APIs return 406 when API versioning via Accept header doesn't match.
Related codes: 415 Unsupported Media Type (server can't read the client's format), 300 Multiple Choices (server offers alternatives).
RFC: RFC 7231, Section 6.5.6.`,

	407: `Common causes: A proxy server between the client and the target server requires authentication.
Real-world usage: Corporate proxy servers that require user authentication before allowing outbound HTTP requests. The response includes a Proxy-Authenticate header specifying the auth scheme.
Related codes: 401 Unauthorized (the target server needs auth), 403 Forbidden (access denied after auth).
Troubleshooting: Configure your HTTP client with proxy credentials. Check environment variables like HTTP_PROXY and HTTPS_PROXY and their associated credential settings.
RFC: RFC 7235, Section 3.2.`,

	408: `Common causes: The client took too long to send the complete request (headers or body). The server's patience ran out.
Real-world usage: Slow or unreliable network connections, very large file uploads on slow connections, or a client that opened a connection but stalled. Servers use this to free up connections from idle clients. Load balancers like AWS ALB return 408 for idle timeout.
Related codes: 504 Gateway Timeout (upstream server didn't respond in time), 429 Too Many Requests (rate limiting, not timeout).
Troubleshooting: Check network connectivity, increase client timeout settings, or use chunked transfer encoding for large payloads.
RFC: RFC 7231, Section 6.5.7.`,

	409: `Common causes: Concurrent modification conflicts, attempting to create a resource that already exists, or state-based conflicts (e.g., trying to delete a resource that's in use).
Real-world usage: Optimistic locking failures in REST APIs (two users editing the same resource), duplicate key errors, or trying to transition a resource to an invalid state. The response body usually explains the conflict.
Related codes: 412 Precondition Failed (conditional request failed due to ETag mismatch), 422 Unprocessable Content (semantic validation error).
Troubleshooting: Re-fetch the resource, merge changes, and retry. Use ETags and If-Match headers to detect conflicts.
RFC: RFC 7231, Section 6.5.8.`,

	410: `Common causes: The resource was intentionally and permanently removed by the server owner.
Real-world usage: Used when an API endpoint is permanently retired, a user deletes their account, or content is taken down permanently. Tells search engines to remove the URL from their index (stronger signal than 404). Useful for API versioning when old versions are sunset.
Related codes: 404 Not Found (resource might come back or might never have existed), 301 Moved Permanently (resource moved, not deleted).
Troubleshooting: This is intentional — the resource is gone. Check if the API has a new version or if the resource moved to a new URL.
RFC: RFC 7231, Section 6.5.9.`,

	411: `Common causes: The client sent a request without a Content-Length header for a method that requires it.
Real-world usage: Some servers and proxies require Content-Length for POST/PUT requests rather than accepting chunked transfer encoding. This is a server configuration requirement.
Related codes: 400 Bad Request (general malformed request), 413 Payload Too Large (body too big).
Troubleshooting: Add the Content-Length header to your request. If using chunked encoding, switch to a known-length body.
RFC: RFC 7231, Section 6.5.10.`,

	412: `Common causes: The client sent conditional headers (If-Match, If-Unmodified-Since) and the condition evaluated to false because the resource changed.
Real-world usage: Optimistic concurrency control — the client fetches a resource with an ETag, modifies it, and sends a PUT with "If-Match" to ensure no one else modified it. If the ETag doesn't match, the server returns 412.
Related codes: 409 Conflict (general conflict, not necessarily conditional), 304 Not Modified (conditional GET, not PUT).
RFC: RFC 7232, Section 4.2.`,

	413: `Common causes: Request body exceeds the server's configured maximum size limit.
Real-world usage: File upload limits (e.g., Nginx default is 1MB), API request body limits, or form data exceeding server configuration. The server may close the connection or include a Retry-After header if the condition is temporary.
Related codes: 400 Bad Request (general client error), 414 URI Too Long (URL too long instead of body too large).
Troubleshooting: Check the server's body size limit configuration (e.g., client_max_body_size in Nginx, LimitRequestBody in Apache). Consider chunked uploads for large files.
RFC: RFC 7231, Section 6.5.11.`,

	414: `Common causes: Extremely long query strings, often from stuffing too much data into GET parameters instead of using POST.
Real-world usage: Rare under normal use. Can occur with deeply nested search filters encoded in the URL, overly long redirect chains with accumulated query params, or misconfigured URL rewriting. Most servers limit URLs to 8KB-64KB.
Related codes: 400 Bad Request (general client error), 413 Payload Too Large (body too large).
Troubleshooting: Move large parameters into a POST request body. Check for redirect loops that accumulate query parameters.
RFC: RFC 7231, Section 6.5.12.`,

	415: `Common causes: The Content-Type header doesn't match what the server expects (e.g., sending text/plain when the server requires application/json).
Real-world usage: Common in REST APIs when the client forgets the Content-Type header or sends the wrong one. Also occurs when uploading files in an unsupported format.
Related codes: 406 Not Acceptable (the response format isn't acceptable to the client — the inverse), 400 Bad Request (general error).
Troubleshooting: Check the Content-Type header matches the API's expected format. Don't forget charset if required (e.g., application/json; charset=utf-8).
RFC: RFC 7231, Section 6.5.13.`,

	416: `Common causes: The Range header requests bytes beyond the end of the file, or the range format is invalid.
Real-world usage: Occurs when resuming a download after the file has been truncated or replaced, or when a media player seeks past the end of a video/audio file. The response includes Content-Range with the actual size.
Related codes: 206 Partial Content (successful range request), 200 OK (server ignores Range and returns full resource).
Troubleshooting: Check the file size on the server and ensure your Range header is within bounds.
RFC: RFC 7233, Section 4.4.`,

	417: `Common causes: The server cannot meet the requirements of the Expect header (usually "Expect: 100-continue").
Real-world usage: Uncommon. Returned when a server or intermediary proxy does not support the Expect header. HTTP/1.0 servers commonly return this because they don't understand the Expect mechanism.
Related codes: 100 Continue (the positive response to Expect: 100-continue).
Troubleshooting: Remove the Expect header from the request, or configure the client to not send it.
RFC: RFC 7231, Section 6.5.14.`,

	418: `Common causes: This is a joke status code from the Hyper Text Coffee Pot Control Protocol (HTCPCP).
Real-world usage: Used as an Easter egg by many websites and APIs (Google has a famous one). Some services use it as a humorous alternative to 403 for blocked requests. The IETF has confirmed this code is not going to be assigned for any real purpose, so it remains a permanent joke.
Related codes: 403 Forbidden (what you'd use in production instead).
RFC: RFC 2324 (April Fools' 1998), RFC 7168 (HTCPCP-TEA, April Fools' 2014).`,

	421: `Common causes: HTTP/2 connection coalescing directed a request to a server that doesn't handle that hostname.
Real-world usage: In HTTP/2, multiple hostnames can share a single TCP connection if they resolve to the same IP and share a TLS certificate. If the server can't actually serve one of those hostnames, it returns 421. The client should retry on a new connection.
Related codes: 400 Bad Request (general client error), 503 Service Unavailable (server overloaded).
RFC: RFC 7540, Section 9.1.2.`,

	422: `Common causes: The request body is syntactically valid (e.g., well-formed JSON) but contains semantic errors (e.g., invalid field values, business logic violations).
Real-world usage: Very common in REST APIs for validation errors. For example: email format is invalid, a date is in the past when it must be future, or a referenced entity doesn't exist. Originally WebDAV, but now widely adopted. The response body typically lists specific field errors.
Related codes: 400 Bad Request (syntax-level errors like malformed JSON), 409 Conflict (state conflict rather than validation).
RFC: RFC 4918, Section 11.2 (originally); RFC 9110, Section 15.5.21 (HTTP Semantics).`,

	423: `Common causes: A WebDAV LOCK is held on the resource, preventing modification.
Real-world usage: WebDAV file locking to prevent concurrent edits. Some collaborative editing systems use this concept. If a file is locked by another user, attempts to modify it return 423 until the lock is released.
Related codes: 409 Conflict (general conflict), 424 Failed Dependency (related request failed).
RFC: RFC 4918, Section 11.3.`,

	424: `Common causes: A WebDAV request that depended on another sub-request which failed.
Real-world usage: In a WebDAV batch operation (e.g., MOVE involving multiple files), if one file operation fails, dependent operations return 424. Occasionally used in non-WebDAV APIs for cascading failures in batch requests.
Related codes: 207 Multi-Status (the response envelope), 423 Locked (one common cause of dependency failure).
RFC: RFC 4918, Section 11.4.`,

	425: `Common causes: The server received a request via TLS early data (0-RTT) and is unwilling to process it because it might be a replay.
Real-world usage: TLS 1.3 early data (0-RTT) allows clients to send data in the first TLS flight, but this data can be replayed by an attacker. Servers return 425 for non-idempotent requests (e.g., POST) received as early data to prevent replay attacks.
Related codes: 400 Bad Request (general client error), 503 Service Unavailable (general retry scenario).
RFC: RFC 8470.`,

	426: `Common causes: The server requires the client to upgrade to a different protocol (e.g., TLS, HTTP/2, WebSocket).
Real-world usage: A server that requires HTTPS can return 426 with an Upgrade header indicating "TLS/1.3". Also used when a WebSocket endpoint receives a plain HTTP request without the Upgrade header. In practice, most HTTPS enforcement uses 301/308 redirects instead.
Related codes: 101 Switching Protocols (successful protocol upgrade), 301 Moved Permanently (more common for HTTP-to-HTTPS).
RFC: RFC 7231, Section 6.5.15.`,

	428: `Common causes: The server requires the request to include conditional headers (If-Match, If-Unmodified-Since) to prevent lost updates.
Real-world usage: Servers that enforce optimistic concurrency control. Without conditional headers, concurrent updates could silently overwrite each other. The server returns 428 to force clients to use ETags, preventing the "lost update" problem.
Related codes: 412 Precondition Failed (conditional headers present but condition failed), 409 Conflict (conflict detected by other means).
RFC: RFC 6585, Section 3.`,

	429: `Common causes: The client exceeded the server's rate limit (too many requests in a given time window).
Real-world usage: Extremely common in APIs. Rate limiting protects servers from abuse and overload. The response usually includes a Retry-After header indicating when to retry. Different API providers use different rate limit windows (per second, minute, hour, or day).
Related codes: 503 Service Unavailable (server overloaded, not rate limit), 403 Forbidden (blocked, not rate-limited).
Troubleshooting: Check the Retry-After header. Implement exponential backoff. Review the API's rate limit documentation. Consider caching responses to reduce request volume.
RFC: RFC 6585, Section 4.`,

	431: `Common causes: Individual header fields or the total header size exceeds server limits. Often caused by large cookies, excessively long Authorization tokens, or many custom headers.
Real-world usage: Can occur when cookies accumulate (especially across subdomains), with very large JWT tokens, or when proxies add many headers. Server limits vary: Nginx defaults to 8KB for headers, Apache to 8KB per line.
Related codes: 400 Bad Request (general malformed request), 414 URI Too Long (URL too long).
Troubleshooting: Clear cookies for the domain, reduce token size, or increase the server's header size limit (e.g., large_client_header_buffers in Nginx).
RFC: RFC 6585, Section 5.`,

	451: `Common causes: A legal demand (court order, government directive, DMCA takedown) requires the server to deny access.
Real-world usage: Government censorship, DMCA takedowns, GDPR right-to-be-forgotten removals, and court-ordered content blocking. The response should include a Link header pointing to the legal authority responsible. Named after Ray Bradbury's novel Fahrenheit 451.
Related codes: 403 Forbidden (access denied for non-legal reasons), 410 Gone (permanently removed but not for legal reasons).
RFC: RFC 7725.`,

	500: `Common causes: Unhandled exceptions, null pointer dereferences, database connection failures, misconfigured servers, or bugs in application code.
Real-world usage: The generic "something went wrong" error. In production, the response body should not reveal internal details (stack traces, SQL queries) as this is a security risk. Good APIs return a correlation/request ID for debugging.
Related codes: 502 Bad Gateway (upstream server error), 503 Service Unavailable (overloaded/maintenance), 504 Gateway Timeout (upstream timeout).
Troubleshooting: Check server logs for stack traces. Look for recent deployments that may have introduced bugs. Verify database and external service connectivity. Check disk space and memory usage.
RFC: RFC 7231, Section 6.6.1.`,

	501: `Common causes: The server doesn't support the HTTP method used in the request (e.g., PROPFIND on a non-WebDAV server), or a feature hasn't been implemented yet.
Real-world usage: Returned when a server encounters an HTTP method it doesn't understand at all (vs. 405, where the method is known but not allowed for that specific resource). Also used as a placeholder during API development for endpoints not yet implemented.
Related codes: 405 Method Not Allowed (method recognized but not allowed for this resource), 500 Internal Server Error (server error, not missing implementation).
RFC: RFC 7231, Section 6.6.2.`,

	502: `Common causes: A reverse proxy or load balancer received an invalid or no response from the upstream application server.
Real-world usage: Very common in microservice architectures. The upstream server crashed, returned garbage, closed the connection prematurely, or the response was malformed. Often seen with Nginx, HAProxy, AWS ALB/ELB, or Cloudflare when the origin server fails.
Related codes: 503 Service Unavailable (upstream is overloaded), 504 Gateway Timeout (upstream didn't respond in time), 500 Internal Server Error (the origin server itself errored).
Troubleshooting: Check the upstream server's health and logs. Verify the proxy configuration (upstream URL, port). Check for network issues between the proxy and upstream. Look at proxy error logs for specific connection errors.
RFC: RFC 7231, Section 6.6.3.`,

	503: `Common causes: Server overload, scheduled maintenance, resource exhaustion (CPU, memory, connections), or deployment in progress.
Real-world usage: Returned when the server is temporarily unable to handle requests. Should include a Retry-After header indicating when the client should try again. CDNs and load balancers often show custom 503 pages during maintenance windows.
Related codes: 502 Bad Gateway (upstream error), 504 Gateway Timeout (upstream timeout), 429 Too Many Requests (rate limiting).
Troubleshooting: Check server resource usage (CPU, memory, disk, connections). Review recent deployments. Check for cascading failures from downstream services. Verify health check endpoints.
RFC: RFC 7231, Section 6.6.4.`,

	504: `Common causes: A reverse proxy or gateway timed out waiting for a response from the upstream server.
Real-world usage: Common in microservice architectures when an upstream service is slow or unresponsive. Often seen when a database query takes too long, an external API call hangs, or the application is deadlocked. Load balancers have configurable timeout thresholds.
Related codes: 502 Bad Gateway (upstream returned an invalid response), 408 Request Timeout (server timed out waiting for the client), 503 Service Unavailable (upstream is down entirely).
Troubleshooting: Increase proxy timeout settings. Optimize slow upstream operations. Add request timeouts to upstream service calls. Check for deadlocks or resource contention. Consider async processing for long operations.
RFC: RFC 7231, Section 6.6.5.`,

	505: `Common causes: The client used an HTTP version the server doesn't support (e.g., HTTP/0.9 or HTTP/3 on a server that only supports HTTP/1.1).
Real-world usage: Rare in practice since most servers support HTTP/1.0 and HTTP/1.1, and HTTP/2 negotiation happens at the TLS layer. Could occur with very old clients or misconfigured custom HTTP implementations.
Related codes: 400 Bad Request (general protocol error), 426 Upgrade Required (server wants a different protocol).
RFC: RFC 7231, Section 6.6.6.`,

	506: `Common causes: A misconfiguration where the server's content negotiation setup creates a circular reference — the chosen variant itself requires negotiation.
Real-world usage: Extremely rare. Indicates a server configuration bug in transparent content negotiation. If encountered, it's almost certainly a server-side misconfiguration that needs to be fixed by the server administrator.
Related codes: 300 Multiple Choices (normal content negotiation), 500 Internal Server Error (general server error).
RFC: RFC 2295, Section 8.1.`,

	507: `Common causes: The server ran out of disk space or storage quota while trying to store the request's content.
Real-world usage: WebDAV file operations (PUT, COPY, MOVE) when the server's file system is full or the user's quota is exceeded. Some cloud storage APIs also use this code when storage limits are reached.
Related codes: 413 Payload Too Large (request too big, not storage full), 500 Internal Server Error (general server error).
Troubleshooting: Free up disk space on the server, increase storage quotas, or clean up old files.
RFC: RFC 4918, Section 11.5.`,

	508: `Common causes: The server detected an infinite loop while processing a WebDAV request with internal references.
Real-world usage: Very rare. Occurs in WebDAV when a PROPFIND with "Depth: infinity" encounters circular bindings (similar to symbolic link loops in a file system). The server aborts the request to prevent infinite processing.
Related codes: 208 Already Reported (used to avoid duplicate reporting in multistatus), 507 Insufficient Storage (another WebDAV error).
RFC: RFC 5842, Section 7.2.`,

	510: `Common causes: The request requires additional HTTP extensions (specified via the Extension Framework) that the server doesn't understand.
Real-world usage: Extremely rare. Part of the HTTP Extension Framework which never gained significant adoption. In theory, the server should inform the client which extensions are needed.
Related codes: 501 Not Implemented (server doesn't support the method), 426 Upgrade Required (protocol upgrade needed).
RFC: RFC 2774, Section 7.`,

	511: `Common causes: A captive portal (Wi-Fi hotspot, hotel network, airline WiFi) is intercepting the request and requiring authentication before granting internet access.
Real-world usage: Common in public WiFi networks. When you connect to a coffee shop or airport WiFi, the captive portal intercepts your HTTP requests and returns 511 until you log in or accept terms of service. The response body typically contains a login page.
Related codes: 401 Unauthorized (server-level auth, not network-level), 407 Proxy Authentication Required (proxy auth).
Troubleshooting: Open a browser and navigate to any HTTP (not HTTPS) page to trigger the captive portal login. Some devices detect captive portals automatically.
RFC: RFC 6585, Section 6.`,
}

// verboseSectionLabels maps known verbose section prefixes to their display colors.
var verboseSectionLabels = []struct {
	prefix string
	color  string
}{
	{"Common causes:", yellow},
	{"Real-world usage:", green},
	{"Related codes:", cyan},
	{"Troubleshooting:", red},
	{"RFC:", blue},
}

func printStatus(codeStr string, verbose bool) {
	code, err := strconv.Atoi(codeStr)
	if err != nil {
		fmt.Printf("Invalid code: %s\n", codeStr)
		return
	}

	color := statusColor(code)

	if description, ok := statusDescriptions[code]; ok {
		fmt.Printf("%s%s%d: %s%s — %s%s%s\n", bold, color, code, http.StatusText(code), reset, dim, description, reset)
	} else {
		text := http.StatusText(code)
		if text == "" {
			fmt.Printf("Unknown code: %d\n", code)
		} else {
			fmt.Printf("%s%s%d: %s%s\n", bold, color, code, text, reset)
		}
	}

	if verbose {
		if ext, ok := verboseDescriptions[code]; ok {
			fmt.Printf("\n%s───%s\n\n", dim, reset)
			lines := strings.Split(ext, "\n")
			prevWasSection := false
			for _, line := range lines {
				colored := false
				for _, sec := range verboseSectionLabels {
					if strings.HasPrefix(strings.TrimSpace(line), sec.prefix) {
						if prevWasSection {
							fmt.Println()
						}
						idx := strings.Index(line, sec.prefix)
						before := line[:idx]
						after := line[idx+len(sec.prefix):]
						fmt.Printf("%s%s%s%s%s%s\n", before, bold, sec.color, sec.prefix, reset, after)
						colored = true
						prevWasSection = true
						break
					}
				}
				if !colored {
					fmt.Println(line)
					prevWasSection = false
				}
			}
		}
	}
}
