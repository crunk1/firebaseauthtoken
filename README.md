# firebaseauthtoken
A CLI tool for obtaining a Firebase auth token.

# Installation
Make sure your `$PATH` variable includes your `$GOPATH/bin` directory then
run:

`go get github.com/crunk1/firebaseauthtoken`

# Usage
```shell script
# email/pw method 1
> firebaseauthtoken --project-key XYZ --email foo@bar.com --pw p4ssw0rd

# email/pw method 2
> firebaseauthtoken --project-key XYZ --email foo@bar.com
Password: _

# Other authentication methods not yet implemented.
```
Output is the JWT returned from Firebase.