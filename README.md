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