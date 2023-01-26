package common

var (
	WebappFront    func()
	WebappBack     func()
	ShortPath      = "/"
	PrivatePath    = ShortPath + "private"
	PublicPath     = ShortPath + "public"
	RemovePath     = ShortPath + "remove"
	DevShortPath   = "/testapp"
	DevPrivatePath = DevShortPath + "/private"
	DevPublicPath  = DevShortPath + "/public"
	DevRemovePath  = DevShortPath + "/remove"
	devBuild       bool
)

func init() {
	if devBuild {
		ShortPath = DevShortPath
		PrivatePath = DevPrivatePath
		PublicPath = DevPublicPath
		RemovePath = DevRemovePath
	}
}

// names used in both backend and frontend
const (
	PasswordProtected = "limitedAccess"
	FAdminKey         = "adKey"        // field name for admin keyK
	FAdminToken       = "adTok"        // field name for admin token
	FFileName         = "name"         // field name for uploaded file
	FShortKey         = "sKey"         // field name for short key used to access locked shorts (public/private/remove)
	FRequestTokenSeed = "RTS"          // field name for token seed generated by the frontend
	FSaltTokenID      = "stid"         // field name for salt token generated by the server
	FTokenID          = "TID"          // field name for token id used by create short url (generated using stid and user-agent) - used for controlling the rate for creating short url (FFU user token?)
	FShortDesc        = "sDesc"        // field name for short description
	FNamedPublic      = "sNamedPublic" // field name for short named public (instead of random)
	FPrvPassToken     = "sPrvPassTok"  // field name for short private pass token for access a locked short
	FPubPassToken     = "sPubPassTok"  // field name for short public pass token for access a locked short
	FRemPassToken     = "sRemPassTok"  // field name for short remove pass token for access a locked short
	FExpiration       = "sExp"         // field name for short expiration
	FRemove           = "sRem"         // field name for short remove feature
	FPrivate          = "sPrv"         // field name for short private feature
	FPass             = "pass"         // field name for clear pass for access a locked short
	// FPass             = "%70%61%73%73" // field name for clear pass for access a locked short

	HashLengthNamedFixedSize = 40
)

type ShortType int

const (
	ShortPublic ShortType = iota
	ShortPrivate
	ShortRemove
)
