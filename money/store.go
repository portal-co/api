package money

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"errors"
	"io"

	"github.com/holiman/uint256"
)

type Nft[T any] interface {
	New(ctx context.Context, me uint256.Int, data T) (uint256.Int, error)
	Get(ctx context.Context, name uint256.Int) (NftData[T], bool, error)
	Burn(ctx context.Context, name uint256.Int) error
}
type NftData[T any] struct {
	Data   T
	Id     uint256.Int
	Minter uint256.Int
}
type DumpNft[T any, N Nft[[]byte]] struct {
	Internal N
}

func (d DumpNft[T, N]) New(ctx context.Context, me uint256.Int, data T) (uint256.Int, error) {
	j, err := json.Marshal(data)
	if err != nil {
		return uint256.Int{}, err
	}
	return d.Internal.New(ctx, me, j)
}
func (d DumpNft[T, N]) Burn(ctx context.Context, name uint256.Int) error {
	return d.Internal.Burn(ctx, name)
}
func (d DumpNft[T, N]) Get(ctx context.Context, name uint256.Int) (NftData[T], bool, error) {
	n, ok, err := d.Internal.Get(ctx, name)
	if !ok || err != nil {
		return NftData[T]{}, ok, err
	}
	var o T
	err = json.Unmarshal(n.Data, &o)
	if err != nil {
		return NftData[T]{}, false, err
	}
	return NftData[T]{Data: o, Id: n.Id, Minter: n.Minter}, true, nil
}

type CryptNft[N Nft[[]byte]] struct {
	Internal N
	Key      [32]byte
}

func encrypt(plaintext []byte, key []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
func (c CryptNft[N]) New(ctx context.Context, me uint256.Int, data []byte) (uint256.Int, error) {
	o, err := encrypt(data, c.Key[:])
	if err != nil {
		return uint256.Int{}, err
	}
	return c.Internal.New(ctx, me, o)
}
func (c CryptNft[N]) Burn(ctx context.Context, name uint256.Int) error {
	return c.Internal.Burn(ctx, name)
}
func (c CryptNft[N]) Get(ctx context.Context, name uint256.Int) (NftData[[]byte], bool, error) {
	n, ok, err := c.Internal.Get(ctx, name)
	if !ok || err != nil {
		return NftData[[]byte]{}, ok, err
	}
	o, err := decrypt(n.Data, c.Key[:])
	if err != nil {
		return NftData[[]byte]{}, false, err
	}
	return NftData[[]byte]{Data: o, Id: n.Id, Minter: n.Minter}, true, nil
}
