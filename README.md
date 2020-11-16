# Command Line Usage

```
Usage: urlfetchoninterval <url>

Fetches a URL on an interval.

Arguments:
  <url>    ($FETCH_URL)

Flags:
  -h, --help                     Show context-sensitive help.
  -i, --interval=1m              ($FETCH_INTERVAL)
  -t, --timeout=5s               ($FETCH_TIMEOUT)
      --headers=KEY=VALUE;...    ($FETCH_HEADERS)
  -p, --proxy=PROXY              ($PROXY_URL)
  -v, --verbose                  ($VERBOSE)
  ```

  # Docker Examples

```
  docker run -d scjalliance/urlfetchoninterval -i 5m somesite.example.com/wp-cron.php?doing_wp_cron
```
```
  docker run -d -e "FETCH_INTERVAL=5m" -e "FETCH_URL=somesite.example.com/wp-cron.php?doing_wp_cron" scjalliance/urlfetchoninterval 
```