# Timeclock (hledger) output format

**You must install `hledger` for your operating system to make use of this output format.** https://hledger.org/install.html

The first thing you'll notice is that the generated timeclock output from `timelog -O timeclock` is an intermediate format that expands entries in a format that is understood by `hledger` (https://hledger.org/). `hledger` is general ledgering software and tracks commodities (in this case, hours) in "accounts". Below, you will find commands we can use to view the data.

## Timeclock output

```sh
➜ timelog -O timeclock > 2022-12.timeclock <<EOF
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

```sh
➜ cat 2022-12.timeclock
```
```
i 2022-12-15 09:30 Daily Meeting
o 2022-12-15 10:15
i 2022-12-15 10:15 Improving the thingy:Putting the doodad in a different spot
o 2022-12-15 12:00
i 2022-12-15 12:00 Lunch
o 2022-12-15 12:30
i 2022-12-15 12:30 Improving the thingy:Code Review
o 2022-12-15 14:00
i 2022-12-15 14:00 Break:Phooey! That's a lot of work!
o 2022-12-15 14:30
i 2022-12-15 14:30 Improving the thingy:Deployment
o 2022-12-15 17:30
i 2022-12-16 09:30 Daily Meeting
o 2022-12-16 10:00
i 2022-12-16 10:00 Improving the thingy:Removing the doodad altogether
o 2022-12-16 11:30
```

# `hledger register`

The `register` command will list and sum entries.

```sh
➜ hledger -ftimeclock:- register < 2022-12.timeclock
```
```
2022-12-15 09:30-10:15   (Daily Meeting)                                                 0.75h    0.75h
2022-12-15 10:15-12:00   (Improving the thingy:Putting the doodad in a different spot)   1.75h    2.50h
2022-12-15 12:00-12:30   (Lunch)                                                         0.50h    3.00h
2022-12-15 12:30-14:00   (Improving the thingy:Code Review)                              1.50h    4.50h
2022-12-15 14:00-14:30   (Break:Phooey! That's a lot of work!)                           0.50h    5.00h
2022-12-15 14:30-17:30   (Improving the thingy:Deployment)                               3.00h    8.00h
2022-12-16 09:30-10:00   (Daily Meeting)                                                 0.50h    8.50h
2022-12-16 10:00-11:30   (Improving the thingy:Removing the doodad altogether)           1.50h   10.00h
```

The `--daily` flag makes this look much nicer.

```sh
➜ hledger -ftimeclock:- register --daily < 2022-12.timeclock
```
```
2022-12-15   Break:Phooey! That's a lot of work!                           0.50h    0.50h
             Daily Meeting                                                 0.75h    1.25h
             Improving the thingy:Code Review                              1.50h    2.75h
             Improving the thingy:Deployment                               3.00h    5.75h
             Improving the thingy:Putting the doodad in a different spot   1.75h    7.50h
             Lunch                                                         0.50h    8.00h
2022-12-16   Daily Meeting                                                 0.50h    8.50h
             Improving the thingy:Removing the doodad altogether           1.50h   10.00h
```

Filter by date: `date:2022-12-16` or a range of dates: `date:2022-12-16..2022-12-17`

```sh
➜ hledger -ftimeclock:- register --daily date:2022-12-16 < 2022-12.timeclock
```
```
2022-12-16   Daily Meeting                                         0.50h   0.50h
             Improving the thingy:Removing the doodad altogether   1.50h   2.00h
```

Apply filtering via regular expression on entries that start with "Lunch".

```sh
➜ hledger -ftimeclock:- register --daily not:acct:^Lunch < 2022-12.timeclock
```
```
2022-12-15   Break:Phooey! That's a lot of work!                           0.50h   0.50h
             Daily Meeting                                                 0.75h   1.25h
             Improving the thingy:Code Review                              1.50h   2.75h
             Improving the thingy:Deployment                               3.00h   5.75h
             Improving the thingy:Putting the doodad in a different spot   1.75h   7.50h
2022-12-16   Daily Meeting                                                 0.50h   8.00h
             Improving the thingy:Removing the doodad altogether           1.50h   9.50h
```

Aggregate hours per day

```sh
➜ hledger -ftimeclock:- register --daily --pivot=date < 2022-12.timeclock
```
```
2022-12-15   8.00h         8.00h
2022-12-16   2.00h        10.00h
```

# `hledger balance`

Print chart of accounts in tree form, sort by amount

```sh
➜ hledger -ftimeclock:- balance --tree --sort-amount < 2022-12.timeclock
```
```
  7.75h  Improving the thingy
  3.00h    Deployment
  1.75h    Putting the doodad in a different spot
  1.50h    Code Review
  1.50h    Removing the doodad altogether
  1.25h  Daily Meeting
  0.50h  Break:Phooey! That's a lot of work!
  0.50h  Lunch
-------
 10.00h
```

We can obtain more focused output with `--daily` and filter by `date`.

```sh
➜ hledger -ftimeclock:- balance --tree --daily date:2022-12-15 < 2022-12.timeclock
```
```
Balance changes in 2022-12-15:

                                          || 2022-12-15
==========================================++============
 Break:Phooey! That's a lot of work!      ||      0.50h
 Daily Meeting                            ||      0.75h
 Improving the thingy                     ||      6.25h
   Code Review                            ||      1.50h
   Deployment                             ||      3.00h
   Putting the doodad in a different spot ||      1.75h
 Lunch                                    ||      0.50h
------------------------------------------++------------
                                          ||      8.00h
```