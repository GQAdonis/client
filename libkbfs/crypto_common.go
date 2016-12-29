// Copyright 2016 Keybase Inc. All rights reserved.
// Use of this source code is governed by a BSD
// license that can be found in the LICENSE file.

package libkbfs

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"

	"github.com/keybase/client/go/libkb"
	"github.com/keybase/client/go/protocol/keybase1"
	"github.com/keybase/kbfs/kbfsblock"
	"github.com/keybase/kbfs/kbfscodec"
	"github.com/keybase/kbfs/kbfscrypto"
	"github.com/keybase/kbfs/kbfshash"
	"github.com/keybase/kbfs/tlf"
	"golang.org/x/crypto/nacl/box"
	"golang.org/x/crypto/nacl/secretbox"
)

// TODO: Wrap errors coming from Crypto.

// CryptoCommon contains many of the function implementations need for
// the Crypto interface, which can be reused by other implementations.
type CryptoCommon struct {
	codec kbfscodec.Codec
}

var _ cryptoPure = (*CryptoCommon)(nil)

// MakeCryptoCommon returns a default CryptoCommon object.
func MakeCryptoCommon(codec kbfscodec.Codec) CryptoCommon {
	return CryptoCommon{codec}
}

// MakeRandomTlfID implements the Crypto interface for CryptoCommon.
func (c CryptoCommon) MakeRandomTlfID(isPublic bool) (tlf.ID, error) {
	return tlf.MakeRandomID(isPublic)
}

// MakeRandomBranchID implements the Crypto interface for CryptoCommon.
func (c CryptoCommon) MakeRandomBranchID() (BranchID, error) {
	var id BranchID
	// Loop just in case we randomly pick the null or local squash
	// branch IDs.
	for id == NullBranchID || id == PendingLocalSquashBranchID {
		err := kbfscrypto.RandRead(id.id[:])
		if err != nil {
			return BranchID{}, err
		}
	}
	return id, nil
}

// MakeMdID implements the Crypto interface for CryptoCommon.
func (c CryptoCommon) MakeMdID(md BareRootMetadata) (MdID, error) {
	// Make sure that the serialized metadata is set; otherwise we
	// won't get the right MdID.
	if md.GetSerializedPrivateMetadata() == nil {
		return MdID{}, MDMissingDataError{md.TlfID()}
	}

	buf, err := c.codec.Encode(md)
	if err != nil {
		return MdID{}, err
	}

	h, err := kbfshash.DefaultHash(buf)
	if err != nil {
		return MdID{}, err
	}
	return MdID{h}, nil
}

// MakeMerkleHash implements the Crypto interface for CryptoCommon.
func (c CryptoCommon) MakeMerkleHash(md *RootMetadataSigned) (MerkleHash, error) {
	buf, err := c.codec.Encode(md)
	if err != nil {
		return MerkleHash{}, err
	}
	h, err := kbfshash.DefaultHash(buf)
	if err != nil {
		return MerkleHash{}, err
	}
	return MerkleHash{h}, nil
}

// MakeTLFWriterKeyBundleID implements the Crypto interface for CryptoCommon.
func (c CryptoCommon) MakeTLFWriterKeyBundleID(wkb TLFWriterKeyBundleV3) (
	TLFWriterKeyBundleID, error) {
	if len(wkb.Keys) == 0 {
		return TLFWriterKeyBundleID{}, errors.New(
			"Writer key bundle with no keys")
	}
	buf, err := c.codec.Encode(wkb)
	if err != nil {
		return TLFWriterKeyBundleID{}, err
	}
	h, err := kbfshash.DefaultHash(buf)
	if err != nil {
		return TLFWriterKeyBundleID{}, err
	}
	return TLFWriterKeyBundleID{h}, nil
}

