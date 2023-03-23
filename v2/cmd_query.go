package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/pflag"
)

var flagText string
var flagFile string
var flagInput []string
var flagJSON bool
var flagCSV bool
var flagValidFrom string
var flagValidTo string
var flagDuration time.Duration
var flagColWidth int
var flagExtendedFormat bool
var flagUnquoteStrings bool

var flagsQuery *pflag.FlagSet

func init() {
	flagsQuery = pflag.NewFlagSet("query", pflag.ContinueOnError)
	flagsQuery.StringVarP(&flagText, "query", "Q", "", "OPAL query text")
	flagsQuery.StringVarP(&flagFile, "file", "F", "", "file containing OPAL query text")
	flagsQuery.StringSliceVarP(&flagInput, "input", "I", nil, "input datasets: ID or workspace:name")
	flagsQuery.BoolVarP(&flagJSON, "json", "J", false, "output in nd-JSON format")
	flagsQuery.Lookup("json").NoOptDefVal = "true"
	flagsQuery.BoolVarP(&flagCSV, "csv", "C", false, "output in CSV format")
	flagsQuery.Lookup("csv").NoOptDefVal = "true"
	flagsQuery.StringVarP(&flagValidFrom, "valid-from", "B", "", "beginnig time of query window")
	flagsQuery.StringVarP(&flagValidTo, "valid-to", "E", "", "end time of query window")
	flagsQuery.DurationVarP(&flagDuration, "duration", "D", 0, "duration of query window")
	flagsQuery.IntVarP(&flagColWidth, "col-width", "W", 64, "maximum column width for table format; 0 for unlimited")
	flagsQuery.BoolVarP(&flagExtendedFormat, "extended-format", "X", false, "print one column value per row rather than table format")
	flagsQuery.Lookup("extended-format").NoOptDefVal = "true"
	flagsQuery.BoolVarP(&flagUnquoteStrings, "unquote-strings", "U", false, "print embedded control characters literally")
	flagsQuery.Lookup("unquote-strings").NoOptDefVal = "true"
	RegisterCommand(&Command{
		Name:  "query",
		Help:  "Run an OPAL query. Provide the query either on command line, or in file. Provide the query time window using some combination of start time, end time, and duration, with defaults being 1 hour, leading up to 'now'. Local times are read from the local machine clock.",
		Flags: flagsQuery,
		Func:  cmdQuery,
	})
}

const DefaultQueryWindowDuration = time.Hour
const MaxQueryTextLength = 100000

var ErrNeedQueryOrFile = ObserveError{Msg: "need one of --query and --file for the query text"}
var ErrOnlyOneQueryOrFile = ObserveError{Msg: "only one of --query and --file may be specified"}
var ErrAtMostTwoTimeSpecifiers = ObserveError{Msg: "at most two of --valid-from, --valid-to, and --duration may be specified"}
var ErrValidToMustBeAfterValidFrom = ObserveError{Msg: "--valid-to time must be after --vaild-from time"}
var ErrTooLongQueryText = ObserveError{Msg: fmt.Sprintf("the query text is longer than %d characters", MaxQueryTextLength)}
var ErrAtMostOneOutputFormat = ObserveError{Msg: "at most one of --csv and --json may be specified"}
var ErrAnInputIsRequired = ObserveError{Msg: "at least one --input is required"}

