package trie

import (
	"bytes"
	"fmt"
	"io"
)

// WitnessVersion represents the current version of the block witness
// in case of incompatible changes it should be updated and the code to migrate the
// old witness format should be present
const WitnessVersion = uint8(1)

// WitnessHeader contains version information and maybe some future format bits
// the version is always the 1st bit.
type WitnessHeader struct {
	Version uint8
}

func (h *WitnessHeader) WriteTo(out *WitnessStatsCollector) error {
	_, err := out.WithColumn(ColumnStructure).Write([]byte{h.Version})
	return err
}

func (h *WitnessHeader) LoadFrom(input io.Reader) error {
	version := make([]byte, 1)
	if _, err := input.Read(version); err != nil {
		return err
	}

	h.Version = version[0]
	return nil
}

func defaultWitnessHeader() WitnessHeader {
	return WitnessHeader{WitnessVersion}
}

type Witness struct {
	Header   WitnessHeader
	Operands []WitnessOperand
}

func NewWitness(operands []WitnessOperand) *Witness {
	return &Witness{
		Header:   defaultWitnessHeader(),
		Operands: operands,
	}
}

func (w *Witness) WriteTo(out io.Writer) (*BlockWitnessStats, error) {
	statsCollector := NewWitnessStatsCollector(out)

	if err := w.Header.WriteTo(statsCollector); err != nil {
		return nil, err
	}

	for _, op := range w.Operands {
		if err := op.WriteTo(statsCollector); err != nil {
			return nil, err
		}
	}
	return statsCollector.GetStats(), nil
}

func NewWitnessFromReader(input io.Reader, trace bool) (*Witness, error) {
	var header WitnessHeader
	if err := header.LoadFrom(input); err != nil {
		return nil, err
	}

	if header.Version != WitnessVersion {
		return nil, fmt.Errorf("unexpected witness version: expected %d, got %d", WitnessVersion, header.Version)
	}

	opcode := make([]byte, 1)
	var err error
	operands := make([]WitnessOperand, 0)
	for _, err = input.Read(opcode); ; _, err = input.Read(opcode) {
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		var op WitnessOperand
		switch Instruction(opcode[0]) {
		case OpHash:
			op = &OperandHash{}
		case OpLeaf:
			op = &OperandLeafValue{}
		case OpAccountLeaf:
			op = &OperandLeafAccount{}
		case OpCode:
			op = &OperandCode{}
		case OpBranch:
			op = &OperandBranch{}
		case OpEmptyRoot:
			op = &OperandEmptyRoot{}
		case OpExtension:
			op = &OperandExtension{}
		default:
			return nil, fmt.Errorf("unexpected opcode while reading witness: %x", opcode[0])
		}

		err = op.LoadFrom(input)
		if err != nil {
			return nil, err
		}

		if trace {
			fmt.Printf("read op %T -> %+v\n", op, op)
		}

		operands = append(operands, op)
	}
	if trace {
		fmt.Println("end of read ***** ")
	}
	if err != nil && err != io.EOF {
		return nil, err
	}

	return &Witness{Header: header, Operands: operands}, nil
}

