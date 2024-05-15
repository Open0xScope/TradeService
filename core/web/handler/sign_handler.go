package handler

import (
	"errors"

	"github.com/ChainSafe/gossamer/lib/common"
	"github.com/ChainSafe/gossamer/lib/crypto/sr25519"
)

func VerifySign(msg, pubkey, sigstr string) error {
	pubbytes, err := common.HexToBytes("0x" + pubkey)
	if err != nil {
		return err
	}

	pub, err := sr25519.NewPublicKey(pubbytes)
	if err != nil {
		return err
	}

	sighex, err := common.HexToBytes("0x" + sigstr)
	if err != nil {
		return err
	}

	ok, err := pub.Verify([]byte(msg), sighex)
	if err != nil {
		return err
	}

	if !ok {
		return errors.New("failed to verify signature")
	}

	return nil
}

func CheckAddress(pubkey, addr string) error {
	// randomly generated from subkey
	pub, err := common.HexToBytes("0x" + pubkey)
	if err != nil {
		return err
	}

	pk, err := sr25519.NewPublicKey(pub)
	if err != nil {
		return err
	}

	a := pk.Address()
	if addr != string(a) {
		return errors.New("address and pubkey not match")
	}

	return nil
}