func cmdQuery(cfg *Config, op Output, args []string, hc *http.Client) error {

	// parse query arguments

	nText := CountFlags(flagsQuery, "query", "file")
	var queryText string
	var err error
	switch nText {
	default:
		return ErrOnlyOneQueryOrFile
	case 0:
		return ErrNeedQueryOrFile
	case 1:
		// I have to grudgingly accept this
		if flagsQuery.Lookup("file").Changed {
			queryText, err = LoadQueryTextFromFile(flagFile)
		} else {
			queryText = flagText
		}
	}
	if err != nil {
		return err
	}
	if len(queryText) > MaxQueryTextLength {
		return ErrTooLongQueryText
	}

	var fromTime, toTime time.Time
	nTime := CountFlags(flagsQuery, "valid-from", "valid-to", "duration")
	switch nTime {
	case 0:
		toTime = time.Now().Add(-15 * time.Second).Truncate(time.Minute)
		fromTime = toTime.Add(-DefaultQueryWindowDuration)
	case 1:
		if flagsQuery.Lookup("valid-from").Changed {
			fromTime, err = ParseTime(flagValidFrom)
			toTime = fromTime.Add(DefaultQueryWindowDuration)
		} else if flagsQuery.Lookup("valid-to").Changed {
			toTime, err = ParseTime(flagValidTo)
			fromTime = toTime.Add(-DefaultQueryWindowDuration)
		} else {
			toTime = time.Now().Add(-15 * time.Second).Truncate(time.Minute)
			fromTime = toTime.Add(-flagDuration)
		}
	case 2:
		if !flagsQuery.Lookup("valid-from").Changed {
			toTime, err = ParseTime(flagValidTo)
			fromTime = toTime.Add(-flagDuration)
		} else if !flagsQuery.Lookup("valid-to").Changed {
			fromTime, err = ParseTime(flagValidFrom)
			toTime = fromTime.Add(flagDuration)
		} else {
			fromTime, err = ParseTime(flagValidFrom)
			if err == nil {
				toTime, err = ParseTime(flagValidTo)
			}
		}
	default:
		return ErrAtMostTwoTimeSpecifiers
	}
	if err != nil {
		return NewObserveError(err, "bad time format")
	}
	if toTime.Sub(fromTime) <= 0 {
		return ErrValidToMustBeAfterValidFrom
	}

	nFmt := CountFlags(flagsQuery, "csv", "json")
	switch nFmt {
	case 0:
		// format is "table text"
	case 1:
		// format is specified
	default:
		return ErrAtMostOneOutputFormat
	}

	// TODO: we can remove this when in-text inputs are complete
	if len(flagInput) == 0 {
		return ErrAnInputIsRequired
	}
	var inputs []StageQueryInput
	for i, in := range flagInput {
		pieces := strings.SplitN(in, "=", 2)
		if len(pieces) == 1 {
			if i == 0 {
				pieces = append([]string{"_"}, pieces[0])
			} else {
				return NewObserveError(nil, "input at index %d must be of the form id=dataset", i)
			}
		}
		for j, k := range inputs {
			if k.InputName == pieces[0] {
				return NewObserveError(nil, "input at index %d duplicates input name %q from index %d", j, k.InputName, i)
			}
		}
		var i64 int64
		if i64, err = strconv.ParseInt(pieces[1], 10, 64); err == nil {
			inputs = append(inputs, StageQueryInput{
				InputName: pieces[0],
				DatasetID: &i64,
			})
			op.Debug("input[%d] @%s <- datasetId(%d)\n", i, pieces[0], i64)
		} else {
			inputs = append(inputs, StageQueryInput{
				InputName:   pieces[0],
				DatasetPath: &pieces[1],
			})
			op.Debug("input[%d] @%s <- datasetPath(%q)\n", i, pieces[0], pieces[1])
		}
	}

	// I'm now ready to formulate the query
	req := V1ExportQueryRequest{
		Query: Query{
			OutputStage: "query",
			Stages: []StageQuery{
				{
					Inputs:   inputs,
					StageID:  "query",
					Pipeline: queryText,
				},
			},
		},
	}

	var output io.Writer
	acceptHeader := "text/csv"
	switch {
	case flagJSON:
		op.Debug("format=JSON\n")
		output = op
		acceptHeader = "application/x-ndjson"
	case flagCSV:
		op.Debug("format=CSV\n")
		output = op
	default:
		op.Debug("format=table\n")
		// text format
		tfmt := &ColumnFormatter{Output: op, ColWidth: flagColWidth, ExtendedFormat: flagExtendedFormat, UnquoteStrings: flagUnquoteStrings}
		defer tfmt.Close()
		output = tfmt
	}

	uri := fmt.Sprintf("/v1/meta/export/query?startTime=%s&endTime=%s", fromTime.Format(time.RFC3339), toTime.Format(time.RFC3339))
	err, _ = RequestPOSTWithBodyOutput(cfg, op, hc, uri, &req, map[string]string{
		"Accept":        acceptHeader,
		"Authorization": cfg.AuthHeader(),
	}, output)
	if err != nil {
		return err
	}
	return nil
}

type Query struct {
	OutputStage string `json:"outputStage"`
	// The API wants this marshaled as an array, but we only do one stage,
	// because inline stages are now a thing.
	Stages []StageQuery `json:"stages"`
	// no parameters
	// no parameterValues
	// no layout
}

type StageQuery struct {
	// The API wants this, but once inline inputs are complete, we may be able
	// to remove them. Might want to support them still, though.
	Inputs   []StageQueryInput `json:"input"`
	StageID  string            `json:"stageID"`
	Pipeline string            `json:"pipeline"`
	// no layout
	// no parameters
	// no parameterValues
	// no stageIndex
}

type StageQueryInput struct {
	InputName   string  `json:"inputName"`
	DatasetID   *int64  `json:"datasetId,string,omitempty"`
	DatasetPath *string `json:"datasetPath,omitempty"`
}

type V1ExportQueryRequest struct {
	Query Query `json:"query"`
	// no rowCount
	// no presentation
}
