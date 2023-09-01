package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/pflag"
)

var (
	flagsLogin            *pflag.FlagSet
	flagLoginSSO          bool
	flagLoginReadPassword bool
	flagLoginNoSaveConfig bool
)

func init() {
	flagsLogin = pflag.NewFlagSet("login", pflag.ContinueOnError)
	flagsLogin.BoolVarP(&flagLoginSSO, "sso", "o", false, "Use SSO to log in")
	flagsLogin.Lookup("sso").NoOptDefVal = "true"
	flagsLogin.BoolVarP(&flagLoginReadPassword, "read-password", "r", false, "Read password from stdin (avoid putting it on command line).")
	flagsLogin.Lookup("read-password").NoOptDefVal = "true"
	flagsLogin.BoolVarP(&flagLoginNoSaveConfig, "no-save", "s", false, "Don't save the authtoken in the config file.")
	flagsLogin.Lookup("no-save").NoOptDefVal = "true"
	RegisterCommand(&Command{
		Name:            "login",
		Help:            "Generate an authorization token, from email or web account.",
		Flags:           flagsLogin,
		Func:            cmdLogin,
		Unauthenticated: true,
	})
}

var (
	ErrLoginClashingOptions          = ObserveError{Msg: "cannot use --sso and --read-password together"}
	ErrLoginSaveRequiresProfile      = ObserveError{Msg: "--save requires --profile"}
	ErrLoginEmailRequired            = ObserveError{Msg: "email address required"}
	ErrLoginEmailOrUserIdRequired    = ObserveError{Msg: "email address or user ID required"}
	ErrLoginEmailAndPasswordRequired = ObserveError{Msg: "email address and password required"}
	ErrLoginPasswordIsNotValid       = ObserveError{Msg: "password is not valid"}
	ErrLoginRequiresCustomerId       = ObserveError{Msg: "customerid is required for login"}
	ErrLoginRequiresSite             = ObserveError{Msg: "site is required for login"}
)

// This must match the ID issued for this tool on the server side.
const ObserveToolIntegrationId = "observe-tool-abdaf0"

type V1LoginRequest struct {
	UserEmail    string `json:"user_email"`
	UserPassword string `json:"user_password"`
	TokenName    string `json:"tokenName"`
}

type V1LoginResponse struct {
	Ok        bool   `json:"ok"`
	Message   string `json:"message,omitempty"`
	AccessKey string `json:"access_key,omitempty"`
}

type V1LoginDelegatedRequest struct {
	UserEmail   string `json:"userEmail"`
	ClientToken string `json:"clientToken"`
	Integration string `json:"integration"`
}

type V1LoginDelegatedResponse struct {
	Ok          bool   `json:"ok"`
	Message     string `json:"message,omitempty"` // really from error response
	Url         string `json:"url,omitempty"`
	ServerToken string `json:"serverToken,omitempty"`
}

type V1LoginDelegatedStatus struct {
	Ok        bool   `json:"ok"`
	Settled   bool   `json:"settled"`
	AccessKey string `json:"accessKey,omitempty"`
	Message   string `json:"message,omitempty"`
}

func cmdLogin(fa FuncArgs) error {
	if fa.cfg.CustomerIdStr == "" {
		return ErrLoginRequiresCustomerId
	}
	if fa.cfg.SiteStr == "" {
		return ErrLoginRequiresSite
	}
	if flagLoginSSO && flagLoginReadPassword {
		return ErrLoginClashingOptions
	}
	if !flagLoginNoSaveConfig && *FlagProfile == "" {
		return ErrLoginSaveRequiresProfile
	}
	fa.args = fa.args[1:]
	if flagLoginSSO {
		return cmdLoginDelegated(fa)
	} else if flagLoginReadPassword {
		return cmdLoginReadPassword(fa)
	} else {
		return cmdLoginEmailPassword(fa)
	}
}

func cmdLoginReadPassword(fa FuncArgs) error {
	if len(fa.args) != 1 {
		return ErrLoginEmailRequired
	}
	email := fa.args[0]
	pwdata, err := ReadPasswordFromTerminal(fmt.Sprintf("Password for %s.%s: %q: ", fa.cfg.CustomerIdStr, fa.cfg.SiteStr, email))
	if err != nil {
		return err
	}
	if len(pwdata) < 8 {
		return ErrLoginPasswordIsNotValid
	}
	fa.args = append(fa.args, string(pwdata))
	return cmdLoginEmailPassword(fa)
}