// MakeTLFReaderKeyBundleID implements the Crypto interface for CryptoCommon.
func (c CryptoCommon) MakeTLFReaderKeyBundleID(rkb TLFReaderKeyBundleV3) (
	TLFReaderKeyBundleID, error) {
	buf, err := c.codec.Encode(rkb)
	if err != nil {
		return TLFReaderKeyBundleID{}, err
	}
	h, err := kbfshash.DefaultHash(buf)
	if err != nil {
		return TLFReaderKeyBundleID{}, err
	}
	return TLFReaderKeyBundleID{h}, nil
}

// MakeTemporaryBlockID implements the Crypto interface for CryptoCommon.
func (c CryptoCommon) MakeTemporaryBlockID() (kbfsblock.ID, error) {
	return kbfsblock.MakeTemporaryID()
}

// MakeBlockRefNonce implements the Crypto interface for CryptoCommon.
func (c CryptoCommon) MakeBlockRefNonce() (nonce kbfsblock.RefNonce, err error) {
	return kbfsblock.MakeRefNonce()
}

// MakeRandomBlockCryptKeyServerHalf implements the Crypto interface
// for CryptoCommon.
func (c CryptoCommon) MakeRandomBlockCryptKeyServerHalf() (
	kbfscrypto.BlockCryptKeyServerHalf, error) {
	return kbfscrypto.MakeRandomBlockCryptKeyServerHalf()
}

// MakeRandomTLFEphemeralKeys implements the Crypto interface for
// CryptoCommon.
func (c CryptoCommon) MakeRandomTLFEphemeralKeys() (
	kbfscrypto.TLFEphemeralPublicKey, kbfscrypto.TLFEphemeralPrivateKey,
	error) {
	keyPair, err := libkb.GenerateNaclDHKeyPair()
	if err != nil {
		return kbfscrypto.TLFEphemeralPublicKey{},
			kbfscrypto.TLFEphemeralPrivateKey{}, err
	}

	ePubKey := kbfscrypto.MakeTLFEphemeralPublicKey(keyPair.Public)
	ePrivKey := kbfscrypto.MakeTLFEphemeralPrivateKey(*keyPair.Private)

	return ePubKey, ePrivKey, nil
}

// MakeRandomTLFKeys implements the Crypto interface for CryptoCommon.
func (c CryptoCommon) MakeRandomTLFKeys() (kbfscrypto.TLFPublicKey,
	kbfscrypto.TLFPrivateKey, kbfscrypto.TLFCryptKey, error) {
	publicKey, privateKey, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return kbfscrypto.TLFPublicKey{}, kbfscrypto.TLFPrivateKey{},
			kbfscrypto.TLFCryptKey{}, err
	}

	pubKey := kbfscrypto.MakeTLFPublicKey(*publicKey)
	privKey := kbfscrypto.MakeTLFPrivateKey(*privateKey)

	var data [32]byte
	err = kbfscrypto.RandRead(data[:])
	if err != nil {
		return kbfscrypto.TLFPublicKey{}, kbfscrypto.TLFPrivateKey{},
			kbfscrypto.TLFCryptKey{}, err
	}

	cryptKey := kbfscrypto.MakeTLFCryptKey(data)

	return pubKey, privKey, cryptKey, nil
}

// MakeRandomTLFCryptKeyServerHalf implements the Crypto interface for
// CryptoCommon.
func (c CryptoCommon) MakeRandomTLFCryptKeyServerHalf() (
	serverHalf kbfscrypto.TLFCryptKeyServerHalf, err error) {
	var data [32]byte
	err = kbfscrypto.RandRead(data[:])
	if err != nil {
		return kbfscrypto.TLFCryptKeyServerHalf{}, err
	}
	serverHalf = kbfscrypto.MakeTLFCryptKeyServerHalf(data)
	return serverHalf, nil
}

