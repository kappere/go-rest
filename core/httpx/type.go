package httpx

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"time"
)

// 常用业务状态码
const (
	STATUS_SUCCESS           = 0
	STATUS_ERROR_COMMON      = -1
	STATUS_ERROR_LIMIT       = -899
	STATUS_NO_AUTHENTICATION = -999
	STATUS_NO_AUTHORIZATION  = -989
)

type Response interface {
	Create(data interface{}, code int, message string) Response
	GetData() interface{}
	Error() error
}

type DefaultResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func (r *DefaultResponse) Create(data interface{}, code int, message string) Response {
	return &DefaultResponse{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

func (r *DefaultResponse) GetData() interface{} {
	return r.Data
}

func (r *DefaultResponse) Error() error {
	if r.Code == 0 {
		return nil
	}
	return errors.New(r.Message)
}

func SetResponse(response Response) {
	responseGenerator = response
}

var responseGenerator Response = &DefaultResponse{}

func Ok(data interface{}) Response {
	return responseGenerator.Create(data, STATUS_SUCCESS, "")
}

func Error(msg string) Response {
	return responseGenerator.Create(nil, STATUS_ERROR_COMMON, msg)
}

func ErrorWithCode(msg string, code int) Response {
	return responseGenerator.Create(nil, code, msg)
}

type Time time.Time
type Date time.Time

const (
	TIME_FORMAT = "2006-01-02 15:04:05"
	DATE_FORMAT = "2006-01-02"
)

func (t *Time) UnmarshalJSON(data []byte) (err error) {
	now, err := time.ParseInLocation(`"`+TIME_FORMAT+`"`, string(data), time.Local)
	*t = Time(now)
	return
}

func (t Time) MarshalJSON() ([]byte, error) {
	b := make([]byte, 0, len(TIME_FORMAT)+2)
	b = append(b, '"')
	b = time.Time(t).AppendFormat(b, TIME_FORMAT)
	b = append(b, '"')
	return b, nil
}

func (t Time) String() string {
	return time.Time(t).Format(TIME_FORMAT)
}

func (t *Time) Scan(src interface{}) error {
	switch src := src.(type) {
	case nil:
		return nil
	case string:
		if src == "" {
			return nil
		}
		tm, err := time.ParseInLocation(TIME_FORMAT, src, time.Local)
		if err != nil {
			return fmt.Errorf("Scan: %v", err)
		}
		*t = Time(tm)
		return nil
	default:
		return fmt.Errorf("Scan: unable to scan type %T into rest.Time", src)
	}
}

func (t Time) Value() (driver.Value, error) {
	return time.Time(t).Format(TIME_FORMAT), nil
}

func (t *Date) UnmarshalJSON(data []byte) (err error) {
	now, err := time.ParseInLocation(`"`+DATE_FORMAT+`"`, string(data), time.Local)
	*t = Date(now)
	return
}

func (t Date) MarshalJSON() ([]byte, error) {
	b := make([]byte, 0, len(DATE_FORMAT)+2)
	b = append(b, '"')
	b = time.Time(t).AppendFormat(b, DATE_FORMAT)
	b = append(b, '"')
	return b, nil
}

func (t Date) String() string {
	return time.Time(t).Format(DATE_FORMAT)
}

func (t *Date) Scan(src interface{}) error {
	switch src := src.(type) {
	case nil:
		return nil
	case string:
		if src == "" {
			return nil
		}
		tm, err := time.ParseInLocation(DATE_FORMAT, src, time.Local)
		if err != nil {
			return fmt.Errorf("Scan: %v", err)
		}
		*t = Date(tm)
		return nil
	default:
		return fmt.Errorf("Scan: unable to scan type %T into rest.Date", src)
	}
}

func (t Date) Value() (driver.Value, error) {
	return time.Time(t).Format(DATE_FORMAT), nil
}

func FormatTime(t time.Time, format string) string {
	return time.Time(t).Format(format)
}

func GetTime(t string, format string) time.Time {
	tm, err := time.ParseInLocation(format, t, time.Local)
	if err != nil {
		panic(err)
	}
	return time.Time(tm)
}
