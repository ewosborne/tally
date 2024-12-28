`tally` is a simple way to count the number of unique things in a stream of input.  It's basically `sort | uniq -c | sort`.

```
eric@Erics-MacBook-Air tally % tally < sample-file.txt
5 foo
6 baz
7 bar

eric@Erics-MacBook-Air tally % sort < sample-file.txt| uniq -c | sort
   5 foo
   6 baz
   7 bar
```

It doesn't do very much but I find myself reaching for it all the time.

`tally` takes a few arguments:

```
Usage:
  tally [flags]

Flags:
  -h, --help      help for tally
  -m, --min int   minimum number of matches to print a line
  -r, --reverse   Sort in reverse (descending count)
  -s, --string    Sort by string, not count
      --sum       Show sum of count
```

`-m, --min int` is a filter; if a given line has fewer than `-m` matches it's skipped entirely.  This is useful for filtering out the long tail of noise in a large set of matches.

```
eric@Erics-MacBook-Air tally % cat sample-file.txt| tally -m 5
5 foo
6 baz
7 bar
eric@Erics-MacBook-Air tally % cat sample-file.txt| tally -m 6
6 baz
7 bar
```


`-r, --reverse` reverses the output.


`-s, --string` changes how the columns are sorted:

```
eric@Erics-MacBook-Air tally % cat sample-file.txt| tally
5 foo
6 baz
7 bar
eric@Erics-MacBook-Air tally % cat sample-file.txt| tally  -s
7 bar
6 baz
5 foo
```