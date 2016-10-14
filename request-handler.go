package apperix

import (
	"fmt"
	"runtime"
	"strings"
	"sync/atomic"
	"net/http"
	"github.com/dgrijalva/jwt-go"
)

type targetResource struct {
	identifier string
	variables map[string] string
}

func parseAuth(authHeader string, signatureSecret []byte) (*Client, error) {
	defer func() {
		err := recover()
		if err == nil {
			return
		}
	}()
	client := Client {}
	if len(authHeader) < 1 {
		return &client, nil
	}
	if len(authHeader) != 197 {
		return nil, fmt.Errorf("Wrong access token length (%d)", len(authHeader))
	}
	//parse access token
	token, err := jwt.Parse(authHeader, func(token *jwt.Token) (interface{}, error) {
		//validate signing algorithm
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Wrong signing method: %v", token.Header["alg"])
		}
		return signatureSecret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims);
	if !ok || !token.Valid {
		return nil, fmt.Errorf("Invalid access token")
	}
	client.Identifier = &Identifier {}
	client.Identifier.FromString(claims["iss"].(string))
	return &client, nil
}

func identifyTargetResource(urlPath string, service *Service) (targetResource, error) {
	target := targetResource {
		variables: make(map[string] string),
	}
	tmpPath := strings.Split(urlPath, "/")
	path := make([]string, 0)
	for _, str := range tmpPath {
		if str != "" {
			path = append(path, str)
		}
	}
	currentResource := service.resources["root"]
	var numOfVar int
	var matched bool
	for _, segment := range path {
		matched = false
		child, err := currentResource.StaticChildIdentifier(segment)
		if err == nil {
			//static resource identified
			currentResource = service.resources[child]
		} else if numOfVar = currentResource.NumberOfVariableChildren(); numOfVar > 0 {
			//maybe the resource is a matching placeholder
			parentId := currentResource.Identifier()
			for itr := 0; itr < numOfVar; itr++ {
				child, err = currentResource.VariableChildIdentifier(itr)
				if err != nil {
					continue
				}
				currentResource = service.resources[child]
				if currentResource.(*variableResource).MatchPattern(segment) {
					matched = true
					break
				}
			}
			if !matched {
				//resource not found
				if parentId == "root" {
					return target, fmt.Errorf("Resource '%s' not found in root", segment)
				} else {
					return target, fmt.Errorf("Resource '%s' not found in '%s'", segment, parentId)
				}
			}
			target.variables[currentResource.Identifier()] = segment
		} else {
			//resource not found
			parentId := currentResource.Identifier()
			if parentId == "root" {
				return target, fmt.Errorf("Resource '%s' not found in root", segment)
			} else {
				return target, fmt.Errorf("Resource '%s' not found in '%s'", segment, parentId)
			}
		}
	}
	target.identifier = currentResource.Identifier()
	return target, nil
}

func writeReponse(data Response, response *http.ResponseWriter) {
	switch data.(type) {
	case *ResponseJson:
		(*response).Header().Set("Content-Type", "application/json")
	default:
		(*response).Header().Set("Content-Type", "text/plain")
	}
	(*response).WriteHeader(data.Status())
	(*response).Write([]byte(*data.String()))
	(*response).(http.Flusher).Flush()
}

func finishProcessing(service *Service) {
	//decrement amount of currently processed requests
	atomic.AddUint32(&service.reqsInProcess, ^uint32(0))
	if service.shutdownRequested && service.reqsInProcess < 1 {
		//shutdown service
	}
	service.syncGroup.Done()
}

type apperixRequestHandler struct {
	service *Service
}

