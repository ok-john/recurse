package main

import (
	"os"
	"testing"
)

func Test_DecodeAGEKeys(t *testing.T) {
	recip, err := PEMDecodeRecipiant(rdgaf("data/publickey"))
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout.Write([]byte(recip.String()))

	id, err := PEMDecodeIdentity(rdgaf("data/privatekey"))
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout.Write([]byte(id.String()))

}
