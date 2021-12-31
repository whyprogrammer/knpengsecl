/*
Copyright (c) Huawei Technologies Co., Ltd. 2021.
kunpengsecl licensed under the Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
    http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
PURPOSE.
See the Mulan PSL v2 for more details.

Author: wucaijun/lixinda
Create: 2021-11-12
Description: Implement a privacy CA to sign identity key(AIK).

<<A Practical Guide to TPM2.0>>
	--Using the Trusted Platform Module in the New Age of Security
	    Will Arthur and David Challener
	    With Kenneth Goldman

FIGURE 9-1. Activating a Credential (CHAPTER 9/Page 109)
        Credential Provider(Privacy CA)                 TPM
                            Public Key, TPM Encryption Key Certificate
                                            <<===
1.         Validate Certificate chain <=
2.                 Examine Public Key <=
3.                Generate Credential <=
4.Generate Secret and wrap Credential <=
5.        Generate Seed, encrypt Seed <=
              with TPM Encryption Key
6.     Use Seed in KDF (with Name) to <=
        derive HMAC key and Symmetric
        Key, Wrap Secret in Symmetric
        Key and protect with HMAC Key
                                Credential wrapped by Secret,
                    Secret wrapped by Symmetric Key derived from Seed,
                            Seed encrypted by TPM Encryption Key.
                                            ===>>
                                                    1.=> Decrypt Seed using TPM Encryption Key.
                                                    2.=> Compute Name.
                                                    3.=> Use Seed KDF (with Name) to derive
                                                         HMAC Key and Symmetric Key.
                                                    4.=> Use Symmetric Key to unwrap Secret.
                                                    5.=> Use Secret to unwrap Credential.

The following happens at the credential provider: (Page 110)
1. The credential provider receives the Key's public area and a certificate for
an Encryption Key. The Encryption Key is typically a primary key in the
endorsement hierarchy, and its certificate is issued by the TPM and/or platform
manufacturer.
2. The credential provider walks the Encryption Key certificate chain back to
the issuer's root. Typically, the provider verifies that the Encryption Key is
fixed to a known compliant hardware TPM.
3. The provider examines the Key's public area and decides whether to issue a
certificate, and what the certificate should say. In a typical case, the
provider issues a certificate for a restricted Key that is fixed to the TPM.
4. The requester may have tried to alter the Key's public area attributes.
This attack won't be successful. See step 5 in the process that occurs at the
TPM.
5. The provider generates a credential for the Key.
6. The provider generates a Secret that is used to protect the credential.
Typically, this is a symmetric encryption key, but it can be a secret used to
generate encryption and integrity keys. The format and use of this secret aren't
mandated by the TCG.
7. The provider generates a 'Seed' to a key derivation function(KDF). If the
Encryption Key is an RSA key, the Seed is simply a random number, because an
RSA key can directly encrypt and decrypt. If the Decryption Key is an elliptic
curve cryptography(ECC) key, a more complex procedure using a Diffie-Hellman
protocol is required.
8. This Seed is encrypted by the Encryption Key public key. It can later only
be decrypted by the TPM.
9. The Seed is used in a TCG-specified KDF to generate a symmetric encryption
key and an HMAC key. The symmetric key is used to encrypt the Secret, and the
HMAC key provides integrity. Subtle but important is that the KDF also uses
the key's Name. You'll see why later.
10. The encrypted Secret and its integrity value are sent to the TPM in a
credential blob. The encrypted Seed is sent as well.

If you follow all this, you have the following:
#) A credential protected by a Secret
#) A Secret encrypted by a key derived from a Seed and the key's Name
#) A Seed encrypted by a TPM Encryption Key

These thins happen at the TPM:
1. The encrypted Seed is applied against the TPM Encryption Key, and the Seed
is recovered. The Seed remains inside the TPM.
2. The TPM computes the loaded key's Name.
3. The Name and the Seed are combined using the same TCG KDF to produce a
symmetric encryption key and an HMAC key.
4. The two keys are applied to the protected Secret, checking its integrity
and decrypting it.
5. This is where an attack on the key's public area attributes is detected.
If the attacker presents a key to the credential provider that is different
from the key loaded in the TPM, the Name will differ, and thus the symmetric
and HMAC keys will differ, and this step will fail.
6. The TPM returns the Secret.

Outside the TPM, the Secret is applied to the credential in some agreed upon
way. This can be as simple as using the Secret as a symmetric decryption key
to decrypt the credential. This protocol assures the credential provider that
the credential can only be recovered if:
#) The TPM has the private key associated with the Encryption key certificate.
#) The TPM has a key identical to the one presented to the credential provider.
The privacy administrator should control the use of the Endorsement Key, both
as a signing key and in the activate-credential protocol, and thus control its
correlation to another TPM key.

Other Privacy Considerations
The TPM owner can clear the storage hierarchy, changing the storage primary
seed and effectively erasing all storage hierarchy keys.
The platform owner controls the endorsement hierarchy. The platform owner
typically doesn't allow the endorsement primary seed to be changed, because
this would render the existing TPM certificates useless, with no way to recover.
The user can create other primary keys in the endorsement hierarchy using a
random number in the template. The user can erase these keys by flushing the
key from the TPM, deleting external copies, and forgetting the random number.
However, these keys do not have a manufacturer certificate.
When keys are used to sign(attest to) certain data, the attestation response
structure contains what are possibly privacy-sensitive fields: resetCount(the
number of times the TPM has been reset), restartCount(the number of times the
TPM has been restarted or resumed), and the firmware version. Although these
values don't map directly to a TPM, they can aid in correlation.
To avoid this issue, the values are obfuscated when the signing key isn't in
the endorsement or platform hierarchy. The obfuscation is consistent when
using the same key so the receiver can detect a change in the values while
not seeing the actual values.
*/

