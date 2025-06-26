package test_util

import (
	"bytes"
	"encoding/base64"
	"encoding/json"

	"github.com/redhatinsights/platform-go-middlewares/v2/identity"
	"github.com/subpop/xrhidgen"
)

func GenerateIdentityStructFromTemplate(identityTemplate xrhidgen.Identity, userTemplate xrhidgen.User, entitlementsTemplate xrhidgen.Entitlements) identity.XRHID {
	xrhidgen.SetSeed(103)
	fn := "John"
	ln := "Doe"
	an := "1234567890"
	ui := "user-123"

	internelIdentity := identityTemplate
	if identityTemplate.AccountNumber == nil {
		internelIdentity.AccountNumber = &an
	}

	internalUser := userTemplate
	if userTemplate.FirstName == nil {
		internalUser.FirstName = &fn
	}
	if userTemplate.LastName == nil {
		internalUser.LastName = &ln
	}
	if userTemplate.UserID == nil {
		internalUser.UserID = &ui
	}

	id, err := xrhidgen.NewUserIdentity(internelIdentity, internalUser, entitlementsTemplate)
	if err != nil {
		panic(err)
	}

	return *id
}

func GenerateIdentityStruct() identity.XRHID {
	return GenerateIdentityStructFromTemplate(xrhidgen.Identity{}, xrhidgen.User{}, xrhidgen.Entitlements{})
}

func GenerateIdentityHeaderFromTemplate(identityTemplate xrhidgen.Identity, userTemplate xrhidgen.User, entitlementsTemplate xrhidgen.Entitlements) string {

	id := GenerateIdentityStructFromTemplate(identityTemplate, userTemplate, entitlementsTemplate)

	var buf bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &buf)
	err := json.NewEncoder(encoder).Encode(id)
	if err != nil {
		panic(err)
	}
	encoder.Close()
	return buf.String()
}

func GenerateIdentityHeader() string {
	return GenerateIdentityHeaderFromTemplate(
		xrhidgen.Identity{},
		xrhidgen.User{},
		xrhidgen.Entitlements{},
	)
}
