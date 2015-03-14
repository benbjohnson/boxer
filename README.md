Timeboxer
=========

Timeboxer is a small utility application that runs commands on a given interval
and steps within that interval. For example, I may want to have my computer make
a tick noise every five minutes but make a beep every hour.

The timeboxer application is still in early development but it's meant to be
configurable so you can run any script on a given step or interval.


## History

I've been intrigued by time management techniques for a long time but I haven't
found any traditional approaches work well for me. The Pomodoro technique is
the most famous timeboxing technique and I've found limited success with it.
It requires that I plan what I'm going to do in the future and it has me take
breaks every 25 minutes. I find that too restrictive since I sometimes have to
jump on an ad hoc IM and I sometimes like to work for several hours without
taking a break.

The best solution I've found is to do reverse timeboxing. Instead of planning
for the next half hour, I simply write down what I did for the previous half
hour. This helps me do several things. First, I can see what I did during the
day. Some days it feels like I didn't accomplish anything but when I look back
at my list I can see all the little tasks. Second, it gives me a way to check
in frequently to make sure that I'm not focusing on a task for too long. For
example, I might think I can add a small feature in 30 minutes but sometimes
several hours go by and I'm in a rabbit hole. Checking in helps prevent that.

Reverse timeboxing has been great... when I do it. Unfortunately, it still
requires that I remember to write down something every half hour. So I decided
to build a tool that helps remind me of these intervals and gives me a way to
visualize the steps within these intervals. Thus, timeboxer was born.


## Usage

To install, simply run `go get` from the command line:

```sh
$ go get github.com/benbjohnson/timeboxer/...
```

Then run timeboxer:

```sh
$ timeboxer
Timeboxer running with 15m intervals and 1m steps...
```

Timeboxer will log whenever a new interval or step occurred.