package pca

import (
	"bytes"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"math"
)

const (
	// algorithms define from tpm2
	AlgNull = 0x0000
	AlgRSA  = 0x0001
	AlgAES  = 0x0006
	AlgOAEP = 0x0017
	AlgCTR  = 0x0040
	AlgOFB  = 0x0041
	AlgCBC  = 0x0042
	AlgCFB  = 0x0043

	KEYSIZE = 32
)

var (
	ErrUnsupported = errors.New("unsupported parameters")
)

// GetRandomBytes gets random bytes
func GetRandomBytes(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	return b, err
}

func pkcs7Pad(c []byte, n int) []byte {
	pad := n - len(c)%n
	pt := bytes.Repeat([]byte{byte(pad)}, pad)
	return append(c, pt...)
}

func pkcs7Unpad(d []byte) []byte {
	n := len(d)
	pad := int(d[n-1])
	return d[:n-pad]
}

func aesCBCEncrypt(key, iv, plaintext []byte) ([]byte, error) {
	cb, err := aes.NewCipher(key)
	if err != nil {
		return []byte{}, err
	}
	if iv == nil {
		iv = bytes.Repeat([]byte("\x00"), cb.BlockSize())
	}
	n := cb.BlockSize()
	d := pkcs7Pad(plaintext, n)
	bm := cipher.NewCBCEncrypter(cb, iv)
	ciphertext := make([]byte, len(d))
	bm.CryptBlocks(ciphertext, d)
	return ciphertext, nil
}

func aesCBCDecrypt(key, iv, ciphertext []byte) ([]byte, error) {
	cb, err := aes.NewCipher(key)
	if err != nil {
		return []byte{}, err
	}
	if iv == nil {
		iv = bytes.Repeat([]byte("\x00"), cb.BlockSize())
	}
	bm := cipher.NewCBCDecrypter(cb, iv)
	plaintext := make([]byte, len(ciphertext))
	bm.CryptBlocks(plaintext, ciphertext)
	return pkcs7Unpad(plaintext), nil
}

func aesCFBEncrypt(key, iv, plaintext []byte) ([]byte, error) {
	cb, err := aes.NewCipher(key)
	if err != nil {
		return []byte{}, err
	}
	if iv == nil {
		iv = bytes.Repeat([]byte("\x00"), cb.BlockSize())
	}
	st := cipher.NewCFBEncrypter(cb, iv)
	ciphertext := make([]byte, len(plaintext))
	st.XORKeyStream(ciphertext, plaintext)
	return ciphertext, nil
}

