package apperix

import (
	"fmt"
	"bytes"
	"net/http"
	"net/url"
)

type Request struct {
	requestObject *http.Request
	Parameters url.Values
}

func (req *Request) Data(key string) string {
	return req.requestObject.FormValue(key)
}

func (req *Request) File(key string) (
	metadata map[string] string,
	data bytes.Buffer,
	err error,
) {
	metadata = make(map[string] string)
	file, header, err := req.requestObject.FormFile(key)
	if err != nil {
		return metadata, data, NotFoundError {
			message: fmt.Sprintf("Could not find file '%s'", key),
		}
	}
	_, err = data.ReadFrom(file)
	if err != nil {
		return metadata, data, fmt.Errorf("Could not read file '%s': %s\n", key, err)
	}
	metadata["name"] = header.Filename
	metadata["type"] = header.Header["Content-Type"][0]
	return metadata, data, nil
}