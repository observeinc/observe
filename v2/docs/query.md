# query

The query command allows you to run OPAL queries on data in your Observe
workspace. You can specify one or more input datasets, and perform joins,
unions, aggregation, and all the rest of the powerful OPAL temporal relational
algebra operations. There is no requirement for your query to be accelerable,
because the `query` command directly returns the data (like when using a
worksheet, or viewing a dashboard) and does not publish a new dataset.

The output of the query command can be in one of three formats:

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

The default table format by default quotes special characters like newlines to
avoid breaking the formatting of the table; if you want to print such
characters literally, you can use the `--unquote-strings` option. CSV and JSON
data are well-defined by their specific formats, and are not affected by this
option, nor by column max length.

Extended format, specified by `--extended-format`, is like table format, except
each column gets its own row and label in the output.

The query OPAL can be taken from the command line (you will likely want to use
single quotes to avoid problems with shell interpolation) or from a file with
the `--file=filename` command line option. The query has the same format as the
OPAL console you will see for a stage in a worksheet.

The Inputs section of the stage is provided with the `--input=inputlist` option,
which is a comma-separated list of dataset IDs (or paths) and input names. The
first input will be given the name `_` if not specified; each additional input
needs a name. As an example:

    observe query -Q 'leftjoin host_ip=@right.dst_ip, source:@right.src_ip' \
        -I '41021818,right=41012929'

You can also use the name of a workspace and dataset as a dataset path, as long
as the names don't include a comma. You specify these as:

    observe query -Q `leftjoin host_ip=@right.dst_ip, source:@right.src_ip' \
        -I 'Default.aws/Instance Events,right=network/Flow Logs'

Each query is evaluated in a particular time window. By default, this time
window is "the last hour" truncated to full minutes. You can specify time in
three different quantities, but you can specify at most two of these quantities
for any one query:

1. `--valid-from=date` specifies the beginning of the time window. This should
   be a string in the ISO/RFC3339 date format like `"YYYY-MM-DDTHH:MM:SSZ"` or
   it can be a number of seconds, milliseconds, or nanoseconds since the UNIX
   epoch. Which of the units is intended is determined from the magnitude of
   the number when using this form.
2. `--valid-to=date` specifies the end of the time window. This should be a
   string in the ISO/RFC3339 date format like `"YYYY-MM-DDTHH:MM:SSZ"` or it
   can be a number of seconds, milliseconds, or nanoseconds since the UNIX
   epoch. Which of the units is intended is determined from the magnitude of
   the number when using this form.
3. `--duration` is an interval in seconds, minutes, hours, or days. If you
   specify duration together with `--valid-from=date` then the time window will
   extend forward by this amount from that date; if you specify duration
   together with `--valid-to=date` then the time window will extend backward by
   this amount from that date.

If you only specify one value, the default duration is one hour, and the
default time anchor is "valid to is now truncated to last minute."
