# Timelog

The command processes timelog input from stdin and outputs data in the following formats:
* Timelog output format: our own per-day tree-like event aggregation
* Timeclock (hledger) format: for subsequent processing with hledger (https://hledger.org/)

## Installation

```sh
# Install by building from source
➜ go install github.com/denis-engcom/timelog@latest
```

In typical usage, file contents are redirected into the program (stdin) and output from stdout can be redirected to another file.

```sh
# timelog [global options] (stdin)
timelog < 2023-01_timelog.md > 2023-01_timelog_aggregated.md
timelog -O timeclock < 2023-01_timelog.md > 2023-01_timelog.timeclock
timelog -O timeclock < 2023-01_timelog.md | hledger -ftimeclock:- register --daily @/tmp/filter.args > 2023-01_timelog_reports_register.txt
```

## Example usage

```sh
➜ timelog <<EOF
# 2022-12-15

9:30 Improving the thingy - Changing the doodad
10:00 Daily Meeting
10:15 Improving the thingy - Putting the doodad in a different spot
11:00 Pair Programming
12:00 Lunch
12:30 Improving the thingy - Code Review
14:00 Break - Phooey! That's a lot of work!
14:30 Improving the thingy - Deployment
17:30 EOD
EOF
```
```
# 2022-12-15

- 8h
	- 5h45m: Improving the thingy
		- 3h: Deployment
		- 1h30m: Code Review
		- 45m: Putting the doodad in a different spot
		- 30m: Changing the doodad
	- 1h: Pair Programming
	- 30m: Break
		- 30m: Phooey! That's a lot of work!
	- 30m: Lunch
	- 15m: Daily Meeting

```
```sh
➜ timelog --output-format timeclock <<EOF
# 2022-12-15

9:30 Improving the thingy - Changing the doodad
10:00 Daily Meeting
10:15 Improving the thingy - Putting the doodad in a different spot
11:00 Pair Programming
12:00 Lunch
12:30 Improving the thingy - Code Review
14:00 Break - Phooey! That's a lot of work!
14:30 Improving the thingy - Deployment
17:30 EOD
EOF
```
```
i 2022-12-15 09:30 Improving the thingy:Changing the doodad
o 2022-12-15 10:00
i 2022-12-15 10:00 Daily Meeting
o 2022-12-15 10:15
i 2022-12-15 10:15 Improving the thingy:Putting the doodad in a different spot
o 2022-12-15 11:00
i 2022-12-15 11:00 Pair Programming
o 2022-12-15 12:00
i 2022-12-15 12:00 Lunch
o 2022-12-15 12:30
i 2022-12-15 12:30 Improving the thingy:Code Review
o 2022-12-15 14:00
i 2022-12-15 14:00 Break:Phooey! That's a lot of work!
o 2022-12-15 14:30
i 2022-12-15 14:30 Improving the thingy:Deployment
o 2022-12-15 17:30
```

## Next steps

Instructions/tutorial on how to fill up time log

Configurable...
* Log level
* Indentation, spaces vs tabs for indentation
* Parsing format
    * Section start format (YYYY-MM-DD has no special meaning other than identifying the section)
    * Line separator ` - `

Use TOML and/or command line args to...
* `-item-exclude`: Define items to exclude (ex: "Lunch.*", "Break.*")
* `-section-include` + `-item-include`: Define sections (days) to include (filter) and/or tasks to include (filter)
* Use case: Look at one day at a time using `timelog -section-include 2023-01-09 -item-exclude "Lunch.*" "Break.*"`

More stats? Across multiple sections (days)?
