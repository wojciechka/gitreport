package gitpg

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	version1String = "Q1"
	PrefixLength   = 2
	AllPrefix      = "AL"
	AuthorPrefix   = "AU"
	RefPrefix      = "ID"
	FromTimePrefix = "FT"
	ToTimePrefix   = "TT"
	querySeparator = "\n"
)

type LogQuery struct {
	AuthorRegexp string
	All          bool
	Ref          string
	FromTime     *time.Time
	ToTime       *time.Time
}

func NormalizeLogQuery(query *LogQuery) (*LogQuery, error) {
	str, err := ExportLogQuery(query)
	if err != nil {
		return nil, err
	}

	return ImportLogQuery(str)
}

func ExportLogQueryFilename(query *LogQuery) (string, error) {
	exportString, err := ExportLogQuery(query)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString([]byte(exportString)), nil
}

func ImportLogQueryFilename(filename string) (*LogQuery, error) {
	data, err := base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(filename)
	if err != nil {
		return nil, err
	}

	return ImportLogQuery(string(data))
}

func ExportLogQuery(query *LogQuery) (string, error) {
	params := make([]string, 0)

	params = append(params, version1String)

	if query.All {
		params = append(params, AllPrefix)
	} else if len(query.Ref) > 0 {
		params = append(params, RefPrefix+query.Ref)
	}

	if len(query.AuthorRegexp) > 0 {
		params = append(params, AuthorPrefix+query.AuthorRegexp)
	}

	if query.FromTime != nil {
		params = append(params, FromTimePrefix+formatTime(*query.FromTime))
	}
	if query.ToTime != nil {
		params = append(params, ToTimePrefix+formatTime(*query.ToTime))
	}
	return strings.Join(params, querySeparator), nil
}

func ImportLogQuery(q string) (*LogQuery, error) {
	query := strings.Split(q, querySeparator)

	if len(query) < 1 {
		// TODO: error handling
		return nil, errors.New("Invalid Log Query")
	}

	if query[0] == version1String {
		return importLogVersion1Query(query[1:])
	} else {
		// TODO: error handling
		return nil, errors.New("Invalid Log Query")
	}
}

func formatTime(t time.Time) string {
	return fmt.Sprintf("0x%x", t.UTC().Unix())
}

func parseTime(value string) (*time.Time, error) {
	var unixTime int64
	n, err := fmt.Sscanf(value, "0x%x", &unixTime)
	if err != nil {
		return nil, err
	}
	if n < 1 {
		return nil, errors.New(fmt.Sprintf("Invalid time export format %s", value))
	}

	t := time.Unix(unixTime, 0)
	return &t, nil
}

func importLogVersion1Query(query []string) (*LogQuery, error) {
	var err error
	q := LogQuery{}
	for _, param := range query {
		if len(param) < PrefixLength {
			return nil, errors.New("Invalid Log Parameter " + param)
		}

		paramName := param[0:PrefixLength]
		paramValue := param[PrefixLength:]

		switch paramName {
		case AuthorPrefix:
			q.AuthorRegexp = paramValue
		case RefPrefix:
			q.Ref = paramValue
		case AllPrefix:
			q.All = true
		case FromTimePrefix:
			q.FromTime, err = parseTime(paramValue)
			if err != nil {
				return nil, err
			}
		case ToTimePrefix:
			q.ToTime, err = parseTime(paramValue)
			if err != nil {
				return nil, err
			}
		default:
			return nil, errors.New("Invalid Log Parameter value " + param)
		}
	}

	return &q, nil
}
