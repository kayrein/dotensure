# dotensure

Pipe the results of `go test` or `encore test` through this tool to check whether [dotsql](https://github.com/qustavo/dotsql)
queries are being executed by the unit tests.

You must specify -json as one of the arguments to the test program, so that it outputs JSON that this tool can read
(you may send it whatever other arguments you like).

## Usage
`encore test -json ./... | dotensure`

You will also need to add log lines in your tests printing a line of the form

`ExpectedQuery: query-name\n`

Each time a query is loaded (dotsql includes a convenient `QueryMap` function which suits this)

And also

`ExecutedQuery: query-name\n`

When it is run. Or just first run, that's sufficient. I suggest a wrapper around your `*dotsql.DotSql` instance inside
the unit tests.

## Tags

Optionally, you can include a tag at the end of each query:
`ExpectedQuery: query-name; tag\n`

This tag is not used in matching executing and expected queries but does display during output. 

If multiple ExpectedQueries have the same name but different tags, behaviour is undefined (at the moment it will take
the first one, but I suggest not relying on this).

It will also count the number of missing queries for each tag value.

Tags are also syntactically accepted for ExecutedQueries but currently do not do anything.

## Installation
`go install github.com/kayrein/dotensure`

## Arguments

Pass `-verbose` to dotensure to print information about which test executed each test query.

Pass `-version` to dotensure to print the version.
