package types

import (
	"bytes"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/ledgerwatch/turbo-geth/common"
	"github.com/ledgerwatch/turbo-geth/rlp"
	"github.com/ledgerwatch/turbo-geth/trie"
)

func genTransactions(n uint64) Transactions {
	txs := Transactions{}

	for i := uint64(0); i < n; i++ {
		tx := NewTransaction(i, common.Address{}, big.NewInt(1000+int64(i)), 10+i, big.NewInt(1000+int64(i)), []byte(fmt.Sprintf("hello%d", i)))
		txs = append(txs, tx)
	}

	return txs
}

func TestEncodeUint(t *testing.T) {
	for i := 0; i < 64000; i++ {
		bbOld := bytes.NewBuffer(make([]byte, 10))
		bbNew := bytes.NewBuffer(make([]byte, 10))
		bbOld.Reset()
		bbNew.Reset()
		_ = rlp.Encode(bbOld, uint(i))

		bbNew.Reset()
		encodeUint(uint(i), bbNew)

		if !bytes.Equal(bbOld.Bytes(), bbNew.Bytes()) {
			t.Errorf("unexpected byte sequence. got: %x (expected %x)", bbNew.Bytes(), bbOld.Bytes())
		}
	}
}

type testcase struct {
	list           DerivableList
	expectedShaHex string
}

func TestDeriveSha(t *testing.T) {
	tests := []testcase{
		{Transactions{}, "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421"},
		{genTransactions(1), "0x1b18d86a2d8b21c60cdd147160af5bf2500cd822b93e3c0dc5ed3801b9721c62"},
		{genTransactions(2), "0xa11afe4b9db3814f46efc1bff52dc0643f536d2ae3aedc01c3381fe48e2a3665"},
		{genTransactions(4), "0xff19377cf77f25caa88628861212c455d134876cad344d27c9cd49e1ba73eb77"},
		{genTransactions(10), "0x48ebdecfd8da114d2a7997a6956d531e812f8f2b0f59726807f6eb5d542f2242"},
		{genTransactions(100), "0xb527630238d74db8346725270bd7a06abfed8eb867ee22ef55f1c4a6ae325338"},
		{genTransactions(1000), "0xfac5325c99717fb1b03007121d05d7cf7a7629cf814b9c4d99180b0f9104c0ee"},
		{genTransactions(10000), "0x87f89cf66cd7a5a191947bd2ad515f28e95b5370e20bdec7886314253feb2892"},
		{genTransactions(100000), "0xffcb005d4cf0ff66a9549eb69f9501e021ec43ea309ba4a47c1a2f15a91879cf"},
	}

	for _, test := range tests {
		checkDeriveSha(t, test)
	}
}

func checkDeriveSha(t *testing.T, tc testcase) {
	legacySha := legacyDeriveSha(tc.list)
	deriveSha := DeriveSha(tc.list)

	if !strings.EqualFold(deriveSha.Hex(), tc.expectedShaHex) {
		t.Errorf("unexpected hash: %v (expected: %v)\n", deriveSha.Hex(), tc.expectedShaHex)
	}

	if !strings.EqualFold(legacySha.Hex(), tc.expectedShaHex) {
		t.Errorf("unexpected hash: %v (expected: %v)\n", legacySha.Hex(), tc.expectedShaHex)
	}

	if !hashesEqual(legacySha, deriveSha) {
		t.Errorf("unexpected hash: %v (expected: %v)\n", deriveSha.Hex(), legacySha.Hex())
	}
}

func hashesEqual(h1, h2 common.Hash) bool {
	if len(h1) != len(h2) {
		return false
	}
	return h1.Hex() == h2.Hex()
}

func legacyDeriveSha(list DerivableList) common.Hash {
	keybuf := new(bytes.Buffer)
	trie := trie.New(common.Hash{})
	for i := 0; i < list.Len(); i++ {
		keybuf.Reset()
		_ = rlp.Encode(keybuf, uint(i))
		trie.Update(keybuf.Bytes(), list.GetRaw(i), 0)
	}
	return trie.Hash()
}

var (
	smallTxList = genTransactions(100)
	largeTxList = genTransactions(100000)
)

func BenchmarkLegacySmallList(b *testing.B) {
	for i := 0; i < b.N; i++ {
		legacyDeriveSha(smallTxList)
	}
}

func BenchmarkCurrentSmallList(b *testing.B) {
	for i := 0; i < b.N; i++ {
		DeriveSha(smallTxList)
	}
}

func BenchmarkLegacyLargeList(b *testing.B) {
	for i := 0; i < b.N; i++ {
		legacyDeriveSha(largeTxList)
	}
}

func BenchmarkCurrentLargeList(b *testing.B) {
	for i := 0; i < b.N; i++ {
		DeriveSha(largeTxList)
	}
}