// EncryptTLFCryptKeyClientHalf implements the Crypto interface for
// CryptoCommon.
func (c CryptoCommon) EncryptTLFCryptKeyClientHalf(
	privateKey kbfscrypto.TLFEphemeralPrivateKey,
	publicKey kbfscrypto.CryptPublicKey,
	clientHalf kbfscrypto.TLFCryptKeyClientHalf) (
	encryptedClientHalf EncryptedTLFCryptKeyClientHalf, err error) {
	var nonce [24]byte
	err = kbfscrypto.RandRead(nonce[:])
	if err != nil {
		return
	}

	keypair, err := libkb.ImportKeypairFromKID(publicKey.KID())
	if err != nil {
		return
	}

	dhKeyPair, ok := keypair.(libkb.NaclDHKeyPair)
	if !ok {
		err = libkb.KeyCannotEncryptError{}
		return
	}

	clientHalfData := clientHalf.Data()
	privateKeyData := privateKey.Data()
	encryptedBytes := box.Seal(nil, clientHalfData[:], &nonce, (*[32]byte)(&dhKeyPair.Public), &privateKeyData)

	encryptedClientHalf = EncryptedTLFCryptKeyClientHalf{
		encryptedData{
			Version:       EncryptionSecretbox,
			Nonce:         nonce[:],
			EncryptedData: encryptedBytes,
		},
	}
	return
}

func (c CryptoCommon) encryptData(data []byte, key [32]byte) (encryptedData, error) {
	var nonce [24]byte
	err := kbfscrypto.RandRead(nonce[:])
	if err != nil {
		return encryptedData{}, err
	}

	sealedData := secretbox.Seal(nil, data, &nonce, &key)

	return encryptedData{
		Version:       EncryptionSecretbox,
		Nonce:         nonce[:],
		EncryptedData: sealedData,
	}, nil
}

// EncryptPrivateMetadata implements the Crypto interface for CryptoCommon.
func (c CryptoCommon) EncryptPrivateMetadata(
	pmd PrivateMetadata, key kbfscrypto.TLFCryptKey) (
	encryptedPmd EncryptedPrivateMetadata, err error) {
	encodedPmd, err := c.codec.Encode(pmd)
	if err != nil {
		return
	}

	encryptedData, err := c.encryptData(encodedPmd, key.Data())
	if err != nil {
		return
	}

	encryptedPmd = EncryptedPrivateMetadata{encryptedData}
	return
}

func (c CryptoCommon) decryptData(encryptedData encryptedData, key [32]byte) ([]byte, error) {
	if encryptedData.Version != EncryptionSecretbox {
		return nil, UnknownEncryptionVer{encryptedData.Version}
	}

	var nonce [24]byte
	if len(encryptedData.Nonce) != len(nonce) {
		return nil, InvalidNonceError{encryptedData.Nonce}
	}
	copy(nonce[:], encryptedData.Nonce)

	decryptedData, ok := secretbox.Open(nil, encryptedData.EncryptedData, &nonce, &key)
	if !ok {
		return nil, libkb.DecryptionError{}
	}

	return decryptedData, nil
}

// DecryptPrivateMetadata implements the Crypto interface for CryptoCommon.
func (c CryptoCommon) DecryptPrivateMetadata(
	encryptedPmd EncryptedPrivateMetadata, key kbfscrypto.TLFCryptKey) (
	PrivateMetadata, error) {
	encodedPmd, err := c.decryptData(encryptedPmd.encryptedData, key.Data())
	if err != nil {
		return PrivateMetadata{}, err
	}

	var pmd PrivateMetadata
	err = c.codec.Decode(encodedPmd, &pmd)
	if err != nil {
		return PrivateMetadata{}, err
	}

	return pmd, nil
}

const minBlockSize = 256

// nextPowerOfTwo returns next power of 2 greater than the input n.
// https://en.wikipedia.org/wiki/Power_of_two#Algorithm_to_round_up_to_power_of_two
func nextPowerOfTwo(n uint32) uint32 {
	if n < minBlockSize {
		return minBlockSize
	}
	if n&(n-1) == 0 {
		// if n is already power of 2, get the next one
		n++
	}

	n--
	n = n | (n >> 1)
	n = n | (n >> 2)
	n = n | (n >> 4)
	n = n | (n >> 8)
	n = n | (n >> 16)
	n++

	return n
}