func (w *Witness) WriteDiff(w2 *Witness, output io.Writer) {
	if w.Header.Version != w2.Header.Version {
		fmt.Fprintf(output, "w1 header %d; w2 header %d\n", w.Header.Version, w2.Header.Version)
	}

	if len(w.Operands) != len(w2.Operands) {
		fmt.Fprintf(output, "w1 operands: %d; w2 operands: %d\n", len(w.Operands), len(w2.Operands))
	}

	for i := 0; i < len(w.Operands); i++ {
		switch o1 := w.Operands[i].(type) {
		case *OperandBranch:
			o2, ok := w2.Operands[i].(*OperandBranch)
			if !ok {
				fmt.Fprintf(output, "o1[%d] = %T; o2[%d] = %T\n", i, o1, i, o2)
			}
			if o1.Mask != o2.Mask {
				fmt.Fprintf(output, "o1[%d].Mask = %v; o2[%d].Mask = %v", i, o1.Mask, i, o2.Mask)
			}
		case *OperandHash:
			o2, ok := w2.Operands[i].(*OperandHash)
			if !ok {
				fmt.Fprintf(output, "o1[%d] = %T; o2[%d] = %T\n", i, o1, i, o2)
			}
			if !bytes.Equal(o1.Hash.Bytes(), o2.Hash.Bytes()) {
				fmt.Fprintf(output, "o1[%d].Hash = %s; o2[%d].Hash = %s\n", i, o1.Hash.Hex(), i, o2.Hash.Hex())
			}
		case *OperandCode:
			o2, ok := w2.Operands[i].(*OperandCode)
			if !ok {
				fmt.Fprintf(output, "o1[%d] = %T; o2[%d] = %T\n", i, o1, i, o2)
			}
			if !bytes.Equal(o1.Code, o2.Code) {
				fmt.Fprintf(output, "o1[%d].Code = %x; o2[%d].Code = %x\n", i, o1.Code, i, o2.Code)
			}
		case *OperandEmptyRoot:
			o2, ok := w2.Operands[i].(*OperandEmptyRoot)
			if !ok {
				fmt.Fprintf(output, "o1[%d] = %T; o2[%d] = %T\n", i, o1, i, o2)
			}
		case *OperandExtension:
			o2, ok := w2.Operands[i].(*OperandExtension)
			if !ok {
				fmt.Fprintf(output, "o1[%d] = %T; o2[%d] = %T\n", i, o1, i, o2)
			}
			if !bytes.Equal(o1.Key, o2.Key) {
				fmt.Fprintf(output, "extension o1[%d].Key = %x; o2[%d].Key = %x\n", i, o1.Key, i, o2.Key)
			}
		case *OperandLeafAccount:
			o2, ok := w2.Operands[i].(*OperandLeafAccount)
			if !ok {
				fmt.Fprintf(output, "o1[%d] = %T; o2[%d] = %T\n", i, o1, i, o2)
			}
			if !bytes.Equal(o1.Key, o2.Key) {
				fmt.Fprintf(output, "leafAcc o1[%d].Key = %x; o2[%d].Key = %x\n", i, o1.Key, i, o2.Key)
			}
			if o1.Nonce != o2.Nonce {
				fmt.Fprintf(output, "leafAcc o1[%d].Nonce = %v; o2[%d].Nonce = %v\n", i, o1.Nonce, i, o2.Nonce)
			}
			if o1.Balance.String() != o2.Balance.String() {
				fmt.Fprintf(output, "leafAcc o1[%d].Balance = %v; o2[%d].Balance = %v\n", i, o1.Balance.String(), i, o2.Balance.String())
			}
			if o1.HasCode != o2.HasCode {
				fmt.Fprintf(output, "leafAcc o1[%d].HasCode = %v; o2[%d].HasCode = %v\n", i, o1.HasCode, i, o2.HasCode)
			}
			if o1.HasStorage != o2.HasStorage {
				fmt.Fprintf(output, "leafAcc o1[%d].HasStorage = %v; o2[%d].HasStorage = %v\n", i, o1.HasStorage, i, o2.HasStorage)
			}
		case *OperandLeafValue:
			o2, ok := w2.Operands[i].(*OperandLeafValue)
			if !ok {
				fmt.Fprintf(output, "o1[%d] = %T; o2[%d] = %T\n", i, o1, i, o2)
			}
			if !bytes.Equal(o1.Key, o2.Key) {
				fmt.Fprintf(output, "leafVal o1[%d].Key = %x; o2[%d].Key = %x\n", i, o1.Key, i, o2.Key)
			}
			if !bytes.Equal(o1.Value, o2.Value) {
				fmt.Fprintf(output, "leafVal o1[%d].Value = %x; o2[%d].Value = %x\n", i, o1.Value, i, o2.Value)
			}
		default:
			o2 := w2.Operands[i]
			fmt.Fprintf(output, "unexpected o1[%d] = %T; o2[%d] = %T\n", i, o1, i, o2)
		}
	}
}