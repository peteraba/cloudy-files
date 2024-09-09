package util

import "strings"

func NegotiateContentType(accept string, supportedTypes []string) string {
	if accept == "" {
		// No Accept header, assume the first supported type
		return supportedTypes[0]
	}

	// Parse the Accept header
	// See https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Accept
	// Supported formats are separated by commas, with optional parameters
	// separated by semicolons. We only care about the media type, which always
	// comes before the semicolon.
	acceptedTypes := strings.Split(accept, ",")
	for _, acceptedType := range acceptedTypes {
		mediaType := strings.TrimSpace(strings.Split(acceptedType, ";")[0])
		for _, supportedType := range supportedTypes {
			if mediaType == supportedType || mediaType == "*/*" {
				return supportedType
			}
		}
	}

	// No match found, assume the first supported type
	return supportedTypes[0]
}