func aesCFBDecrypt(key, iv, ciphertext []byte) ([]byte, error) {
	cb, err := aes.NewCipher(key)
	if err != nil {
		return []byte{}, err
	}
	if iv == nil {
		iv = bytes.Repeat([]byte("\x00"), cb.BlockSize())
	}
	st := cipher.NewCFBDecrypter(cb, iv)
	plaintext := make([]byte, len(ciphertext))
	st.XORKeyStream(plaintext, ciphertext)
	return plaintext, nil
}

func aesOFBEncDec(key, iv, in []byte) ([]byte, error) {
	cb, err := aes.NewCipher(key)
	if err != nil {
		return []byte{}, err
	}
	if iv == nil {
		iv = bytes.Repeat([]byte("\x00"), cb.BlockSize())
	}
	st := cipher.NewOFB(cb, iv)
	out := make([]byte, len(in))
	st.XORKeyStream(out, in)
	return out, nil
}

func aesCTREncDec(key, iv, in []byte) ([]byte, error) {
	cb, err := aes.NewCipher(key)
	if err != nil {
		return []byte{}, err
	}
	if iv == nil {
		iv = bytes.Repeat([]byte("\x00"), cb.BlockSize())
	}
	st := cipher.NewCTR(cb, iv)
	out := make([]byte, len(in))
	st.XORKeyStream(out, in)
	return out, nil
}

// SymmetricEncrypt uses key/iv to encrypt the plaintext with symmetric algorithm/mode.
func SymmetricEncrypt(alg, mod uint16, key, iv, plaintext []byte) ([]byte, error) {
	switch alg {
	case AlgAES:
		switch mod {
		case AlgCBC:
			return aesCBCEncrypt(key, iv, plaintext)
		case AlgCFB:
			return aesCFBEncrypt(key, iv, plaintext)
		case AlgOFB:
			return aesOFBEncDec(key, iv, plaintext)
		case AlgCTR:
			return aesCTREncDec(key, iv, plaintext)
		}
	}
	return []byte{}, ErrUnsupported
}

// SymmetricDecrypt uses key/iv to decrypt the ciphertext with symmetric algorithm/mode.
func SymmetricDecrypt(alg, mod uint16, key, iv, ciphertext []byte) ([]byte, error) {
	switch alg {
	case AlgAES:
		switch mod {
		case AlgCBC:
			return aesCBCDecrypt(key, iv, ciphertext)
		case AlgCFB:
			return aesCFBDecrypt(key, iv, ciphertext)
		case AlgOFB:
			return aesOFBEncDec(key, iv, ciphertext)
		case AlgCTR:
			return aesCTREncDec(key, iv, ciphertext)
		}
	}
	return []byte{}, ErrUnsupported
}

// AsymmetricEncrypt encrypts a byte array by public key and label using RSA
func AsymmetricEncrypt(alg, mod uint16, pubKey crypto.PublicKey, plaintext, label []byte) ([]byte, error) {
	switch alg {
	case AlgRSA:
		switch mod {
		case AlgOAEP:
			return rsa.EncryptOAEP(sha256.New(), rand.Reader, pubKey.(*rsa.PublicKey), plaintext, label)
		default:
			return rsa.EncryptPKCS1v15(rand.Reader, pubKey.(*rsa.PublicKey), plaintext)
		}
	}
	return []byte{}, ErrUnsupported
}

// AsymmetricDecrypt decrypts a byte array by private key and label using RSA
func AsymmetricDecrypt(alg, mod uint16, priKey crypto.PrivateKey, ciphertext, label []byte) ([]byte, error) {
	switch alg {
	case AlgRSA:
		switch mod {
		case AlgOAEP:
			return rsa.DecryptOAEP(sha256.New(), rand.Reader, priKey.(*rsa.PrivateKey), ciphertext, label)
		default:
			return rsa.DecryptPKCS1v15(rand.Reader, priKey.(*rsa.PrivateKey), ciphertext)
		}
	}
	return []byte{}, ErrUnsupported
}