func (handler *apperixRequestHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	//increment amount of currently processed requests
	atomic.AddUint32(&handler.service.reqsInProcess, 1)
	handler.service.syncGroup.Add(1)

	defer func() {
		//in case of unexpected panic
		if err := recover(); err != nil {
			buf := make([]byte, 1 << 16)
			fmt.Printf("ERROR: %s\n", err)
			fmt.Printf("STACKTRACE: %s\n", string(buf[0:runtime.Stack(buf, true)]))
			responseErr := ResponseJson {}
			responseErr.ReplyServerError(
				"INTERNAL_ERROR",
				fmt.Sprintf("Could not process request: %s", err),
			)
			writeReponse(&responseErr, &response)
		}
	}()
	defer finishProcessing(handler.service)

	//parse method
	var method Method
	switch request.Method {
	case "POST":
		method = CREATE
	case "GET":
		method = READ
	case "PUT":
		method = UPDATE
	case "DELETE":
		method = DELETE
	case "PATCH":
		method = PATCH
	case "HEAD":
		method = READ_HEADERS
	case "OPTIONS":
		method = READ_OPTIONS
	case "PURGE":
		method = PURGE
	case "COPY":
		method = COPY
	case "MOVE":
		method = MOVE
	case "LINK":
		method = LINK
	case "UNLINK":
		method = UNLINK
	case "LOCK":
		method = LOCK
	case "UNLOCK":
		method = UNLOCK
	case "PROPFIND":
		method = READ_PROPERTIES
	case "PROPPATCH":
		method = UPDATE_PROPERTIES
	case "MKCOL":
		method = CREATE_COLLECTION
	}

	//authenticate client
	client, err := parseAuth(
		request.Header.Get("Authorization"),
		handler.service.Config.JwtSignatureSecret(),
	)
	if err != nil {
		responseErr := ResponseJson {}
		responseErr.ReplyCustomError(
			http.StatusBadRequest,
			"INVALID_ACCESS_TOKEN",
			fmt.Sprintf("Invalid access token: %s", err),
		)
		writeReponse(&responseErr, &response)
		return
	}

	//identify target resource
	target, err := identifyTargetResource(request.URL.Path, handler.service)
	if err != nil {
		responseErr := ResponseJson {}
		responseErr.ReplyNotFound(fmt.Sprintf("%s", err))
		writeReponse(&responseErr, &response)
		return
	}

	//verify permissions
	allowed := false
	var permissions Permissions
	resourceId, err := handler.service.GetResourceIdentifier(
		target.identifier,
		target.variables,
	)
	if err != nil {
		panic(fmt.Errorf("Could not get resource identifier: %s", err))
	}
	if client.Identifier != nil {
		_, permissions, err = handler.service.ResolvePermissionsFor(
			resourceId,
			client.Identifier,
		)
		if err != nil {
			panic(fmt.Errorf(
				"Could not resolve permissions for '%s': %s",
				resourceId,
				err,
			))
		}
	} else {
		_, permissions, err = handler.service.ResolvePermissionsFor(
			resourceId,
			GUESTS,
		)
		if err != nil {
			panic(fmt.Errorf(
				"Could not resolve permissions for '%s': %s",
				resourceId,
				err,
			))
		}
	}
	if err != nil {
		panic(fmt.Errorf("Could not resolve permissions: %s", err))
	}
	switch method {
	case CREATE:
		allowed = permissions.Create
	case READ:
		allowed = permissions.Read
	case UPDATE:
		allowed = permissions.Update
	case DELETE:
		allowed = permissions.Delete
	case PATCH:
		allowed = permissions.Patch
	case READ_HEADERS:
		allowed = permissions.ReadHeaders
	case READ_OPTIONS:
		allowed = permissions.ReadOptions
	case PURGE:
		allowed = permissions.Purge
	case COPY:
		allowed = permissions.Copy
	case MOVE:
		allowed = permissions.Move
	case LINK:
		allowed = permissions.Link
	case UNLINK:
		allowed = permissions.Unlink
	case LOCK:
		allowed = permissions.Lock
	case UNLOCK:
		allowed = permissions.Unlock
	case READ_PROPERTIES:
		allowed = permissions.ReadProperties
	case UPDATE_PROPERTIES:
		allowed = permissions.UpdateProperties
	case CREATE_COLLECTION:
		allowed = permissions.CreateCollection
	}
	if !allowed {
		responseErr := ResponseJson {}
		responseErr.ReplyForbidden(
			"Insufficient permissions",
		)
		writeReponse(&responseErr, &response)
		return
	}

	//verify method support
	handlerFunction, err := handler.service.resources[target.identifier].Handler(method)
	if err != nil {
		//method not supported
		responseErr := ResponseJson {}
		responseErr.ReplyCustomError(
			http.StatusMethodNotAllowed,
			"METHOD_NOT_SUPPORTED",
			fmt.Sprintf(
				"Method '%s' not supported on resource '%s'",
				request.Method,
				target.identifier,
			),
		)
		writeReponse(&responseErr, &response)
		return
	}

	//execute handler
	if request.Header["Content-Type"][0] == "multipart/form-data" {
		err = request.ParseMultipartForm(65536)
		if err != nil {
			panic(fmt.Sprintf("Could not parse multipart form data: %s", err))
		}
	}
	responseData := handlerFunction(
		client,
		&Request {
			requestObject: request,
			Parameters: request.URL.Query(),
		},
		handler.service,
	)
	writeReponse(responseData, &response)
}
