package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/subpop/xrhidgen"
)

func main() {
	xrhidgen.SetSeed(103)
	fn := "John"
	ln := "Doe"
	an := "1234567890"
	ui := "user-123"
	id, err := xrhidgen.NewUserIdentity(xrhidgen.Identity{
		AccountNumber: &an,
	}, xrhidgen.User{
		FirstName: &fn,
		LastName:  &ln,
		UserID:    &ui,
	}, xrhidgen.Entitlements{})
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &buf)
	err = json.NewEncoder(encoder).Encode(id)
	if err != nil {
		panic(err)
	}
	if err := encoder.Close(); err != nil {
		panic(err)
	}
	fmt.Println(buf.String())
}