/*
Trusted Platform Module Library
Part 1: Architecture
Family "2.0"
Level 00 Revision 01.59
November 8, 2019
Published
  Contact admin@trustedcomputinggroup.org
  TCG Published
  Copyright(c) TCG 2006-2020

Page 43
11.4.10 Key Derivation Function
11.4.10.1 Introduction
The TPM uses a hash-based function to generate keys for multiple purposes. This
specification uses two different schemes: one for ECDH and one for all other
uses of a KDF.
The ECDH KDF is from SP800-56A. The Counter mode KDF, from SP800-108, uses HMAC
as the pseudo-random function (PRF). It is refered to in the specification as
KDFa().

11.4.10.2 KDFa()
With the exception of ECDH, KDFa() is used in all cases where a KDF is required.
KDFa() uses Counter mode from SP800-108, with HMAC as the PRF.
As defined in SP800-108, the inner loop for building the key stream is:
        K(i) := HMAC(K, [i] || Label || 00 || Context || [L])           (6)
where
    K(i)    the i(th) iteration of the KDF inner loop
    HMAC()  the HMAC algorithm using an approved hash algorithm
    K       the secret key material
    [i]     a 32-bit counter that starts at 1 and increments on each iteration
    Label   a octet stream indicating the use of the key produced by this KDF
    00      Added only if Label is not present or if the last octect of Label is not zero
    Context a binary string containing information relating to the derived keying material
    [L]     a 32-bit value indicating the number of bits to be returned from the KDF
NOTE1
Equation (6) is not KDFa(). KDFa() is the function call defined below.

As shown in equation (6), there is an octet of zero that separates Label from
Context. In SP800-108, Label is a sequence of octets that may or may not have
a final octet that is zero. If Label is not present, a zero octet is added.
If Label is present and the last octet is not zero, a zero octet is added.
After each iteration, the HMAC digest data is concatenated to the previously
produced value until the size of the concatenated string is at least as large
as the requested value. The string is then truncated to the desired size (which
causes the loss of some of the most recently added bits), and the value is
returned.

When this specification calls for use of this KDF, it uses a function reference
to KDFa(). The function prototype is:
        KDFa(hashAlg, key, label, contextU, contextV, bits)             (7)
where
    hashAlg  a TPM_ALG_ID to be used in the HMAC in the KDF
    key      a variable-sized value used as K
    label    a variable-sized octet stream used as Label
    contextU a variable-sized value concatenated with contextV to create the
             Context parameter used in equation (6) above
    contextV a variable-sized value concatenated with contextU to create the
             Context parameter used in equation (6) above
    bits     a 32-bit value used as [L], and is the number of bits returned
             by the function

The values of contextU and contextV are passed as sized buffers and only the
buffer data is used to construct the Context parameter used in equation (6)
above. The size fields of contextU and contextV are not included in the
computation. That is:
        Context := contextU.buffer || contextV.buffer                   (8)

The 32-bit value of bits is in TPM canonical form, with the least significant
bits of the value in the highest numbered octet.

The implied return from this function is a sequence of octets with a length
equal to (bits+7)/8. If bits is not an even multiple of 8, then the returned
value occupies the least significant bits of the returned octet array, and
the additional, high-order bits in the 0(th) octet are CLEAR.
The unused bits of the most significant octet(MSO) are masked off and not shifted.

EXAMPLE
If KDFa() were used to produce a 521-bit ECC private key, the returned value
would occupy 66 octets, with the upper 7 bits of the octet at offset zero
set to 0.
*/
func KDFa(alg crypto.Hash, key []byte, label string, contextU, contextV []byte, bits int) ([]byte, error) {
	bufLen := ((bits + 7) / 8)
	if bufLen > math.MaxInt16 {
		return []byte{}, ErrUnsupported
	}
	buf := []byte{}
	h := hmac.New(alg.New, key)
	for i := 1; len(buf) < bufLen; i++ {
		h.Reset()
		binary.Write(h, binary.BigEndian, uint32(i))
		if len(label) > 0 {
			h.Write([]byte(label))
			h.Write([]byte{0})
		}
		if len(contextU) > 0 {
			h.Write(contextU)
		}
		if len(contextV) > 0 {
			h.Write(contextV)
		}
		binary.Write(h, binary.BigEndian, uint32(bits))
		buf = h.Sum(buf)
	}
	buf = buf[:bufLen]
	mask := uint8(bits % 8)
	if mask > 0 {
		buf[0] &= (1 << mask) - 1
	}
	return buf, nil
}

