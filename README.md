# Observe Command-line Tool

This tool allows you to interact with an Observe tenant from the command line
using the Observe API, without needing to use curl. Observe is a cloud based
observability platform that models machine data to help you debug issues with
software and businesses fast, and you can learn more about it at [https://observeinc.com/](https://observeinc.com)

To install this tool, if you have `go` installed, you can run:

    go install https://github.com/observeinc/observe

If you don't have `go` installed, you can download pre-built binaries for
popular operating systems at [https://github.com/observeinc/observe/releases](https://github.com/observeinc/observe/releases)

The Observe API is documented at:

[https://developer.observeinc.com/](https://developer.observeinc.com/)

## Quick Start

Assuming you know your customer id (the numeric identifier for your tenant
instance,) your cluster URL (the hostname part of the URL) and you have a login
with email/password available, you can run it like so:

    $ observe --customerid 180316196377 --cluster observeinc.com login myname@example.com --read-password --save
    Password for 180316196377.observeinc.com: "myname@example.com": 
    XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
    login: saved authtoken to section "default" in config file "/home/myname/.config/observe.yaml"
    
    $ observe query -Q 'pick_col timestamp, log | limit 10' -I 'Default.kubernetes/Container Logs'
    | timestamp           | log                                                       |
    -----------------------------------------------------------------------------------
    | 1679497866772157872 | E0420 16:20:00.123456 1 package/file.go:123] Hello World! |
    ...

The login command will save the authentication token generated in a local
profile that will be re-used when you run other commands. If you do not want to
save this credential in that file, you can specify `--authtoken` on the command
line instead, and should not use `--save` when logging in.

If you use an SSO integration such as Okta, Azure AD, Google, or PingOne, then
see the section about the `login` command for how to generate credentials.

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
2. Cluster URL. This is the hostname part of the URL you use to interact with
   the Observe application. If your URL looks something like this,
   `180316196377.observeinc.com`, then it is the `observeinc.com` part.
3. Authentication token. This is the credential that tells the system who you
   are and what privileges you have. This credential is simliar to a password,
   so you should not leak it into source control or public places, and you may
   want to use a password manager to store it.

The customer ID is specified with the `--customerid` configuration option.

The cluster URL is specified with the `--cluster` configuration option.

The authentication token is specified with the `--authtoken` option.

You can also configure these into a profile (such as the `default` profile) in
the configuration file (see below) and select the profile with `--profile` or
the `OBSERVE_PROFILE` environment variable.

Note that the set of `config-options` is different from `command-options` and
each command has its own `command-options`. The `config-options` may not be
specified after the command, and the `command-options` may not be specified
before the command.

To see a list of all options and commands, use `observe --help`

## Profile Files

The file `~/.config/observe.yaml` can store your commonly used `config options`
like customer ID and cluster URL and even authtoken (if handled with care.) The
file is a YAML file with a list of profiles, where you can choose a profile
using the `--profile` command-line option. A profile named `default` will be
used if no other profile is specified. You can specify another file for this
with `--config` or the environment variable `OBSERVE_CONFIG`.

The name of each option within the profile is the same as the name of the
corresponding command-line `config-option`. Here is an example:

    profiles:
      default:
        customerid: "133742069123"
        cluster: "observeinc.com"
        authtoken: "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
      testing:
        customerid: "180316196377"
        cluster: "observe-testing.com"
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

Use `observe help --objects` to get help on object types.

