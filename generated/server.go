// Package generated provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.16.3 DO NOT EDIT.
package generated

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
	"github.com/oapi-codegen/runtime"
)

// GetAuthProviderCallbackParams defines parameters for GetAuthProviderCallback.
type GetAuthProviderCallbackParams struct {
	// State The state parameter containing redirect URI and anti-CSRF token.
	State string `form:"state" json:"state"`
}

// GetAuthProviderLoginParams defines parameters for GetAuthProviderLogin.
type GetAuthProviderLoginParams struct {
	// RedirectUri The URI to redirect the user to after authentication.
	RedirectUri string `form:"redirect_uri" json:"redirect_uri"`
}

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Retrieve a list of connected providers that the user is logged in with
	// (GET /auth/status)
	GetAuthStatus(w http.ResponseWriter, r *http.Request)
	// Handle OAuth callback and store tokens.
	// (GET /auth/{provider}/callback)
	GetAuthProviderCallback(w http.ResponseWriter, r *http.Request, provider string, params GetAuthProviderCallbackParams)
	// Redirect to the provider's OAuth login page.
	// (GET /auth/{provider}/login)
	GetAuthProviderLogin(w http.ResponseWriter, r *http.Request, provider string, params GetAuthProviderLoginParams)
	// Log out the user and invalidate the session.
	// (POST /auth/{provider}/logout)
	PostAuthProviderLogout(w http.ResponseWriter, r *http.Request, provider string)
	// Retrieve an OAuth token for a specific provider.
	// (GET /auth/{provider}/token)
	GetAuthProviderToken(w http.ResponseWriter, r *http.Request, provider string)
}

// Unimplemented server implementation that returns http.StatusNotImplemented for each endpoint.

type Unimplemented struct{}

// Handle OAuth callback and store tokens.
// (GET /auth/{provider}/callback)
func (_ Unimplemented) GetAuthProviderCallback(w http.ResponseWriter, r *http.Request, provider string, params GetAuthProviderCallbackParams) {
	w.WriteHeader(http.StatusNotImplemented)
}

// Redirect to the provider's OAuth login page.
// (GET /auth/{provider}/login)
func (_ Unimplemented) GetAuthProviderLogin(w http.ResponseWriter, r *http.Request, provider string, params GetAuthProviderLoginParams) {
	w.WriteHeader(http.StatusNotImplemented)
}

// Retrieve an OAuth token for a specific provider.
// (GET /auth/{provider}/token)
func (_ Unimplemented) GetAuthProviderToken(w http.ResponseWriter, r *http.Request, provider string) {
	w.WriteHeader(http.StatusNotImplemented)
}

// ServerInterfaceWrapper converts contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler            ServerInterface
	HandlerMiddlewares []MiddlewareFunc
	ErrorHandlerFunc   func(w http.ResponseWriter, r *http.Request, err error)
}

type MiddlewareFunc func(http.Handler) http.Handler

// GetAuthProviderCallback operation middleware
func (siw *ServerInterfaceWrapper) GetAuthProviderCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "provider" -------------
	var provider string

	err = runtime.BindStyledParameterWithLocation("simple", false, "provider", runtime.ParamLocationPath, chi.URLParam(r, "provider"), &provider)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "provider", Err: err})
		return
	}

	// Parameter object where we will unmarshal all parameters from the context
	var params GetAuthProviderCallbackParams

	// ------------- Required query parameter "state" -------------

	if paramValue := r.URL.Query().Get("state"); paramValue != "" {

	} else {
		siw.ErrorHandlerFunc(w, r, &RequiredParamError{ParamName: "state"})
		return
	}

	err = runtime.BindQueryParameter("form", true, true, "state", r.URL.Query(), &params.State)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "state", Err: err})
		return
	}

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetAuthProviderCallback(w, r, provider, params)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r.WithContext(ctx))
}

// GetAuthProviderLogin operation middleware
func (siw *ServerInterfaceWrapper) GetAuthProviderLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "provider" -------------
	var provider string

	err = runtime.BindStyledParameterWithLocation("simple", false, "provider", runtime.ParamLocationPath, chi.URLParam(r, "provider"), &provider)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "provider", Err: err})
		return
	}

	// Parameter object where we will unmarshal all parameters from the context
	var params GetAuthProviderLoginParams

	// ------------- Required query parameter "redirect_uri" -------------

	if paramValue := r.URL.Query().Get("redirect_uri"); paramValue != "" {

	} else {
		siw.ErrorHandlerFunc(w, r, &RequiredParamError{ParamName: "redirect_uri"})
		return
	}

	err = runtime.BindQueryParameter("form", true, true, "redirect_uri", r.URL.Query(), &params.RedirectUri)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "redirect_uri", Err: err})
		return
	}

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetAuthProviderLogin(w, r, provider, params)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r.WithContext(ctx))
}

// GetAuthProviderToken operation middleware
func (siw *ServerInterfaceWrapper) GetAuthProviderToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "provider" -------------
	var provider string

	err = runtime.BindStyledParameterWithLocation("simple", false, "provider", runtime.ParamLocationPath, chi.URLParam(r, "provider"), &provider)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "provider", Err: err})
		return
	}

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetAuthProviderToken(w, r, provider)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r.WithContext(ctx))
}

type UnescapedCookieParamError struct {
	ParamName string
	Err       error
}

func (e *UnescapedCookieParamError) Error() string {
	return fmt.Sprintf("error unescaping cookie parameter '%s'", e.ParamName)
}

