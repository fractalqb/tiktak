# tiktak

*A personal time tracker for terminal users*

---

`tiktak` will help you keep track of time you spent on different activities—probably projects. It
keeps its time records in ridiculous simple JSON files. This should help if it comes to use your
time records for some more fancy application.

## Gettings Started

- Get you the `tiktak` executable by
    * Downloading a binary distribution from Github or
    * If you have a Go SDK installed using the idiomatic `go get -u git.fractalqb.de/fractalqb/tiktak`
- Put the `tiktak` executable in your path. (You pretend to be a terminal user, i.e.
   I won't have to explain…)
- Decide where to keep your time record files, let's say `$HOME/.local/share/tiktak`.
   (Non UN!X users are on there own here. But you're a terminal user… I told you) Otherwise, your
   files will be lying around wherever your current working dir is.
- Set the `TIKTAK_DATA` environment variable to the directory `$HOME/.local/share/tiktak`
- Start working on _something cool_, i.e. type `tiktak /something/cool` in your terminal. This will
   start the subtaks _cool_ withing its parent task _something_.
- Do some miscellaneous work you want to account on the global level and type `tiktak /`
- After a while go back to work on _something_ by typing `tiktak something`. Note that you don't
   need to type a slash '/' because tiktak will find the already known task `/something`.
- 32 minutes later, check the time collected, i.e. type `tiktak`
```
     TASK TIMES Sat, 30 Nov 2019:
       | ⏲ | Task            | Today |  Week |    All |
       +---+-----------------+-------+-------+--------+
       |   | /               | 00:28 | 00:28 |  00:28 |
       | ↻ | /something      | 00:32 | 00:32 |  00:32 |
       |   | /something/cool | 00:19 | 00:19 |  00:19 |
       +---+-----------------+-------+-------+--------+
       |                Sum: | 01:19 | 01:19 |  01:19 |

```
- This is enough work for today, say `tiktak -zzz`
- Check the recorded time spans in detail: `titkat -r spans`
```
     TIMESPANS PER DAY 11/2019:
       | ↹ | Start | Stop  | Dur   | Task            |
       +---+-------+-------+-------+-----------------+
       | Sat, 30 Nov 2019; KW48                      |
       +---+-------+-------+-------+-----------------+
       |   | 10:40 | 10:59 | 00:19 | /something/cool |
       |   | 10:59 | 11:28 | 00:28 | /               |
       |   | 11:28 | 12:00 | 00:32 | /something      |

```
- You can also get a time sheet for the current month showing accumulated time for _something_,
  including subtasks and the rest by typing `tiktak -r sheet /something /`
```
     TIME-SHEET 11/2019:
       | Day              | Start |  Stop |  Break |   Work | /something |      / |
       +------------------+-------+-------+--------+--------+------------+--------+
       | Sat, 30 Nov 2019 | 10:40 | 12:00 |  00:00 |  01:20 |      00:52 |  00:28 |
       +------------------+-------+-------+--------+--------+------------+--------+
       |             Avg: | 10:40 | 12:00 |  00:00 |  01:20 |      00:52 |  00:28 |
       |           Count: | 1     |  Sum: |  00:00 |  01:20 |      00:52 |  00:28 |

```
- Now comes the point where you thing "Nice! But what can I do with this?". I'd suggest you write
  your time sheet into a CSV file to import it with some spreadsheet program:
  `tiktak -r sheet -csv , -d f /something /`
```
     TIME-SHEET 11/2019:
     Day,Start,Stop,Break,Work,/something,/
     Sat, 30 Nov 2019,10:40,12:00,00.00,01.33,00.86,00.47
     Avg:,10:40,12:00,00.00,01.33,00.86,00.47
     Count:,1,Sum:,00.00,01.33,00.86,00.47
```