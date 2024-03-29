![logo](./logo.svg)

*A personal time tracker for terminal users*

---

`tiktak` will help you keep track of time you spent on different
activities—probably projects. It keeps its time records in ridiculous simple
text files. This should help if it comes to use your time records for some more
fancy application or keep them in a VCS repo, etc.

## Disclaimer
The data generated by this program is collected and computed with best effort
for correctness. __There is no guarantee for the fitness of that data for any
specific purpose!__ Don't blame any author of the software if numbers you get
from the program are incorrect in any way.

## Getting Started

- Get you the `tiktak` executable by
    * Downloading a [binary distribution from
      Github](https://github.com/fractalqb/tiktak/releases) or
    * If you have a [Go SDK](https://golang.org/dl/) installed using the
      idiomatic `go install git.fractalqb.de/fractalqb/tiktak/cmd/tiktak@latest`
- Put the `tiktak` executable in your path. (You pretend to be a terminal user,
  i.e. I won't have to explain…)
- `tiktak` will store its files in `$HOME/.local/share/fqb/tiktak` on UN!X like
  systems if `$HOME/.local/share` exists. On Windows, files go into
  `%HOME%\AppData\Roaming\fqb\tiktak`. Otherwise `tiktak` stores files in '.'.
- If you decide to keep files somwhere else, set the `TIKTAK_DATA` environment
  variable to the directory of your liking.
- Start working on _something cool_, i.e. type `tiktak /something/cool` in your
  terminal. This will start the subtaks _cool_ withing its parent task
  _something_.
- Do some miscellaneous work you want to account on the global level and type
  `tiktak /`
- After a while go back to work on _something_ by typing `tiktak something`.
  Note that you don't need to type a slash '/' because tiktak will find the
  already known task `/something`.
- 32 minutes later, check the time collected, i.e. type `tiktak`
```
SUMS: Sat, 01 Apr 2023; Week 13:                                   
-------------------------------------------------------------
Task             Today.  Today/  Week.  Week/  Month.  Month/
-------------------------------------------------------------
/something/cool   00:19       -  00:19      -   00:19       -
/something        00:31   00:50  00:31  00:50   00:31   00:50
/                 00:29   01:20  00:29  01:20   00:29   01:20
```
- This is enough work for today, say `tiktak -zzz`
- Check the recorded time spans in detail: `titkat -r spans`
```
Wed, 05 Apr 2023; Week 14                   
  10:40:09  10:59:12  00:19  /something/cool
  10:59:12  11:28:36  00:29  /              
  11:28:36  12:00:01  00:31  /something
```
- You can also get a time sheet for the current month showing accumulated time
  for _something_, including subtasks and the rest by typing `tiktak -r sheet
  /something /`
```
SHEET: Sat, 01 Apr 2023 – Sat, 01 Apr 2023                                 
---------------------------------------------------------------------------
Day          Start  Stop   Break  Work   /something  /something/cool      /
---------------------------------------------------------------------------
Wed, 05 Apr  10:40  12:00  00:00  01:20       00:31            00:19  00:29
---------------------------------------------------------------------------
   Average:  10:40  12:00  00:00  01:20       00:31            00:19  00:29
     Count:      1   Sum:  00:00  01:20       00:31            00:19  00:29
```
- Now comes the point where you think "Nice! But what can I do with this?". I'd
  suggest you write your time sheet into a CSV file to import it with some
  spreadsheet program: `tiktak -r sheet -layout csv -formats c /something /`
```
SHEET: Wed, 05 Apr 2023 – Wed, 05 Apr 2023;;;;;;;
Day;Start;Stop;Break;Work;/something;/something/cool;/
Wed, 05 Apr;10:40;12:00;00:00;01:20;00:31;00:19;00:29
Average:;10:40;12:00;00:00;01:20;00:31;00:19;00:29
Count:;1;Sum:;00:00;01:20;00:31;00:19;00:29
```

## Detailed Docs

### Build from Source

- Be sure to have [Go SDK installed](https://go.dev/doc/install). The most
  recent one should work. In case of doubt, the minimum required version is in
  `go.mod`.

- Clone the repo: `git clone https://github.com/fractalqb/tiktak.git`

- Enter the project dir: `cd tiktak`

- Run `go run mk/mk.go` and find the executables in `./cmd/tik*`

- To use Go's standard install, run `go run mk/mk.go -install`

### Locating time record files

tiktak keeps its files in a single directory. If you don't do nothing, tiktak
has OS-specific strategies to locate these files. If you use `tiktak -h` you
will find the query option that can show you the tiktak data directory: `tiktak
-q dir`

To control the directory yourself, set the `TIKTAK_DATA` environment variable to
the directory of your choice. Inside that directory you will find:

- Files named `yyyy-mm.tiktak` that store your tasks of month `mm` in year
  `yyyy`. tiktak creates them depending on the current time.

- An optional file `tiktak.yaml` (not `.yml`) with your tiktak configuration –
  if you created one.

- An optional file `template.yaml`. If tiktak has to create a new monthly file
  it will first copy the contents of `template.yaml` unchanged into the new
  file. One might use it to have some preconfigured tasks that are needed every
  month.

### Setting _now_

### Filters

### Migrating old files with `tikmig`
