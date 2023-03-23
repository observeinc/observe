package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/spf13/pflag"
)

var (
	flagSSO          bool
	flagReadPassword bool
	flagSaveConfig   bool
	flagsLogin       *pflag.FlagSet
)

func init() {
	flagsLogin = pflag.NewFlagSet("login", pflag.ContinueOnError)
	flagsLogin.BoolVarP(&flagSSO, "sso", "O", false, "Use SSO to log in")
	flagsLogin.Lookup("sso").NoOptDefVal = "true"
	flagsLogin.BoolVarP(&flagReadPassword, "read-password", "R", false, "Read password from stdin (avoid putting it on command line).")
	flagsLogin.Lookup("read-password").NoOptDefVal = "true"
	flagsLogin.BoolVarP(&flagSaveConfig, "save", "S", false, "Save the authtoken in the config file.")
	RegisterCommand(&Command{
		Name:            "login",
		Help:            "Generate an authtoken. Either provide email address and password on command line (not recommended,) or the --sso option to get a URL you can open in a browser to finish logging in.",
		Flags:           flagsLogin,
		Func:            cmdLogin,
		Unauthenticated: true,
	})
}

var (
	ErrLoginClashingOptions          = ObserveError{Msg: "cannot use --sso and --read-password together"}
	ErrLoginSaveRequiresProfile      = ObserveError{Msg: "--save requires --profile"}
	ErrLoginEmailRequired            = ObserveError{Msg: "email address required"}
	ErrLoginEmailAndPasswordRequired = ObserveError{Msg: "email address and password required"}
	ErrLoginPasswordIsNotValid       = ObserveError{Msg: "password is not valid"}
	ErrLoginRequiresCustomerId       = ObserveError{Msg: "customerid is required for login"}
	ErrLoginRequiresCluster          = ObserveError{Msg: "cluster is required for login"}
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

func cmdLogin(cfg *Config, op Output, args []string, hc *http.Client) error {
	if cfg.CustomerIdStr == "" {
		return ErrLoginRequiresCustomerId
	}
	if cfg.ClusterStr == "" {
		return ErrLoginRequiresCluster
	}
	if flagSSO && flagReadPassword {
		return ErrLoginClashingOptions
	}
	if flagSaveConfig && *FlagProfile == "" {
		return ErrLoginSaveRequiresProfile
	}
	if flagSSO {
		if len(args) != 2 {
			return ErrLoginEmailRequired
		}
		return cmdLoginDelegated(cfg, op, hc, args[1])
	} else if flagReadPassword {
		if len(args) != 2 {
			return ErrLoginEmailRequired
		}
		return cmdLoginReadPassword(cfg, op, hc, args[1])
	} else {
		if len(args) == 2 {
			op.Info("no password provided; use --read-password to read from stdin without this message\n")
		}
		if len(args) != 3 {
			return ErrLoginEmailAndPasswordRequired
		}
		return cmdLoginEmailPassword(cfg, op, hc, args[1], args[2])
	}
}

func cmdLoginReadPassword(cfg *Config, op Output, hc *http.Client, email string) error {
	pwdata, err := ReadPasswordFromTerminal(fmt.Sprintf("Password for %s.%s: %q: ", cfg.CustomerIdStr, cfg.ClusterStr, email))
	if err != nil {
		return err
	}
	if len(pwdata) < 8 {
		return ErrLoginPasswordIsNotValid
	}
	return cmdLoginEmailPassword(cfg, op, hc, email, string(pwdata))
}

func cmdLoginEmailPassword(cfg *Config, op Output, hc *http.Client, email, password string) error {
	req := &V1LoginRequest{
		UserEmail:    email,
		UserPassword: password,
		TokenName:    fmt.Sprintf("CLI login from %s", GetHostname()),
	}
	var resp V1LoginResponse
	err, status := RequestPOST(cfg, op, hc, "/v1/login", req, &resp)
	if err != nil {
		return err
	}
	if !resp.Ok {
		if len(resp.Message) > 0 {
			return NewObserveError(nil, "%s", resp.Message)
		}
		return NewObserveError(nil, "status %d", status)
	}
	return cmdLoginSuccess(cfg, op, resp.AccessKey, flagSaveConfig, GetConfigFilePath(), *FlagProfile)
}

func cmdLoginDelegated(cfg *Config, op Output, hc *http.Client, email string) error {
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
	if err, _ := RequestPOST(cfg, op, hc, "/v1/login/delegated", &req1, &resp1); err != nil {
		return NewObserveError(err, "POST error")
	}
	if !resp1.Ok {
		return NewObserveError(nil, "server error: %s", resp1.Message)
	}
	fmt.Fprintf(os.Stderr, "Please visit %s\n", resp1.Url)
	for {
		var resp2 V1LoginDelegatedStatus
		if err, _ := RequestGET(cfg, op, hc, "/v1/login/delegated/"+resp1.ServerToken, nil, &resp2); err != nil {
			op.Error("will retry after error: %s\n", err)
			time.Sleep(7 * time.Second) // errors demand slower retries
		} else if !resp2.Settled {
			op.Debug("determined that request %s is undetermined\n", resp1.ServerToken)
			if resp2.Message != "" {
				op.Info("%s\n", resp2.Message)
			}
		} else if resp2.AccessKey != "" {
			op.Debug("determined that request %s was accepted\n", resp1.ServerToken)
			// success!
			return cmdLoginSuccess(cfg, op, resp2.AccessKey, flagSaveConfig, GetConfigFilePath(), *FlagProfile)
		} else {
			// failure!
			op.Debug("determined that request %s was denied\n", resp1.ServerToken)
			return NewObserveError(nil, "failure: %s", resp2.Message)
		}
		// Note that the server side will long-poll, too, so this is mainly
		// to throttle in case something goes wrong in that function.
		time.Sleep(time.Second)
	}
}

// After success, report the token to the user and optionally save it to the profile
func cmdLoginSuccess(cfg *Config, op Output, accessKey string, saveConfig bool, filePath string, profileName string) error {
	op.Write([]byte(accessKey))
	op.Write([]byte("\n"))
	if saveConfig {
		stuff, err := ReadUntypedConfig(GetConfigFilePath(), false)
		if err != nil {
			return err
		}
		if len(stuff) == 0 {
			stuff = map[string]any{"profile": map[string]any{}}
		}
		if profiles, has := stuff["profile"]; has {
			if plist, is := profiles.(map[string]any); is {
				if this, has := plist[profileName]; has {
					if p, is := this.(map[string]any); is {
						p["authtoken"] = accessKey
						p["customerid"] = cfg.CustomerIdStr
						p["cluster"] = cfg.ClusterStr
					} else {
						return ErrCouldNotParseConfig
					}
				} else {
					plist[profileName] = map[string]any{
						"authtoken":  accessKey,
						"customerid": cfg.CustomerIdStr,
						"cluster":    cfg.ClusterStr,
					}
				}
				if err = SaveUntypedConfig(filePath, stuff); err == nil {
					op.Info("saved authtoken to section %q in config file %q\n", profileName, filePath)
				}
				return err
			}
		}
		return ErrCouldNotParseConfig
	}
	return nil
}
