# fup

A simple-minded workstation initializer.

## Why?

One advantage over using established configuration management tools for the same purpose is that you don't need to worry about installing packages or ensuring that SSH access works correctly. On a pristine installation you can use a static `fup` binary which only operates on the local host.

## How?

Just point to a config file (default `~/.config/fup/fup.yml`, override with `-f`, `--file`, can be a remote URL).

There is a config file with a non-zero number of comments under [contrib/simple.yml](https://github.com/femnad/fup/blob/main/contrib/simple.yml).

## Better Alternatives?

https://github.com/comtrya/comtrya, of course.
