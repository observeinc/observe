# Observe Command-line Tool

This tool allows you to interact with an Observe tenant from the command line
using the Observe API, without needing to use curl. Observe is a cloud based
observability platform that models machine data to help you debug issues with
software and businesses fast, and you can learn more about it at [https://observeinc.com/](https://observeinc.com)

Download pre-built binaries for popular operating systems at
[https://github.com/observeinc/observe/releases](https://github.com/observeinc/observe/releases)
Put the binary somewhere in your PATH and enjoy!

To install this tool from source, if you have `go` and `git` installed, you can
run:

    git clone https://github.com/observeinc/observe observe-tool
    cd observe-tool
    go install

The binary will show up in your `go` `bin` directory, whatever that is (this
depends on how you installed `go` and what your environment is.)

The Observe API is documented at:

[https://developer.observeinc.com/](https://developer.observeinc.com/)

## Quick Start

Assuming you know your customer id (the numeric identifier for your tenant
instance,) your site URL (the hostname part of the URL) and you have a login
with email/password available, you can run it like so:

### Log In

    $ observe --customerid 180316196377 --site observeinc.com login myname@example.com --sso
    Please visit https://180316196377.observeinc.com/settings/account?expectedUid=4711&serverToken=YUMMYTOKENFORTHEWINGOESHERE
    XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
    login: saved authtoken to section "default" in config file "/home/myname/.config/observe.yaml"

### Run a Query

    $ observe query -q 'pick_col timestamp, log | limit 10' -i 'Default.kubernetes/Container Logs'
    | timestamp           | log                                                       |
    -----------------------------------------------------------------------------------
    | 1679497866772157872 | E0420 16:20:00.123456 1 package/file.go:123] Hello World! |
    ...

### Getting Help

    observe --help
    observe help
    observe help list
    observe help query

Note that you will need to visit the Observe website using the link provided
when logged in as the appropriate user, to authorize the login request in the
web app. If you don't want to use this flow, and you have a direct
(email/password) login, you can skip the `--sso` flag and enter the password
directly. If you use SSO for Observe for a SSO provider that doesn't provide
your email address as a user property, you can use the User ID instead of the
email address. Your User ID for your user can be found in the Observe web app
in the Settings -> Account Setting tab, at the bottom of the page.

The login command will save the authentication token generated in a local
profile that will be re-used when you run other commands. If you do not want to
save this credential in that file, you must specify `--authtoken` on the command
line instead for each request, and should use `--no-save` when logging in.
The authtoken will be printed to stdout in either case.

If you use an SSO SAML integration such as Okta, Azure AD, Google, or PingOne,
then see the section about the `login` command for how to generate credentials.

The name of the dataset input for your query may vary based on what datasets
you have installed or built in your workspace, as well as the name of your
workspace -- explore the `Datasets` page in the Observe UI to find more
datasets. You can also directly use a dataset ID, like `41007104`, to specify
the dataset to read. To specify more than one dataset to use for joins and
unions, see the section on the `query` command below.

## Command Configuration

The general shape of the command is:

    observe [config-options] "command" [command-options]

The `config-options` are general options that specify which tenant and user
you're interacting as, and where to read and write data.

The `"command"` selects the major operating mode (such as `login` or `query`)
and the `command-options` tell the selected command exactly what to do.

Commands generally need three pieces of information to talk to the Observe
service:

1. Customer ID. This is the numeric identifier of the URL you use to interact
   with the Observe application. If your URL looks something like this,
   `https://180316196377.observeinc.com/`, then it is the `180316196377` part.
2. Site URL. This is the hostname part of the URL you use to interact with
   the Observe application. If your URL looks something like this,
   `180316196377.observeinc.com`, then it is the `observeinc.com` part.
3. Authentication token. This is the credential that tells the system who you
   are and what privileges you have. This credential is simliar to a password,
   so you should not leak it into source control or public places, and you may
   want to use a password manager to store it.

The customer ID is specified with the `--customerid` configuration option.

The site URL is specified with the `--site` configuration option.

The authentication token is specified with the `--authtoken` option.

You can also configure these into a profile (such as the `default` profile) in
the configuration file (see below) and select the profile with `--profile` or
the `OBSERVE_PROFILE` environment variable.

Note that the set of `config-options` is different from `command-options` and
each command has its own `command-options`. The `config-options` may not be
specified after the command, and the `command-options` may not be specified
before the command.

To see a list of all options and commands, use `observe help`

## Profile Files

The file `~/.config/observe.yaml` can store your commonly used `config options`
like customer ID and site URL and even authtoken (if handled with care.) The
file is a YAML file with a list of profiles, where you can choose a profile
using the `--profile` command-line option. A profile named `default` will be
used if no other profile is specified. You can specify another file for this
with `--config` or the environment variable `OBSERVE_CONFIG`.

The name of each option within the profile is the same as the name of the
corresponding command-line `config-option`. Here is an example:

    profile:
      default:
        customerid: "133742069123"
        site: "observeinc.com"
        authtoken: "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
      testing:
        customerid: "180316196377"
        site: "observe-testing.com"
        authtoken: "YYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYY"
        timestamp: true
        workspace: Testing

With this profile, you can use the `default` profile using `observe command
...` and the `testing` profile using `observe --profile=testing command ...`

If no profile is specified on the command line, but the `OBSERVE_PROFILE`
environment variable is set and not empty, that profile will be used.

## Output and Printing

The `--show-config` option will print the current config used, whether it's
coming from command-line or profile or environment or a combination of all
three.

The `--debug` option will print more verbose output about what's going on,
which may be helpful when automating actions and debugging why they fail after
the fact. These outputs are printed to the standard error output stream.

The `--quiet` option will print less verbose output about what's going on,
which may be helpful for people who don't want their terminal to have too much
noise in them.

The `--timestamp` option will timestamp each progress/debug/error message,
which may be helpful when using this tool in automation, but is harder to read
when using the tool interactively.

Error, progress, and debug information will all be sent to the standard error
stream. Actual data output by a particular command will be sent to standard
out, or to a file specified with `--output=filename`. It's worth noting that
data output redirection is a global configuration option and *not* an option to
any specific command.

Use `observe help objects` to get help on object types.

## Shell Completion

There is simple support for shell completion. A script that installs the
complete command for `bash` is included in the source code; you can modify the
script as appropriate for your shell if you use `zsh` or `fish`. Make sure this
script is sourced by your shell on start-up, either by adding it to your shell
profile file, or by installing it in the global shell completions directory for
your system. (Consult your shell/system/search engine for what the appropriate
location is for your particular case.)

