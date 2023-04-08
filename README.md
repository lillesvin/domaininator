# Domaininator

Fast generation and lookups of all domain names matching a specific regexp
pattern using [alixaxel](https://github.com/alixaxel)'s excelent
[genex](https://github.com/alixaxel/genex) library.

Great for keeping an eye on visually similar domains in order to catch
would-be-phishers in-between them registering a domain and deploying their
phishing kits.

## Usage

```
Usage of ./domaininator:
  -config string
        Config file to use
  -ip
        Show IPs on resolving domains
  -verbose
        Show all domain names, even if they are not registered
  -version
        Show version info and exit
  -workers int
        Number of parallel workers to run (default 16)
```

**NOTE:** Domaininator defaults to showing only domains with an A, NS or MX record.

## Config file

Domaininator looks for config files in the following places:

 * `./.domaininator.toml`
 * `./domaininator.toml`
 * `$XDG_CONFIG_HOME/domaininator/config.toml` or
   `$HOME/.config/domaininator/config.toml` on \*nix
 * `$HOME/Library/Application Support/domaininator/config.toml` on macOS
 * `%AppData%\domaininator\config.toml` on Windows
 * `$HOME/.domaininator.toml` on \*nix and macOS
 * `%USERPROFILE%\.domaininator.toml` on Windows
 * `/etc/domaininator.toml`

Other names/paths can be specified with the `-config` flag.

### Example config

```
# $HOME/domaininator/config.toml
pattern = "[gq][o0]{2}[gq][l1i]e\\.com"
whitelist = [
    "google.com"
]
ip = true
workers = 8
```

Run with:

```
$ domaininator -config $HOME/domaininator/config.toml
```

**NOTE:** Config file directives are overridden by command line flags.

## Command line invocation

Except for whitelists, everything supported in config files can be specificed
with command line flags. E.g. to find domains that could (with some fonts) be visually similar to "google.com":

```
$ domaininator "[gq][o0]{2}[gq][l1i]e\.com"
 48 / 48 [=========================================================================] 100.00% 0s

g00q1e.com: A,MX,NS
go0g1e.com: A,MX,NS
g0oqle.com: A,MX,NS
g0ogle.com: A,MX,NS
g0og1e.com: A,MX,NS
g0ogie.com: A,NS
google.com: A,MX,NS
googie.com: A,MX,NS
g00g1e.com: A,NS
g00gle.com: NS
gooqie.com: MX,NS
go0qle.com: A,NS
q00gle.com: A,NS
g00gie.com: NS
go0gle.com: NS
go0gie.com: A
q00qle.com: A,NS
q0ogle.com: A,NS
goog1e.com: NS
qoogie.com: A,MX,NS
qooq1e.com: A
qooqie.com: A,MX,NS
qoogle.com: NS
qoog1e.com: A,NS
```

