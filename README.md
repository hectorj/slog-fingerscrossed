# slog-fingerscrossed

Golang [slog](https://golang.google.cn/pkg/log/slog/).Handler with fingers-crossed strategy

Inspired by [PHP Monolog](https://seldaek.github.io/monolog/)'s finger-crossed strategy.

## What is the fingers-crossed strategy?

It takes your logs no matter their level and buffers them.

It then flushes them only if an error is logged.

This allows logging debug information when an error happens, while not polluting your logs when everything works fine.