const padPrefixSize = 4

// padBlock adds random padding to an encoded block.
func (c CryptoCommon) padBlock(block []byte) ([]byte, error) {
	blockLen := uint32(len(block))
	overallLen := nextPowerOfTwo(blockLen)
	padLen := int64(overallLen - blockLen)

	buf := bytes.NewBuffer(make([]byte, 0, overallLen+padPrefixSize))

	// first 4 bytes contain the length of the block data
	if err := binary.Write(buf, binary.LittleEndian, blockLen); err != nil {
		return nil, err
	}

	// followed by the actual block data
	buf.Write(block)

	// followed by random data
	n, err := io.CopyN(buf, rand.Reader, padLen)
	if err != nil {
		return nil, err
	}
	if n != padLen {
		return nil, kbfscrypto.UnexpectedShortCryptoRandRead{}
	}

	return buf.Bytes(), nil
}

// depadBlock extracts the actual block data from a padded block.
func (c CryptoCommon) depadBlock(paddedBlock []byte) ([]byte, error) {
	buf := bytes.NewBuffer(paddedBlock)

	var blockLen uint32
	if err := binary.Read(buf, binary.LittleEndian, &blockLen); err != nil {
		return nil, err
	}
	blockEndPos := int(blockLen + padPrefixSize)

	if len(paddedBlock) < blockEndPos {
		return nil, PaddedBlockReadError{ActualLen: len(paddedBlock), ExpectedLen: blockEndPos}
	}
	return buf.Next(int(blockLen)), nil
}

// EncryptBlock implements the Crypto interface for CryptoCommon.
func (c CryptoCommon) EncryptBlock(block Block, key kbfscrypto.BlockCryptKey) (
	plainSize int, encryptedBlock EncryptedBlock, err error) {
	encodedBlock, err := c.codec.Encode(block)
	if err != nil {
		return
	}

	paddedBlock, err := c.padBlock(encodedBlock)
	if err != nil {
		return
	}

	encryptedData, err := c.encryptData(paddedBlock, key.Data())
	if err != nil {
		return
	}

	plainSize = len(encodedBlock)
	encryptedBlock = EncryptedBlock{encryptedData}
	return
}

// DecryptBlock implements the Crypto interface for CryptoCommon.
func (c CryptoCommon) DecryptBlock(
	encryptedBlock EncryptedBlock, key kbfscrypto.BlockCryptKey,
	block Block) error {
	paddedBlock, err := c.decryptData(encryptedBlock.encryptedData, key.Data())
	if err != nil {
		return err
	}

	encodedBlock, err := c.depadBlock(paddedBlock)
	if err != nil {
		return err
	}

	err = c.codec.Decode(encodedBlock, &block)
	if err != nil {
		return BlockDecodeError{err}
	}
	return nil
}

// GetTLFCryptKeyServerHalfID implements the Crypto interface for CryptoCommon.
func (c CryptoCommon) GetTLFCryptKeyServerHalfID(
	user keybase1.UID, devicePubKey kbfscrypto.CryptPublicKey,
	serverHalf kbfscrypto.TLFCryptKeyServerHalf) (
	TLFCryptKeyServerHalfID, error) {
	key, err := serverHalf.MarshalBinary()
	if err != nil {
		return TLFCryptKeyServerHalfID{}, err
	}
	data := append(user.ToBytes(), devicePubKey.KID().ToBytes()...)
	hmac, err := kbfshash.DefaultHMAC(key, data)
	if err != nil {
		return TLFCryptKeyServerHalfID{}, err
	}
	return TLFCryptKeyServerHalfID{
		ID: hmac,
	}, nil
}