func cmdLoginEmailPassword(fa FuncArgs) error {
	if len(fa.args) == 1 {
		fa.op.Info("no password provided; use --read-password to read from stdin without this message\n")
	}
	if len(fa.args) != 2 {
		return ErrLoginEmailAndPasswordRequired
	}
	email := fa.args[0]
	password := fa.args[1]
	req := &V1LoginRequest{
		UserEmail:    email,
		UserPassword: password,
		TokenName:    fmt.Sprintf("CLI login from %s", GetHostname()),
	}
	var resp V1LoginResponse
	err, status := RequestPOST(fa.cfg, fa.op, fa.hc, "/v1/login", req, &resp, nil)
	if err != nil {
		return err
	}
	if !resp.Ok {
		if len(resp.Message) > 0 {
			return NewObserveError(nil, "%s", resp.Message)
		}
		return NewObserveError(nil, "status %d", status)
	}
	return cmdLoginSuccess{fa.cfg, fa.op, fa.fs, resp.AccessKey, !flagLoginNoSaveConfig, GetConfigFilePath(), *FlagProfile}.save()
}

func cmdLoginDelegated(fa FuncArgs) error {
	if len(fa.args) != 1 {
		return ErrLoginEmailOrUserIdRequired
	}
	email := fa.args[0]
	// this is a flow:
	// 1. send request to the configured tenant
	// 2. print a response URL for the user to visit
	// 3. if the request is OK, poll the tenant for completion
	// 4. when completed, maybe succeed or fail based on result
	req1 := V1LoginDelegatedRequest{
		UserEmail:   email,
		ClientToken: fmt.Sprintf("CLI login from %s at %s", GetHostname(), time.Now().Format("15:04:05")),
		Integration: ObserveToolIntegrationId,
	}
	var resp1 V1LoginDelegatedResponse
	if err, _ := RequestPOST(fa.cfg, fa.op, fa.hc, "/v1/login/delegated", &req1, &resp1, nil); err != nil {
		return NewObserveError(err, "POST error")
	}
	if !resp1.Ok {
		return NewObserveError(nil, "server error: %s", resp1.Message)
	}
	fmt.Fprintf(os.Stderr, "Please visit %s\n", resp1.Url)
	for {
		var resp2 V1LoginDelegatedStatus
		if err, _ := RequestGET(fa.cfg, fa.op, fa.hc, "/v1/login/delegated/"+resp1.ServerToken, nil, nil, &resp2); err != nil {
			fa.op.Error("will retry after error: %s\n", err)
			time.Sleep(7 * time.Second) // errors demand slower retries
		} else if !resp2.Settled {
			fa.op.Debug("determined that request %s is undetermined\n", resp1.ServerToken)
			if resp2.Message != "" {
				fa.op.Info("%s\n", resp2.Message)
			}
		} else if resp2.AccessKey != "" {
			fa.op.Debug("determined that request %s was accepted\n", resp1.ServerToken)
			// success!
			return cmdLoginSuccess{fa.cfg, fa.op, fa.fs, resp2.AccessKey, !flagLoginNoSaveConfig, GetConfigFilePath(), *FlagProfile}.save()
		} else {
			// failure!
			fa.op.Debug("determined that request %s was denied\n", resp1.ServerToken)
			return NewObserveError(nil, "failure: %s", resp2.Message)
		}
		// Note that the server side will long-poll, too, so this is mainly
		// to throttle in case something goes wrong in that function.
		time.Sleep(time.Second)
	}
}

// After success, report the token to the user and optionally save it to the profile
type cmdLoginSuccess struct {
	cfg         *Config
	op          Output
	fs          fileSystem
	accessKey   string
	saveConfig  bool
	filePath    string
	profileName string
}

func (c cmdLoginSuccess) save() error {
	c.op.Write([]byte(c.accessKey))
	c.op.Write([]byte("\n"))
	if c.saveConfig {
		stuff, err := ReadUntypedConfigFromFile(c.fs, GetConfigFilePath(), false)
		if err != nil {
			return err
		}
		if len(stuff) == 0 {
			stuff = map[string]any{"profile": map[string]any{}}
		}
		if profiles, has := stuff["profile"]; has {
			if plist, is := profiles.(map[string]any); is {
				if this, has := plist[c.profileName]; has {
					if p, is := this.(map[string]any); is {
						p["authtoken"] = c.accessKey
						p["customerid"] = c.cfg.CustomerIdStr
						p["site"] = c.cfg.SiteStr
					} else {
						return ErrCouldNotParseConfig
					}
				} else {
					plist[c.profileName] = map[string]any{
						"authtoken":  c.accessKey,
						"customerid": c.cfg.CustomerIdStr,
						"site":       c.cfg.SiteStr,
					}
				}
				if err = SaveUntypedConfig(c.fs, c.filePath, stuff); err == nil {
					c.op.Info("saved authtoken to section %q in config file %q\n", c.profileName, c.filePath)
				}
				return err
			}
		}
		return ErrCouldNotParseConfig
	}
	return nil
}
