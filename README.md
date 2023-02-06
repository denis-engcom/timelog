# Timelog

Takes time log input and produces aggregated time log output. Typical file names for this content would be `yyyy-mm_time_log.md` -> `yyyy-mm_time_log_aggregated.md`

## Installation

```sh
# Install by building from source
➜ go install github.com/denis-engcom/timelog@latest
```

## Example usage

Example input file `2022-12_time_log.md`:

```
# 2022-12-15

930: Troubleshoot webinar registration - Modal not opening
1000: Daily
1045: Troubleshoot webinar registration - Modal not opening
1130: Troubleshoot webinar registration - Understand why email don't fire outside production
1300: Lunch (and commute)
1400: OAS judging planning - Sync with Angela
1630: Release shadowing with Cenxiao
1715: OAS judging planning - Team sync for admin app
1800: Troubleshoot video attachments not working
1830: Troubleshoot webinar registration - Record ticket to improve testing
1845: EOD

---

Random notes that should just get ignored
```

Command:

```sh
➜ timelog < 2022-12_time_log.md > 2022-12_time_log_aggregated.md
```

Example input file `2022-12_time_log_aggregated.md`:

```
# 2022-12-15

8h15m
- 3h15m: OAS judging planning
    - 2h30m: Sync with Angela
    - 45m: Team sync for admin app
- 3h: Troubleshoot webinar registration
    - 1h30m: Understand why email don't fire outside production
    - 1h15m: Modal not opening
    - 15m: Record ticket to improve testing
- 45m: Daily
- 45m: Release shadowing with Cenxiao
- 30m: Troubleshoot video attachments not working
```

## Next steps

Instructions/tutorial on how to fill up time log

Configurable...
* Log level
* Indentation, spaces vs tabs for indentation
* Parsing format
    * Section start format (YYYY-MM-DD has no special meaning other than identifying the section)
    * Line format to replace default `^([[:digit:]]{3,4}): (.*)$`

Use TOML and/or command line args to...
* `-item-exclude`: Define items to exclude (ex: "Lunch.*", "Break.*")
* `-section-include` + `-item-include`: Define sections (days) to include (filter) and/or tasks to include (filter)
* Use case: Look at one day at a time using `timelog -section-include 2023-01-09 -item-exclude "Lunch.*" "Break.*"`

More stats? Across multiple sections (days)?
