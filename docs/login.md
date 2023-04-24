# login

The login command allows you to specify credentials and obtain an authorization
token that can be used to authorize further interaction with the Observe
service. Remember that these authorization tokens have the same powers that you
do, and thus should be treated with care so they do not leak to the world. It
may be useful to create a new dummy user with restricted powers to use for
automation use cases.

There are three main forms of the `login` command:

1. `observe login email@example.com password`
   This form specifies the password straight on the command line. This is not
   particularly secure if you're using a shell with command history, but may be
   helpful when using the tool from automation where the history is not echoed.
2. `observe login email@example.com --read-password`
   This form reads the password from the standard input. What you type will not
   be echoed back, to avoid it leaking into the shell scrollback log. This is
   the most convenient option if you're using en email/password login to
   authenticate with Observe.
3. `observe login email@example.com --sso`
   This form will issue an authentication request to the specified user, that
   will be visible in the Observe web application. Open the URL printed by this
   command in a browser, and approve the request, to get the authentication
   token. This flow is the best way to generate an authentication token if you
   use SSO sign-in such as with SAML. Note that not all SSO IdPs provide an
   email address, and thus this method may not be available to your account.

To successfully `login` you must specify the `customerid` and `cluster` options,
but the `authtoken` option is not necessary (because creating one of those is
the whole point of the command!)

Unless you specify the `--no-save` option, the `login` command will update the
profile you use to store the authtoken in the `~/.config/observe.yaml` file.
You can specify which profile it will create or update with the `--profile`
configuration option before the `login` command.

Once you have logged in, you do not need to do this again, as long as the
authentication token remains unexpired, and is available either from the config
file or from the command line for each subsequent command. Configuration tokens
that are not used will expire after a period, commonly set to approximately ten
days.

## Example

    observe --profile staging \
        --customerid $CUSTOMERID \
        --cluster observeinc.com \
        login $USEREMAIL --sso

