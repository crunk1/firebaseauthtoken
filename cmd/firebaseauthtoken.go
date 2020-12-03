package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"syscall"

	"github.com/spf13/pflag"
	"golang.org/x/crypto/ssh/terminal"
)

const (
	exitCodeOK  exitCode = iota
	exitCodeInternalError
	exitCodeConnectionError
	exitCodeFlagError
	exitCodeAuthenticationError
)

var (
	emailRgx      = regexp.MustCompile(`(?:[a-z0-9!#$%&'*+/=?^_` + "`" + `{|}~-]+(?:\.[a-z0-9!#$%&'*+/=?^_` + "`" + `{|}~-]+)*|"(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21\x23-\x5b\x5d-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])*")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\[(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?|[a-z0-9-]*[a-z0-9]:(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21-\x5a\x53-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])+)\])`)
	exitCodeNames = []string{"OK", "INTERNAL_ERROR", "CONNECTION_ERROR", "FLAG_ERROR", "AUTHENTICATION_ERROR"}
)

type exitCode int

func (c exitCode) String() string {
	if int(c) < len(exitCodeNames) {
		return exitCodeNames[c]
	}
	return ""
}

func main() {
	// Flag parsing and checking.
	projectKeyP := pflag.String("project-key", "", "firebase project web API key of the project to authenticate against")
	emailP := pflag.String("email", "", "email of user to authenticate")
	pwP := pflag.String("pw", "", "password of user to authenticate")
	pflag.Parse()
	var missingRequiredFlags []string
	if *emailP == "" {
		missingRequiredFlags = append(missingRequiredFlags, "--email")
	}
	if *projectKeyP == "" {
		missingRequiredFlags = append(missingRequiredFlags, "--project-key")
	}
	if len(missingRequiredFlags) > 0 {
		exit(exitCodeFlagError, fmt.Sprintf("missing required flags %v", missingRequiredFlags))
	}
	if !emailRgx.Match([]byte(*emailP)) {
		exit(exitCodeFlagError, fmt.Sprintf("bad email %q", *emailP))
	}
	if *pwP == "" {
		fmt.Print("Password: ")
		pwBs, err := terminal.ReadPassword(syscall.Stdin)
		fmt.Println()
		if err != nil {
			exit(exitCodeInternalError, "reading password - "+err.Error())
		}
		*pwP = string(pwBs)
	}

	// Call Firebase
	firebaseUrl := "https://www.googleapis.com/identitytoolkit/v3/relyingparty/verifyPassword?key=" + *projectKeyP
	resp, err := http.Post(firebaseUrl, "application/json", bytes.NewBufferString(
		fmt.Sprintf(`{"email":"%s","password":"%s","returnSecureToken":true}`, *emailP, *pwP),
	))
	if err != nil {
		exit(exitCodeConnectionError, err.Error())
	}
	if resp.StatusCode == http.StatusBadRequest {
		exit(exitCodeAuthenticationError, "invalid authentication credentials")
	}
	if resp.StatusCode != http.StatusOK {
		exit(exitCodeConnectionError, "response status "+resp.Status)
	}

	// Parse and return response
	defer func() { _ = resp.Body.Close() }()
	respData := map[string]interface{}{}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		exit(exitCodeConnectionError, "could not read all of firebase response")
	}
	err = json.Unmarshal(respBody, &respData)
	if err != nil {
		exit(exitCodeInternalError, "could not unmarshal firebase response")
	}
	fmt.Print(respData["idToken"])
}

func exit(c exitCode, nonZeroMsg string) {
	if c == exitCodeOK {
		os.Exit(0)
	}
	if nonZeroMsg != "" {
		_, _ = fmt.Fprintf(os.Stderr, "%s: %s\n", c.String(), nonZeroMsg)
	} else {
		_, _ = fmt.Fprintln(os.Stderr, c.String())
	}
	os.Exit(int(c))
}
