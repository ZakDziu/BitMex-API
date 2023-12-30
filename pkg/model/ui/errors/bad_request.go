package errors

type UIResponseErrorBadRequest struct {
	Code    int    `example:"400"                  json:"code"`
	Message string `example:"request invalid body" json:"message"`
}
