package main

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/pflag"
)

var (
	flagsQuery              *pflag.FlagSet
	flagQueryText           string
	flagQueryFile           string
	flagQueryInputs         []string
	flagQueryJSON           bool
	flagQueryCSV            bool
	flagQueryStartTime      string
	flagQueryEndTime        string
	flagQueryRelative       time.Duration
	flagQueryColWidth       int
	flagQueryExtended       bool
	flagQueryLiteralStrings bool
	flagQueryFormat         string
)

func init() {
	flagsQuery = pflag.NewFlagSet("query", pflag.ContinueOnError)
	flagsQuery.StringVarP(&flagQueryText, "query", "q", "", "OPAL query text")
	flagsQuery.StringVarP(&flagQueryFile, "file", "f", "", "file containing OPAL query text")
	flagsQuery.StringSliceVarP(&flagQueryInputs, "input", "i", nil, "input datasets: ID or workspace.name")
	flagsQuery.BoolVarP(&flagQueryJSON, "json", "j", false, "output in nd-JSON format")
	flagsQuery.Lookup("json").NoOptDefVal = "true"
	flagsQuery.BoolVarP(&flagQueryCSV, "csv", "c", false, "output in CSV format")
	flagsQuery.Lookup("csv").NoOptDefVal = "true"
	flagsQuery.StringVarP(&flagQueryStartTime, "start-time", "s", "", "start time of query window")
	flagsQuery.StringVarP(&flagQueryEndTime, "end-time", "e", "", "end time of query window")
	flagsQuery.DurationVarP(&flagQueryRelative, "relative", "r", 0, "duration of query window, anchored at either end")
	flagsQuery.IntVarP(&flagQueryColWidth, "col-width", "w", 64, "maximum column width for table format; 0 for unlimited")
	flagsQuery.BoolVarP(&flagQueryExtended, "extended", "x", false, "print one column value per row rather than table format")
	flagsQuery.Lookup("extended").NoOptDefVal = "true"
	flagsQuery.BoolVarP(&flagQueryLiteralStrings, "literal-strings", "l", false, "print embedded control characters literally")
	flagsQuery.Lookup("literal-strings").NoOptDefVal = "true"
	flagsQuery.StringVar(&flagQueryFormat, "format", "", "specify output format: table, extended, csv, ndjson")
	RegisterCommand(&Command{
		Name:  "query",
		Help:  "Run an OPAL query.",
		Flags: flagsQuery,
		Func:  cmdQuery,
	})
}

const DefaultQueryWindowDuration = time.Hour
const MaxQueryTextLength = 100000

var ErrNeedQueryOrFile = ObserveError{Msg: "need one of --query and --file for the query text"}
var ErrOnlyOneQueryOrFile = ObserveError{Msg: "only one of --query and --file may be specified"}
var ErrAtMostTwoTimeSpecifiers = ObserveError{Msg: "at most two of --start-time, --end-time, and --relative may be specified"}
var ErrValidToMustBeAfterValidFrom = ObserveError{Msg: "--end-time time must be after --vaild-from time"}
var ErrTooLongQueryText = ObserveError{Msg: fmt.Sprintf("the query text is longer than %d characters", MaxQueryTextLength)}
var ErrAtMostOneOutputFormat = ObserveError{Msg: "at most one of --csv and --json may be specified"}
var ErrAnInputIsRequired = ObserveError{Msg: "at least one --input is required"}
var ErrQueryTooManyFormats = ObserveError{Msg: "at most one of --format and the format-specific flags may be specified"}
var ErrUnknownFormat = ObserveError{Msg: "the --format is not known"}