/*
Trusted Platform Module Library
Part 1: Architecture
Family "2.0"
Level 00 Revision 01.59
November 8, 2019
Published
  Contact admin@trustedcomputinggroup.org
  TCG Published
  Copyright(c) TCG 2006-2020

Page 161
24 Credential Protection
24.1 Introduction
The TPM supports a privacy preserving protocol for distributing credentials for
keys on a TPM. The process allows a credential provider to assign a credential
to a TPM object, such that the credential provider cannot prove that the object
is resident on a particular TPM, but the credential is not available unless the
object is resident on a device that the credential provider believes is an
authentic TPM.

24.2 Protocol
The initiator of the credential process will provide, to a credential provider,
the public area of a TPM object for which a credential is desired along with
the credentials for a TPM key (usually an EK). The credential provider will
inspect the credentials of the "EK" and the properties indicated in the public
area to determine if the object should receive a credential. If so, the
credential provider will issue a credential for the public area.
The credential provider may require that the credential only be useable if the
public area is a valid object on the same TPM as the "EK". To ensure this, the
credential provider encrypts a challenge and then "wraps" the challenge
encryption key with the public key of the "EK".
NOTE:
  "EK" is used to indicate that an EK is typically used for this process but
any storage key may be used. It is up to the credential provider to decide
what is acceptable for an "EK".

The encrypted challenge and the wrapped encryption key are then delivered to
the initiator. The initiator can decrypt the challenge by loading the "EK"
and the object onto the TPM and asking the TPM to return the challenge. The
TPM will decrypt the challenge using the private "EK" and validate that the
credentialed object (public and private) is loaded on the TPM. If so, the
TPM has validated that the properties of the object match the properties
required by the credential provider and the TPM will return the challenge.
This process preserves privacy by allowing TPM TPM objects to have credentials
from the credential provider that are not tied to a specific TPM. If the
object is a signing key, that key may be used to sign attestations, and the
credential can assert that the signing key is on a valid TPM without disclosing
the exact TPM.
A second property of this protocol is that it prevents the credential provider
from proving anything about the object for which it provided the credential.
The credential provider could have produced the credential with no information
from the TPM as the TPM did not need to provide a proof-of-possession of any
private key in order for the credential provider to create the credential.
The credential provider can know that the credential for the object could not
be in use unless the object was on the same TPM as the "EK", but the credential
provider cannot prove it.

24.3 Protection of Credential
The credential blob (which typically contains the information used to decrypt
the challenge) from the credential provider contains a value that is returned
by the TPM if the TPM2_ActivateCredential() is successful. The value may be
anything that the credential provider wants to place in the credential blob
but is expected to be simply a large random number.
The credential provider protects the credential value (CV) with an integrity
HMAC and encryption in much the same way as a credential blob. The difference
is, when SEED is generated, the label is "IDENTITY" instead of "DUPLICATE".

24.4 Symmetric Encrypt
A SEED is derived from values that are protected by the asymmetric algorithm
of the "EK". The methods of generating the SEED are determined by the
asymmetric algorithm of the "EK" and are described in an annex to this TPM 2.0
Part 1. In the process of creating SEED, the label is required to be "INTEGRITY".
NOTE:
  If a duplication blob is given to the TPM, its HMAC key will be wrong and
the HMAC check will fail.

Given a value for SEED, a key is created by:
        symKey := KDFa(ekNameAlg, SEED, "STORAGE", name, NULL, bits)    (44)
where
    ekNameAlg  the nameAlg of the key serving as the "EK"
    SEED       the symmetric seed value produced using methods specific to
               the type of asymmetric algorithms of the "EK"
    "STORAGE"  a value used to differentiate the uses of the KDF
    name       the Name of the object associated with the credential
    bits       the number of bits required for the symmetric key
The symKey is used to encrypt the CV. The IV is set to 0.
        encIdentity := CFB(symKey, 0, CV)                               (45)
where
    CFB        symmetric encryption in CFB mode using the symmetric
               algorithm of the key serving as "EK"
    symKey     symmetric key from (44)
    CV         the credential value (a TPM2B_DIGEST)

24.5 HMAC
A final HMAC operation is applied to the encIdentity value. This is to ensure
that the TPM can properly associate the credential with a loaded object and
to prevent misuse of or tampering with the CV.
The HMAC key (HMACkey) for the integrity is computed by:
        HMACkey := KDFa(ekNameAlg, SEED, "INTEGRITY", NULL, NULL, bits) (46)
where
    ekNameAlg    the nameAlg of the target "EK"
    SEED         the symmetric seed value used in (44); produced using
                 methods specific to the type of asymmetric algorithms
                 of the "EK"
    "INTEGRITY"  a value used to differentiate the uses of the KDF
    bits         the number of bits in the digest produced by ekNameAlg
NOTE:
  Even though the same value for label is used for each integrit HMAC, SEED
is created in a manner that is unique to the application. Since SEED is
unique to the application, the HMAC is unique to the application.

HMACkey is then used in the integrity computation.
        identityHMAC := HMAC(HMACkey, encIdentity || Name)              (47)
where
    HMAC         the HMAC function using nameAlg of the "EK"
    HMACkey      a value derived from the "EK" symmetric protection
                 value according to equation (46)
    encIdentity  symmetrically encrypted sensitive area produced in (45)
    Name         the Name of the object being protected
The integrity structure is constructed by placing the identityHMAC (size and
hash) in the buffer ahead of the encIdentity.

24.6 Summary of Protection Process
1. Marshal the CV(credential value) into a TPM2B_DIGEST
2. Using methods of the asymmetric "EK", create a SEED value
3. Create a symmetric key for encryption:
     symKey := KDFa(ekNameAlg, SEED, "STORAGE", Name, NULL, bits)
4. Create encIdentity by encryption the CV
     encIdentity := CFB(symKey, 0, CV)
5. Compute the HMACkey
     HMACkey := KDFa(ekNameAlg, SEED, "INTEGRITY", NULL, NULL, bits)
6. Compute the HMAC over the encIdentity from step 4
     outerHMAC := HMAC(HMACkey, encIdentity || Name)

Also reference
Trusted Platform Module Library
Part 3: Commands
Family "2.0"
Level 00 Revision 01.59
November 8, 2019
Page 72
12.6 TPM2_MakeCredential
*/

