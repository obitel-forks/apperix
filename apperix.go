package apperix

import (
	"io/ioutil"
	"fmt"
	"time"
	"strconv"
	"regexp"
	"net"
	"net/http"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"crypto/tls"
)

type Method int
const (
	CREATE Method = iota
	READ
	UPDATE
	DELETE
	PATCH
	READ_HEADERS
	READ_OPTIONS
	PURGE
	COPY
	MOVE
	LINK
	UNLINK
	LOCK
	UNLOCK
	READ_PROPERTIES
	UPDATE_PROPERTIES
	CREATE_COLLECTION
)

type ResourceType int
const (
	STATIC ResourceType = iota
	VARIABLE
)

type Client struct {
	Identifier *Identifier
}

type Handler func(*Client, *Request, *Service) Response

type PermissionInheritance struct {
	Owner bool
	UserPermissions bool
	OtherUserPermissions bool
	GuestPermissions bool
}

type DefaultResourcePermissions struct {
	UserPermissions Permissions
	GuestPermissions Permissions
	Inheritance PermissionInheritance
}

type Resource struct {
	Permissions DefaultResourcePermissions
	Handlers map[Method] Handler
	Name string
	Parent string
	MaxUploadSize int64
	Type ResourceType
	Pattern string
}

type HashAlgorithm int
const (
	HS256 HashAlgorithm = iota
	HS384
	HS512
	RS256
	RS384
	RS512
	ES256
	ES384
	ES512
)

/*
	AuthenticationConfig bundles database related configurations.
*/
type DatabaseConfig struct {
	Location string
	Cache int
}

/*
	AuthenticationConfig bundles authentication related configurations.
*/
type AuthenticationConfig struct {
	Path string
	TokenExpiry time.Duration
	SignatureSecret string
}

/*
	NetworkConfig bundles network related configurations.
*/
type NetworkConfig struct {
	HttpPort uint16
	HttpsPort uint16
}

/*
	SecurityConfig bundles security related configurations.
*/
type SecurityConfig struct {
	Https bool
	Certificate string
	PrivateKey string
	HashAlgorithm HashAlgorithm
}

/*
	DefaultsConfig bundles default values and directories.
*/
type DefaultsConfig struct {
	MaxUploadSize int64
	UploadDirectory string
	AutoCleanUploads bool
}

/*
	ServiceConfig bundles all required configuration bundles.
*/
type ServiceConfig struct {
	Name string
	Database DatabaseConfig
	Authentication AuthenticationConfig
	Network NetworkConfig
	Security SecurityConfig
	Defaults DefaultsConfig
	Resources map[string] Resource
}


