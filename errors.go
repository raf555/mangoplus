package mangoplus

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/raf555/mangoplus/internal/proto"
	"github.com/raf555/mangoplus/internal/slicex"
)

// APIError is an error returned when MangaPlus API returns non-200 status code.
type APIError struct {
	StatusCode int
	Headers    http.Header
	RawBody    []byte
}

func (a *APIError) Error() string {
	return fmt.Sprintf("mangoplus: API returned non-200 status code: %d", a.StatusCode)
}

// ProtoError is an error returned when MangaPlus API returns an error response.
type ProtoError struct {
	Action       ErrorAction
	EnglishPopup PopupOsDefault
	DebugInfo    string
	Popups       []PopupOsDefault
}

func (p *ProtoError) Error() string {
	var b strings.Builder
	b.WriteString("mangoplus: proto error")
	for _, part := range []string{string(p.Action), p.EnglishPopup.Subject, p.EnglishPopup.Body} {
		if part != "" {
			b.WriteString(": ")
			b.WriteString(part)
		}
	}
	return b.String()
}

var (
	// ErrNotFound is an error when MangaPlus returned proto error with Not Found subject.
	ErrNotFound = errors.New("mangoplus: not found")

	// ErrInvalidParameter is an error when MangaPlus returned proto error with Invalid Parameter subject.
	ErrInvalidParameter = errors.New("mangoplus: invalid parameter")

	// ErrNewVersionAvailable is an error when MangaPlus returned proto error with NEW VERSION AVAILABLE subject.
	ErrNewVersionAvailable = errors.New("mangoplus: new version available")

	// ErrInvalidUserAccess is an error when MangaPlus returned proto error with Invalid user subject.
	ErrInvalidUserAccess = errors.New("mangoplus: invalid user access")

	// ErrMaintenance is an error when MangaPlus returned proto error with MAINTENANCE action.
	ErrMaintenance = errors.New("mangoplus: maintenance")
)

func (p *ProtoError) Is(target error) bool {
	switch target {
	case ErrNotFound:
		return p.EnglishPopup.Subject == "Not Found"
	case ErrInvalidParameter:
		return p.EnglishPopup.Subject == "Invalid Parameter"
	case ErrNewVersionAvailable:
		return p.EnglishPopup.Subject == "NEW VERSION AVAILABLE"
	case ErrInvalidUserAccess:
		return p.EnglishPopup.Subject == "Invalid user"
	case ErrMaintenance:
		return p.Action == ErrorActionMaintenance
	}
	return false
}

func protoErrorFromProto(pb *proto.ErrorResult) *ProtoError {
	return &ProtoError{
		Action:       errorActionFromProto(pb.GetAction()),
		EnglishPopup: popupOsDefaultFromProto(pb.GetEnglishPopup()),
		DebugInfo:    pb.GetDebugInfo(),
		Popups:       slicex.Map(pb.GetPopups(), popupOsDefaultFromProto),
	}
}

type ErrorAction string

const (
	ErrorActionDefault       ErrorAction = "DEFAULT"
	ErrorActionUnauthorized  ErrorAction = "UNAUTHORIZED"
	ErrorActionMaintenance   ErrorAction = "MAINTENANCE"
	ErrorActionGeoIPBlocking ErrorAction = "GEOIP_BLOCKING"
)

func errorActionFromProto(pb proto.ErrorResult_Action) ErrorAction {
	val, ok := proto.ErrorResult_Action_name[int32(pb)]
	if !ok {
		return ""
	}
	return ErrorAction(val)
}