func cmdQuery(fa FuncArgs) error {
	nowTime := time.Now().Truncate(time.Second)
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
			queryText, err = LoadQueryTextFromFile(fa.fs, flagQueryFile)
		} else {
			queryText = flagQueryText
		}
	}
	if err != nil {
		return err
	}
	if len(queryText) > MaxQueryTextLength {
		return ErrTooLongQueryText
	}

	var fromTime, toTime time.Time
	nTime := CountFlags(flagsQuery, "start-time", "end-time", "relative")
	switch nTime {
	case 0:
		toTime = nowTime.Add(-15 * time.Second).Truncate(time.Minute)
		fromTime = toTime.Add(-DefaultQueryWindowDuration)
	case 1:
		if flagsQuery.Lookup("start-time").Changed {
			fromTime, err = ParseTime(flagQueryStartTime, nowTime)
			toTime = fromTime.Add(DefaultQueryWindowDuration)
		} else if flagsQuery.Lookup("end-time").Changed {
			toTime, err = ParseTime(flagQueryEndTime, nowTime)
			fromTime = toTime.Add(-DefaultQueryWindowDuration)
		} else {
			toTime = nowTime.Add(-15 * time.Second).Truncate(time.Minute)
			fromTime = toTime.Add(-flagQueryRelative)
		}
	case 2:
		if !flagsQuery.Lookup("start-time").Changed {
			toTime, err = ParseTime(flagQueryEndTime, nowTime)
			fromTime = toTime.Add(-flagQueryRelative)
		} else if !flagsQuery.Lookup("end-time").Changed {
			fromTime, err = ParseTime(flagQueryStartTime, nowTime)
			toTime = fromTime.Add(flagQueryRelative)
		} else {
			fromTime, err = ParseTime(flagQueryStartTime, nowTime)
			if err == nil {
				toTime, err = ParseTime(flagQueryEndTime, nowTime)
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
	if len(flagQueryInputs) == 0 {
		return ErrAnInputIsRequired
	}
	var inputs []StageQueryInput
	for i, in := range flagQueryInputs {
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
			fa.op.Debug("input[%d] @%s <- datasetId(%d)\n", i, pieces[0], i64)
		} else {
			if !strings.Contains(pieces[1], ".") {
				workspaceName := mustGetWorkspaceName(fa.cfg, fa.hc)
				fa.op.Debug("default workspace=%s\n", workspaceName)
				pieces[1] = workspaceName + "." + pieces[1]
			}
			inputs = append(inputs, StageQueryInput{
				InputName:   pieces[0],
				DatasetPath: &pieces[1],
			})
			fa.op.Debug("input[%d] @%s <- datasetPath(%q)\n", i, pieces[0], pieces[1])
		}
	}

	// I'm now ready to formulate the query
	noLinkify := false
	req := V1ExportQueryRequest{
		Query: OpalQuery{
			OutputStage: "query",
			Stages: []StageQuery{
				{
					Inputs:   inputs,
					StageID:  "query",
					Pipeline: queryText,
				},
			},
		},
		Presentation: &Presentation{
			Linkify: &noLinkify,
		},
	}

	nfmts := 0
	if flagQueryFormat != "" {
		fa.op.Debug("formatFlag=%s\n", flagQueryFormat)
		nfmts++
	}
	if flagQueryJSON {
		fa.op.Debug("format=JSON\n")
		nfmts++
	}
	if flagQueryCSV {
		fa.op.Debug("format=CSV\n")
		nfmts++
	}
	if flagQueryExtended {
		fa.op.Debug("format=Extended\n")
		nfmts++
	}
	if nfmts > 1 {
		fa.op.Debug("too many formats: %d\n", nfmts)
		return ErrQueryTooManyFormats
	}
	// I know that if flagQueryFormat is non-empty, none of the format flags are specified
	switch flagQueryFormat {
	case "json", "JSON", "ndjson", "NDJSON", "nd-json", "ND-JSON":
		flagQueryJSON = true
	case "csv", "CSV":
		flagQueryCSV = true
	case "extended":
		flagQueryExtended = true
	case "":
		// don't change what's configured
	default:
		return ErrUnknownFormat
	}
	var output io.Writer
	acceptHeader := "text/csv"
	switch {
	case flagQueryJSON:
		output = fa.op
		acceptHeader = "application/x-ndjson"
	case flagQueryCSV:
		output = fa.op
	default:
		// text format
		tfmt := &CSVParsingColumnFormatter{ColumnFormatter: ColumnFormatter{Output: fa.op, ColWidth: flagQueryColWidth, ExtendedFormat: flagQueryExtended, LiteralStrings: flagQueryLiteralStrings}}
		defer tfmt.Close()
		output = tfmt
	}

	uri := fmt.Sprintf("/v1/meta/export/query?startTime=%s&endTime=%s", fromTime.Format(time.RFC3339), toTime.Format(time.RFC3339))
	err, _ = RequestPOSTWithBodyOutput(fa.cfg, fa.op, fa.hc, uri, &req, headers("Accept", acceptHeader, "Authorization", fa.cfg.AuthHeader()), output)
	if err != nil {
		return err
	}
	return nil
}

type OpalQuery struct {
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

type Presentation struct {
	Limit   *int64 `json:"limit,omitempty"`
	Linkify *bool  `json:"linkify,omitempty"`
}

type V1ExportQueryRequest struct {
	Query OpalQuery `json:"query"`
	// no rowCount
	Presentation *Presentation `json:"presentation,omitempty"`
}
