# Timelog

The purpose of `timelog` is to provide a low-friction approach to logging time at work in a text file, and to sum hours to keep tabs on how much you're under/over working.
* **I was already logging time in a specific format, so I wrote the tool to parse that notation for (personal) reporting purposes.**

The provided `timelog` command is the conversion of time tracking notation (input) to reporting output.

More specifically, this method can work for you if...
* **You think of work in terms of when you start and stop activities (not necessarily in terms of duration).**
* **You prefer the freedom of editing a text file directly (to plan ahead or to enter data retroactively) instead of interacting with a start-stop time tracking tool.**

## Alternatives

If you want a text format geared for durations, [Timedot format](https://hledger.org/dev/hledger.html#timedot) is a light notation that can be used with `hledger` (https://hledger.org/).

If you prefer a start-stop time tracking tool, https://github.com/caarlos0/tasktimer is a simple command line alternative.

If you prefer web-based tooling, https://toggl.com looks to cover a lot of ground in terms of bells-and-whistles.

...and many more!

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
timelog -O timeclock < 2023-01_timelog.md | hledger -ftimeclock:- register --daily > 2023-01_timelog_reports_register.txt
```

## Example usage

### Input format

The input consists of day headers and activities.
* Day headers must be composed of `#`, a space, and the date in `YYYY-MM-DD` format.
    * Example: `# 2022-12-15`
    * All activities under the day header belong to that day
* Activities are composed of a day timestamp (hours and minutes), a space, and a hierarchy of activities and sub-activities separated by ` - ` (space, dash, space).
    * Example: `11:30 Primary activity`
    * Example 2: `11:30 Primary activity - Sub-activity`
    * The provided timestamp can be in the following formats: `9:30`, `09:30`, `930`, `0930`

### Timelog output format

The default reporting output from this tool is a per-day bullet-point list of activity (and sub-activity) durations.

```sh
➜ timelog <<EOF
# 2022-12-15

9:30 Daily Meeting
10:15 Improving the thingy - Putting the doodad in a different spot
12:00 Lunch
12:30 Improving the thingy - Code Review
14:00 Break - Phooey! That's a lot of work!
14:30 Improving the thingy - Deployment
17:30 EOD

# 2022-12-16

9:30 Daily Meeting
10:00 Improving the thingy - Removing the doodad altogether
11:30 EOD
EOF
```
```
# 2022-12-15

- 8h
	- 6h15m: Improving the thingy
		- 3h: Deployment
		- 1h45m: Putting the doodad in a different spot
		- 1h30m: Code Review
	- 45m: Daily Meeting
	- 30m: Break
		- 30m: Phooey! That's a lot of work!
	- 30m: Lunch

# 2022-12-16

- 2h
	- 1h30m: Improving the thingy
		- 1h30m: Removing the doodad altogether
	- 30m: Daily Meeting
```

### Timeclock (hledger) output format

The timeclock output format is an intermediate format that can be interpreted by `hledger` for more reporting flexibility.

See [README-timeclock.md](README-timeclock.md)

## Possible next steps

General notation from https://klog.jotaen.net
* Accepting go durations for retroactive time estimations (like timedot)
    * But how would we print the mix of timestamp and durations?

Idea from third-party libs
* Replacement for `hh:mm EOD` lines by accepting `hh:mm-hh:mm Last event of the day` which makes the ranges unambiguous
* https://klog.jotaen.net/#day-shifting
* https://klog.jotaen.net/#bookmarks
* https://github.com/larose/utt?tab=readme-ov-file#activity-type

Better reporting?
* https://github.com/larose/utt?tab=readme-ov-file#report-1
* https://klog.jotaen.net/#evaluate

Configurable...
* Log level
* Indentation, spaces vs tabs for indentation
* Parsing format
    * Section start format (YYYY-MM-DD has no special meaning other than identifying the section)
    * Allow capability of changing line separator ` - ` to `: ` (as an example)

More reporting flexibility with the default output. Command line args to...
* `-item-exclude`: Define items to exclude (ex: "Lunch.*", "Break.*")
* `-section-include` + `-item-include`: Define sections (days) to include (filter) and/or tasks to include (filter)
* Use case: Look at one day at a time using `timelog -section-include 2023-01-09 -item-exclude "Lunch.*" "Break.*"`

More stats? Across multiple sections (days)?