func MakeCredential(ekPubKey crypto.PublicKey, credential, name []byte) ([]byte, []byte, error) {
	if len(credential) == 0 || len(name) == 0 || len(credential) > crypto.SHA256.Size() {
		return nil, nil, ErrUnsupported
	}
	// step 1, size(uint16) + content
	plaintext := new(bytes.Buffer)
	binary.Write(plaintext, binary.BigEndian, uint16(len(credential)))
	binary.Write(plaintext, binary.BigEndian, credential)
	// step 2,
	seed, _ := GetRandomBytes(KEYSIZE)
	encSeed, _ := AsymmetricEncrypt(AlgRSA, AlgOAEP, ekPubKey, seed, []byte("IDENTITY\x00"))
	// step 3
	symKey, _ := KDFa(crypto.SHA256, seed, "STORAGE", name, nil, KEYSIZE*8)
	// step 4
	encIdentity, _ := SymmetricEncrypt(AlgAES, AlgCFB, symKey, nil, plaintext.Bytes())
	// step 5
	hmacKey, _ := KDFa(crypto.SHA256, seed, "INTEGRITY", nil, nil, crypto.SHA256.Size()*8)
	// step 6
	integrityBuf := new(bytes.Buffer)
	binary.Write(integrityBuf, binary.BigEndian, encIdentity)
	binary.Write(integrityBuf, binary.BigEndian, name)
	mac := hmac.New(sha256.New, hmacKey)
	mac.Write(integrityBuf.Bytes())
	integrity := mac.Sum(nil)
	// last step: prepare output
	allBlob := new(bytes.Buffer)
	binary.Write(allBlob, binary.BigEndian, uint16(len(integrity)))
	binary.Write(allBlob, binary.BigEndian, integrity)
	binary.Write(allBlob, binary.BigEndian, encIdentity)
	return allBlob.Bytes(), encSeed, nil
}
