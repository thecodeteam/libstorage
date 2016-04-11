package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	//log "github.com/Sirupsen/logrus"

	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/types/context"
	apihttp "github.com/emccode/libstorage/api/types/http"
	"github.com/emccode/libstorage/api/utils/schema"
)

// schemaValidator is an HTTP filter for validating incoming request payloads
type schemaValidator struct {
	handler       apihttp.APIFunc
	reqSchema     []byte
	resSchema     []byte
	newReqObjFunc func() interface{}
}

// NewSchemaValidator returns a new filter for validating request payloads and
// response payloads against defined JSON schemas.
func NewSchemaValidator(
	reqSchema, resSchema []byte,
	newReqObjFunc func() interface{}) apihttp.Middleware {

	return &schemaValidator{
		reqSchema:     reqSchema,
		resSchema:     resSchema,
		newReqObjFunc: newReqObjFunc,
	}
}

func (h *schemaValidator) Name() string {
	return "schmea-validator"
}

func (h *schemaValidator) Handler(m apihttp.APIFunc) apihttp.APIFunc {
	return (&schemaValidator{
		m, h.reqSchema, h.resSchema, h.newReqObjFunc}).Handle
}

// Handle is the type's Handler function.
func (h *schemaValidator) Handle(
	ctx context.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	reqBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return fmt.Errorf("validate req schema: read req error: %v", err)
	}

	// do the request validation
	if h.reqSchema != nil {
		err = schema.Validate(ctx, h.reqSchema, reqBody)
		if err != nil {
			return fmt.Errorf("validate req schema: validation error: %v", err)
		}
	}

	// create the object for the request payload if there is a function for it
	if h.newReqObjFunc != nil {
		reqObj := h.newReqObjFunc()
		if len(reqBody) > 0 {
			if err = json.Unmarshal(reqBody, reqObj); err != nil {
				return fmt.Errorf(
					"validate req schema: unmarshal error: %v", err)
			}
		}
		ctx = context.WithValue(ctx, "reqObj", reqObj)
	}

	// if there's not response schema then just return the result of the next
	// handler
	if h.resSchema == nil {
		return h.handler(ctx, w, req, store)
	}

	// at this point we know there's going to be response validation, so
	// we need to record the result of the next handler in order to intercept
	// the response payload to validate it
	rec := httptest.NewRecorder()

	// invoke the next handler with a recorder
	err = h.handler(ctx, rec, req, store)
	if err != nil {
		return err
	}

	// do the response validation
	resBody := rec.Body.Bytes()
	err = schema.Validate(ctx, h.resSchema, resBody)
	if err != nil {
		return err
	}

	// write the recorded result of the next handler to the resposne writer
	w.WriteHeader(rec.Code)
	for k, v := range rec.HeaderMap {
		w.Header()[k] = v
	}
	if _, err = w.Write(resBody); err != nil {
		return err
	}

	return nil
}