func (e *UnescapedCookieParamError) Unwrap() error {
	return e.Err
}

type UnmarshalingParamError struct {
	ParamName string
	Err       error
}

func (e *UnmarshalingParamError) Error() string {
	return fmt.Sprintf("Error unmarshaling parameter %s as JSON: %s", e.ParamName, e.Err.Error())
}

func (e *UnmarshalingParamError) Unwrap() error {
	return e.Err
}

type RequiredParamError struct {
	ParamName string
}

func (e *RequiredParamError) Error() string {
	return fmt.Sprintf("Query argument %s is required, but not found", e.ParamName)
}

type RequiredHeaderError struct {
	ParamName string
	Err       error
}

func (e *RequiredHeaderError) Error() string {
	return fmt.Sprintf("Header parameter %s is required, but not found", e.ParamName)
}

func (e *RequiredHeaderError) Unwrap() error {
	return e.Err
}

type InvalidParamFormatError struct {
	ParamName string
	Err       error
}

func (e *InvalidParamFormatError) Error() string {
	return fmt.Sprintf("Invalid format for parameter %s: %s", e.ParamName, e.Err.Error())
}

func (e *InvalidParamFormatError) Unwrap() error {
	return e.Err
}

type TooManyValuesForParamError struct {
	ParamName string
	Count     int
}

func (e *TooManyValuesForParamError) Error() string {
	return fmt.Sprintf("Expected one value for %s, got %d", e.ParamName, e.Count)
}

// Handler creates http.Handler with routing matching OpenAPI spec.
func Handler(si ServerInterface) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{})
}

type ChiServerOptions struct {
	BaseURL          string
	BaseRouter       chi.Router
	Middlewares      []MiddlewareFunc
	ErrorHandlerFunc func(w http.ResponseWriter, r *http.Request, err error)
}

// HandlerFromMux creates http.Handler with routing matching OpenAPI spec based on the provided mux.
func HandlerFromMux(si ServerInterface, r chi.Router) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{
		BaseRouter: r,
	})
}

func HandlerFromMuxWithBaseURL(si ServerInterface, r chi.Router, baseURL string) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{
		BaseURL:    baseURL,
		BaseRouter: r,
	})
}

// HandlerWithOptions creates http.Handler with additional options
func HandlerWithOptions(si ServerInterface, options ChiServerOptions) http.Handler {
	r := options.BaseRouter

	if r == nil {
		r = chi.NewRouter()
	}
	if options.ErrorHandlerFunc == nil {
		options.ErrorHandlerFunc = func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}
	wrapper := ServerInterfaceWrapper{
		Handler:            si,
		HandlerMiddlewares: options.Middlewares,
		ErrorHandlerFunc:   options.ErrorHandlerFunc,
	}

	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/auth/{provider}/callback", wrapper.GetAuthProviderCallback)
	})
	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/auth/{provider}/login", wrapper.GetAuthProviderLogin)
	})
	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/auth/{provider}/token", wrapper.GetAuthProviderToken)
	})

	return r
}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"H4sIAAAAAAAC/7yVz2/bOgzH/xWBl3fxS9Ifh8K3vj5sCzBgRdqehqJQbSZW60gqSQfLCv/vA2U7zeIW",
	"6DBspxYyv+SX/FDKMzi/DJA/Q4lckIvigocczi/nZhnIrK23K+dX5st5I5WR8IiejfWlsY1U6MUVViUT",
	"yECc1KhajbxC2rgCzfnlHDLYIHGX+Ggym8ygzSBE9DY6yOEkHWUQrVSsVqaae/ocKWxcidROC1vX97Z4",
	"1I8rFP0TIlKqPC8hh48oWvWyV1wM8ZqU7BoFiSH/qt1CngpBBt6u1e5QBjIgfGocYQm5UIMZcFHh2mo5",
	"2UaNZSHnV9C22eHAris0LFbQ7CqaInixzuv4CEtHWIi5Wcy78Xlx/15cLT50M9X5JW9PDdL2xVxK+UvO",
	"bjWYY/CMaZjHs9kY71VTFMi8bOp6u08Sy4k2B9ys15a2kMMn68sae/wDh9QBSyDsN6JTjbjVYaVNvQ/a",
	"5xT8d4kpDQkvdKRC0zCSHtqlMhyv+WuYhgR3DbnfonUyOx7TWvTZ+Sd/+n+HZZiHSfM20a7wEONi12En",
	"HCT/cJ/jUDpimUC/l+V1Cv5jLN/Ycb1v6JM7G2PdM5s+cEjGX/JFUvfiOrVNd+Fu1+BBtQzwW3SEfOf2",
	"PzsvuELS74RLQq7ezNBmw0m4f8BCoNWjQ8bSkOc9qv270GZw+toV/s+WRieGLJlZO2Z9ZhhZ31kz/78X",
	"Ho2FN17RBnLfsTRd733w6Tg4gTQ+iFmGxpfjtRJyuEFj/b7r9NNhDUcs3NIVu22bdJ0z0mZYiYZqyKES",
	"ifl0WofC1lVgyc9mZzNob9sfAQAA//9TpKWPoAYAAA==",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %w", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

// a naive cached of a decoded swagger spec
func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

// Constructs a synthetic filesystem for resolving external references when loading openapi specifications.
func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	res := make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
	resolvePath := PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		pathToFile := url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]
		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)
			return nil, err1
		}
		return getSpec()
	}
	var specData []byte
	specData, err = rawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}
