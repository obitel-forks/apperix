package apperix

import (
	"net/http"
	"encoding/json"
)

type Response interface {
	String() *[]byte
	Status() int
	Header(string, string)
	//server side error
	ReplyServerError(string, string)
	ReplyNotImplemented(string)
	//client side error
	ReplyClientError(string, string)
	ReplyForbidden(string)
	ReplyNotFound(string)
	//success
	ReplyCreated()
	//custom
	ReplyCustom(int)
	ReplyCustomError(int, string, string)
}

type ResponseJson struct {
	headers map[string] string
	data map[string] interface {}
	errorCode string
	errorMessage string
	status int
}

func (response *ResponseJson) String() *[]byte {
	if len(response.errorCode) > 0 {
		buffer, _ := json.Marshal(struct {
			Error struct {
				Code string `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		} {
			Error: struct {
				Code string `json:"code"`
				Message string `json:"message"`
			} {
				Code: response.errorCode,
				Message: response.errorMessage,
			},
		})
		return &buffer	
	}
	if len(response.data) < 1 {
		result := []byte("{\"data\":{}}")
		return &result
	}
	buffer, _ := json.Marshal(struct {
		Data map[string] interface{} `json:"data"`
	} {
		response.data,
	})
	return &buffer
}

func (response *ResponseJson) Status() int {
	if response.status == 0 {
		return http.StatusOK
	}
	return response.status
}

func (response *ResponseJson) Header(head string, value string) {
	response.headers[head] = value
}

func (response *ResponseJson) Data(key string, value interface{}) {
	if response.data == nil {
		response.data = make(map[string] interface {})
	}
	response.data[key] = value
}

func (response *ResponseJson) ReplyServerError(code string, message string) {
	response.status = http.StatusInternalServerError
	response.errorCode = code
	response.errorMessage = message
}

func (response *ResponseJson) ReplyNotImplemented(message string) {
	response.status = http.StatusNotImplemented
	response.errorCode = "NOT_IMPLEMENTED"
	response.errorMessage = message
}

func (response *ResponseJson) ReplyClientError(code string, message string) {
	response.status = http.StatusBadRequest
	response.errorCode = code
	response.errorMessage = message
}

func (response *ResponseJson) ReplyForbidden(message string) {
	response.status = http.StatusForbidden
	response.errorCode = "ACCESS_DENIED"
	response.errorMessage = message
}

func (response *ResponseJson) ReplyNotFound(message string) {
	response.status = http.StatusNotFound
	response.errorCode = "NOT_FOUND"
	response.errorMessage = message
}

func (response *ResponseJson) ReplyCreated() {
	response.status = http.StatusCreated
}

func (response *ResponseJson) ReplyCustom(status int) {
	response.status = status
}

func (response *ResponseJson) ReplyCustomError(status int, code string, message string) {
	response.status = status
	response.errorCode = code
	response.errorMessage = message
}