// VerifyTLFCryptKeyServerHalfID implements the Crypto interface for CryptoCommon.
func (c CryptoCommon) VerifyTLFCryptKeyServerHalfID(
	serverHalfID TLFCryptKeyServerHalfID,
	user keybase1.UID, deviceKID keybase1.KID,
	serverHalf kbfscrypto.TLFCryptKeyServerHalf) error {
	key, err := serverHalf.MarshalBinary()
	if err != nil {
		return err
	}
	data := append(user.ToBytes(), deviceKID.ToBytes()...)
	return serverHalfID.ID.Verify(key, data)
}

// EncryptMerkleLeaf encrypts a Merkle leaf node with the
// kbfscrypto.TLFPublicKey.
func (c CryptoCommon) EncryptMerkleLeaf(leaf MerkleLeaf,
	pubKey kbfscrypto.TLFPublicKey,
	nonce *[24]byte, ePrivKey kbfscrypto.TLFEphemeralPrivateKey) (
	EncryptedMerkleLeaf, error) {
	// encode the clear-text leaf
	leafBytes, err := c.codec.Encode(leaf)
	if err != nil {
		return EncryptedMerkleLeaf{}, err
	}
	// encrypt the encoded leaf
	pubKeyData := pubKey.Data()
	privKeyData := ePrivKey.Data()
	encryptedData := box.Seal(nil, leafBytes[:], nonce, &pubKeyData, &privKeyData)
	return EncryptedMerkleLeaf{
		Version:       EncryptionSecretbox,
		EncryptedData: encryptedData,
	}, nil
}

// DecryptMerkleLeaf decrypts a Merkle leaf node with the
// kbfscrypto.TLFPrivateKey.
func (c CryptoCommon) DecryptMerkleLeaf(encryptedLeaf EncryptedMerkleLeaf,
	privKey kbfscrypto.TLFPrivateKey, nonce *[24]byte,
	ePubKey kbfscrypto.TLFEphemeralPublicKey) (*MerkleLeaf, error) {
	if encryptedLeaf.Version != EncryptionSecretbox {
		return nil, UnknownEncryptionVer{encryptedLeaf.Version}
	}
	pubKeyData := ePubKey.Data()
	privKeyData := privKey.Data()
	leafBytes, ok := box.Open(nil, encryptedLeaf.EncryptedData[:], nonce, &pubKeyData, &privKeyData)
	if !ok {
		return nil, libkb.DecryptionError{}
	}
	// decode the leaf
	var leaf MerkleLeaf
	if err := c.codec.Decode(leafBytes, &leaf); err != nil {
		return nil, err
	}
	return &leaf, nil
}

// EncryptTLFCryptKeys implements the Crypto interface for CryptoCommon.
func (c CryptoCommon) EncryptTLFCryptKeys(
	oldKeys []kbfscrypto.TLFCryptKey, key kbfscrypto.TLFCryptKey) (
	encryptedKeys EncryptedTLFCryptKeys, err error) {
	encodedKeys, err := c.codec.Encode(oldKeys)
	if err != nil {
		return
	}

	encryptedData, err := c.encryptData(encodedKeys, key.Data())
	if err != nil {
		return
	}

	encryptedKeys = EncryptedTLFCryptKeys{encryptedData}
	return
}

// DecryptTLFCryptKeys implements the Crypto interface for CryptoCommon.
func (c CryptoCommon) DecryptTLFCryptKeys(
	encKeys EncryptedTLFCryptKeys, key kbfscrypto.TLFCryptKey) (
	[]kbfscrypto.TLFCryptKey, error) {
	encodedKeys, err := c.decryptData(encKeys.encryptedData, key.Data())
	if err != nil {
		return nil, err
	}

	var oldKeys []kbfscrypto.TLFCryptKey
	err = c.codec.Decode(encodedKeys, &oldKeys)
	if err != nil {
		return nil, err
	}

	return oldKeys, nil
}
