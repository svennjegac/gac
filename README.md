# gac
gimme aws creds utils

## Install
`go install github.com/svennjegac/gac/cmd/gac@e29acb08127399bbfd4906c4864751d58e77bd89`

## Usage
You must have your `gimme-aws-creds` profiles already configured and stored in `~/.okta_aws_login_config` file.

Then just call `gac <profile>` and it will log you into your AWS profile.
If you don't have valid session, it will log you in via standard `gimme-aws-creds` command.

Afterward, you can just call `gac <profile2>` and it will log you into your second AWS profile.
If you don't have valid session, it will log you in via standard `gimme-aws-creds` command.

Now, if you call again `gac <profile>`, it won't prompt you with slow flow via Okta and MFA, but it will use cached credentials.
It caches credentials for as many profiles as you want.
Credentials are cached until they expire. If they expire, it will log you in again via standard `gimme-aws-creds` command.