/*
	setupDatabase sets up the given database
	creating required tables if necessary.
*/
func setupDatabase(database *sql.DB) {
	_, err := database.Exec(`
		CREATE TABLE IF NOT EXISTS resources (
			id INTEGER PRIMARY KEY,
			owner_id BLOB,
			str_id TEXT UNIQUE NOT NULL
		);
	`)
	if err != nil {
		panic(fmt.Errorf("Could not setup table: 'resources': %s", err))
	}
	_, err = database.Exec(`
		CREATE TABLE IF NOT EXISTS resource_permissions (
			resource_id INTEGER,
			user_id BLOB,
			permissions INTEGER NOT NULL,
			PRIMARY KEY(resource_id, user_id)
		);
	`)
	if err != nil {
		panic(fmt.Errorf("Could not setup table: 'resource_permissions': %s", err))
	}
	_, err = database.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id BLOB,
			username TEXT NOT NULL,
			password TEXT NOT NULL,
			PRIMARY KEY (id)
		);
	`)
	if err != nil {
		panic(fmt.Errorf("Could not setup table: 'users': %s", err))
	}

	_, err = database.Exec(`
		CREATE INDEX IF NOT EXISTS str_id
		ON resources (str_id);
	`)
	if err != nil {
		panic(fmt.Errorf("Could not create index: 'resources.resource_strid': %s", err))
	}

	_, err = database.Exec(`
		CREATE INDEX IF NOT EXISTS username
		ON users (username);
	`)
	if err != nil {
		panic(fmt.Errorf("Could not create index: 'users.username': %s", err))
	}
}

/*
	prepareDatabase initializes the database connection and prepares it for usage.
*/
func prepareDatabase(
	location string,
	serviceName string,
) (
	database *sql.DB,
	err error,
) {
	databasePath := ConcatStrings(
		location,
		"/",
		serviceName,
		".db",
	)
	database, err = sql.Open("sqlite3", databasePath)
	if err != nil {
		panic(fmt.Errorf("Could not open database %s, reason: %s", databasePath, err))
	}
	setupDatabase(database)
	return database, err
}

/*
	CreateService constructs and returns an initialized apperix service.

	CAUTION: Faulty configuration will cause panic!
*/
func CreateService(conf ServiceConfig) (
	service *Service,
) {
	service = &Service {
		resources: make(map[string] resourceObject),
		Config: configuration {
			name: conf.Name,
			https: conf.Security.Https,
			authConfig: conf.Authentication,
			networkConfig: conf.Network,
			defaultsConfig: conf.Defaults,
		},
	}
	service.shutdownRequested = false
	service.shutdownSignal = make(chan int)
	service.resources["root"] = &staticResource {
		identifier: "root",
		name: "",
		defaultPermissions: DefaultResourcePermissions {},
		handlers: make(map[Method] Handler),
		staticChildren: make(map[string] string),
		variableChildren: make([]string, 0),
	}

	//prepare certificate and key
	if conf.Security.Https {
		//read certificate
		cert, err := ioutil.ReadFile(conf.Security.Certificate)
		if err != nil {
			panic(fmt.Errorf(
				"Could not load certificate from '%s': %s",
				conf.Security.Certificate,
				err,
			))
		}

		//read private key
		pkey, err := ioutil.ReadFile(conf.Security.PrivateKey)
		if err != nil {
			panic(fmt.Errorf(
				"Could not load private key from '%s': %s",
				conf.Security.PrivateKey,
				err,
			))
		}

		service.Config.certificate = cert
		service.Config.privateKey = pkey
	}

	//register identifiers
	tmpIdRegistry := map[string] bool {}
	tmpIdRegistry["root"] = true
	for identifier, _ := range conf.Resources {
		tmpIdRegistry[identifier] = true
	}

	//verify resources
	var has bool
	for identifier, resource := range conf.Resources {
		if resource.Parent == "" {
			resource.Parent = "root"
		}
		//verify parent resource
		if _, parentRegistered := tmpIdRegistry[resource.Parent]; !parentRegistered {
			panic(fmt.Errorf("Resource '%s' referenced unregistered parent resource ('%s')", identifier, resource.Parent))
		}
		//verify reserved identifier: "auth"
		switch identifier {
		case "auth":
			panic(fmt.Errorf("Resource identifier 'auth' reserved"))
		case "root":
			service.resources["root"] = &staticResource {
				identifier: "root",
				name: "",
				parent: "",
				handlers: resource.Handlers,
				defaultPermissions: resource.Permissions,
				staticChildren: make(map[string] string),
				variableChildren: make([]string, 0),
			}
		}

		switch resource.Type {
		case STATIC:
			//verify name not registered
			has = false
			if service.resources[resource.Parent] != nil {
				has = service.resources[resource.Parent].HasStaticChild(resource.Name)
			}
			if has {
				if resource.Parent == "root" {
					panic(fmt.Errorf("Duplicate static resource name ('%s') in parent root", resource.Name))
				} else {
					panic(fmt.Errorf("Duplicate static resource name ('%s') in parent '%s'", resource.Name, resource.Parent))
				}
			}
			//verify no overlap with auth
			if resource.Parent == "root" && resource.Name == conf.Authentication.Path {
				panic(fmt.Errorf("Resource ('%s') overlaps with authentication path", identifier))
			}
			service.resources[identifier] = &staticResource {
				identifier: identifier,
				name: resource.Name,
				parent: resource.Parent,
				handlers: resource.Handlers,
				defaultPermissions: resource.Permissions,
				staticChildren: make(map[string] string),
				variableChildren: make([]string, 0),
			}
		case VARIABLE:
			//verify name not registered
			has = false
			if service.resources[resource.Parent] != nil {
				has = service.resources[resource.Parent].HasVariableChild(resource.Name)
			}
			if has {
				if resource.Parent == "root" {
					panic(fmt.Errorf("Duplicate variable resource name ('%s') in parent root", resource.Name))
				} else {
					panic(fmt.Errorf("Duplicate variable resource name ('%s') in parent '%s'", resource.Name, resource.Parent))
				}
			}
			//compile regexp
			regex, err := regexp.Compile(resource.Pattern)
			if err != nil {
				panic(fmt.Errorf("Pattern regex ('%s') compilation for resource ('%s') failed", resource.Pattern, identifier))
			}
			service.resources[identifier] = &variableResource {
				identifier: identifier,
				name: resource.Name,
				parent: resource.Parent,
				handlers: resource.Handlers,
				defaultPermissions: resource.Permissions,
				staticChildren: make(map[string] string),
				variableChildren: make([]string, 0),
				pattern: *regex,
			}
		default:
			panic(fmt.Errorf("Wrong resource type (%d)", resource.Type))
		}
	}

	//set children
	for identifier, resource := range conf.Resources {
		if resource.Parent == "" {
			resource.Parent = "root"
		}
		switch resource.Type {
		case STATIC:
			service.resources[resource.Parent].DefineStaticChild(identifier, resource.Name)
		case VARIABLE:
			service.resources[resource.Parent].DefineVariableChild(identifier)
		}
	}

	//prepare authentication resource
	service.resources["auth"] = &staticResource {
		identifier: "auth",
		name: conf.Authentication.Path,
		parent: "root",
		handlers: map[Method] Handler {
			READ: authReadHandler,
		},
		defaultPermissions: DefaultResourcePermissions {
			UserPermissions: Permissions {
				Read: true,
			},
			GuestPermissions: Permissions {
				Read: true,
			},
			Inheritance: PermissionInheritance {},
		},
		staticChildren: make(map[string] string),
		variableChildren: make([]string, 0),
	}
	service.resources["root"].DefineStaticChild("auth", "auth")

	//prepare database
	database, err := prepareDatabase(conf.Database.Location, conf.Name)
	if err != nil {
		panic(fmt.Errorf("Could not prepare database: %s", err))
	}
	service.database = database

	//initialize caches
	service.userProvider.initialize(database, 100)
	service.permissionProvider.initialize(database, 1000)
	service.ownerProvider.initialize(database, 1000)

	//initialize server
	port := conf.Network.HttpPort
	if conf.Security.Https {
		port = conf.Network.HttpsPort
	}
	service.server = &http.Server{
		Addr: net.JoinHostPort("", strconv.Itoa(int(port))),
		Handler: &apperixRequestHandler {
			service: service,
		},
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	if conf.Security.Https {
		certificate, err := tls.X509KeyPair(
			service.Config.certificate,
			service.Config.privateKey,
		)
		if err != nil {
			panic(fmt.Errorf("Could not parse TLS key-pair: %s", err))
		}
		service.server.TLSConfig = &tls.Config {
			CipherSuites: []uint16 {
				tls.TLS_RSA_WITH_RC4_128_SHA,
				tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,
				tls.TLS_RSA_WITH_AES_128_CBC_SHA,
				tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA,
				tls.TLS_RSA_WITH_AES_128_CBC_SHA,
				tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
			},
			ServerName: "qube-AAABBBCCC",
			Certificates: [] tls.Certificate {
				certificate,
			},
		}
	}

	return service
}