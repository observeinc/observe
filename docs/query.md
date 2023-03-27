# query

    observe query -q 'limit 4' -i 'Default.System' -x

The query command allows you to run OPAL queries on data in your Observe
workspace. You can specify one or more input datasets, and perform joins,
unions, aggregation, and all the rest of the powerful OPAL temporal relational
algebra operations. There is no requirement for your query to be accelerable,
because the `query` command directly returns the data (like when using a
worksheet, or viewing a dashboard) and does not publish a new dataset.

## Output File Format

The output of the query command can be in one of four formats:

1. Default table format: This looks a lot like typical SQL table output, with
   vertical bars and space padding between columns, and a header row. The
   maximum width of each column is by default limited to make the output more
   readable, and excess data is cut off with an ellipsis; you can turn this
   behavior off with the `--col-width=0` option, or you can set the limit
   higher or lower than default.
2. CSV file format: The comma separated values file format is a classic in data
   processing. The output is a number of columns, optionally quoted in double
   quotes, each separated with a comma character and the row is terminated with
   a newline. Double quotes inside each quoted value are quoted by repeating
   the double quote character, and newlines inside a quoted string do not
   terminate the row. The CSV data format has a single header row containing
   the names of the columns, followed by data rows. You specify the CSV output
   format with the `--csv` command line option.
3. ND-JSON file format: Each row is output as a JSON object, terminated by a
   newline. Each column is a named key in the object, so column names are
   repeated for each row. There is no header row. You specify this format with
   the `--json` command line option.
4. Extended table format: This prints each field of each record on a line of
   its own, with a row of dashes between each record. This format is helpful if
   there are long column values that you want to look at, such as JSON objects.
   Specify this format with `--extended`.

The default table format by default quotes special characters like newlines to
avoid breaking the formatting of the table; if you want to print such
characters literally, you can use the `--literal-strings` option. CSV and JSON
data are well-defined by their specific formats, and are not affected by this
option, nor by column max length.

Extended format, specified by `--extended-format`, is like table format, except
each column gets its own row and label in the output.

## Query Text

The query OPAL can be taken from the command line (you will likely want to use
single quotes to avoid problems with shell interpolation) or from a file with
the `--file=filename` command line option. The query has the same format as the
OPAL console you will see for a stage in a worksheet.

## Query Inputs

The Inputs section of the stage is provided with the `--input=inputlist` option,
which is a comma-separated list of dataset IDs (or paths) and input names. The
first input will be given the name `_` if not specified; each additional input
needs a name. As an example:

    observe query -q 'leftjoin host_ip=@right.dst_ip, source:@right.src_ip' \
        -i '41021818,right=41012929'

You can also use the name of a workspace and dataset as a dataset path, as long
as the names don't include a comma. You specify these as:

    observe query -q `leftjoin host_ip=@right.dst_ip, source:@right.src_ip' \
        -i 'Default.aws/Instance Events,right=network/Flow Logs'

The text form of dataset names is Workspace.Folder/Name, and if you specify no
workspace in the input, the default workspace name in configuration, if any,
will be used. Thus, you can set the default workspace name per-profile to avoid
having to type it out, or specify it on the command line to easily run the same
query in multiple different workspaces, assuming multiple workspaces are
enabled for the instance you're querying.

## Query Time Window

Each query is evaluated in a particular time window. By default, this time
window is "the last hour" truncated to full minutes. You can specify time in
three different quantities, but you can specify at most two of these quantities
for any one query:

1. `--start-time=date` specifies the beginning of the time window. This can be
   a string in the ISO/RFC3339 date format like `"YYYY-MM-DDTHH:MM:SSZ"` or it
   can be another format, see below.
2. `--end-time=date` specifies the end of the time window. This can be a string
   in the ISO/RFC3339 date format like `"YYYY-MM-DDTHH:MM:SSZ"` or it
   can be another format, see below.
3. `--relative` is an interval in seconds, minutes, hours, or days. If you
   specify duration together with `--start-time=date` then the time window will
   extend forward by this amount from that date; if you specify duration
   together with `--end-time=date` then the time window will extend backward by
   this amount from that date.

If you only specify one value, the default duration is one hour, and the
default time anchor is "end-time is now, truncated to last minute."

"Now" times are read from the local machine clock. Times can be specified as
absolute using ISO UTC time, or relative using '-1h' or 'now-1h' format.

Time can also be snapped to a grid using '@5m' format. Units supported for
snapping are 's' (seconds), m (minutes), h (hours), d (days). Note that
snapping always happens before offset is applied; 'now+7h@1d' is the same as
'now@1d+7h'. Because of ambiguities, specifying an absolute date, followed by
an offset, followed by a snap is not allowed; if you specify an absolute date,
specify the snap before the offset. A snap simply of a unit indicates a snap to
one of those units -- 'd' means '1d'.

Even if times are specified in local time, the date math (snapping to days) is
performed in UTC. Times without explicit time zone will be interpreted in UTC.
The time zone offset is positive because EST is behind UTC, and thus 10:00 in
EST is 15:00 UTC, and times are calculated in UTC.

Some example times are:

* 2023-04-20T16:20:00Z (RFC 3339)
* 2023-04-20T16:20:00.123-08:00
* -1d@1h (1 day ago from local time, snapped to start of hour)
* now@10m-3h (three hours ago, snapped to start of 10-minute period)
* 1682007600@1d (epoch seconds: UNIX, snapped to start of UTC day)
* 1682007600000 (epoch milliseconds: Java and Javascript)
* 1682007600000000000 (epoch nanoseconds: Go, C++, OPAL)

## Example

    observe query \
        --input '41021818,right=41012929' \
        --relative '4h' \
        --end-time '@1h' \
        --col-width 0 \
        --extended \
        --query "leftjoin host_ip=@right.dst_ip, source:@right.src_ip, name:@right.ip_name | filter name ~ /^ru-/ | limit 4"

## Example

    observe query -i 41007104,pods=41007085 -x -q 'join podName=@pods.name, @pods.restartCount | limit 1'

