package common

var (
	WebappFront       func()
	WebappBack        func()
	ShortPath         = "/"
	PrivatePath       = ShortPath + "private"
	DevShortPath      = "/testapp"
	DevPrivatePath    = DevShortPath + "/private"
	PasswordProtected = "limitedAccess"
	devBuild          bool
)

func init() {
	if devBuild {
		ShortPath = DevShortPath
		PrivatePath = DevPrivatePath
	}
}

// names used in both backend and frontend
const (
	FAdminKey         = "adKey"       // field name for admin keyK
	FAdminToken       = "adTok"       // field name for admin token
	FFileName         = "name"        // field name for uploaded file
	FPrivateKey       = "key"         // field name for key used to access private info
	FRequestTokenSeed = "RTS"         // field name for token seed generated by the frontend
	FSaltTokenID      = "stid"        // field name for salt token generated by the server
	FTokenID          = "TID"         // field name for token id used by create short url (generated using stid and user-agent) - used for controlling the rate for creating short url (FFU user token?)
	FShortDesc        = "sDesc"       // field name for short description
	FPrvPassToken     = "sPrvPassTok" // field name for short private pass token for access a locked short info
	FExpiration       = "sExp"        // field name for short expiration
	FRemove           = "sRem"        // field name for short remove feature
)
