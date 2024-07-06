package model

const WebFingerKeyResource = "resource"
const WebFingerPrefixAcct = "acct:"
const WebFingerFmtUrl = "https://%s/.well-known/webfinger?" +
	WebFingerKeyResource + "=" + WebFingerPrefixAcct + "%s@%s"
