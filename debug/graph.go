package debug

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/solver/pb"
	digest "github.com/opencontainers/go-digest"
	"github.com/pkg/errors"
)

type llbOp struct {
	Op         pb.Op
	Digest     digest.Digest
	OpMetadata pb.OpMetadata
}

func LLBToGraph(state llb.State, dot bool) (res string, err error) {
	var (
		definition *llb.Definition
		buffer     bytes.Buffer
	)

	definition, err = state.Marshal()
	if err != nil {
		err = errors.Wrap(err, "marshaling llb state")
		return
	}

	var ops []llbOp
	for _, dt := range definition.Def {
		var op pb.Op

		if err = (&op).Unmarshal(dt); err != nil {
			err = errors.Wrap(err, "failed to parse op")
			return
		}

		dgst := digest.FromBytes(dt)
		ent := llbOp{Op: op, Digest: dgst, OpMetadata: definition.Metadata[dgst]}
		ops = append(ops, ent)

	}

	if dot {
		writeDot(ops, &buffer)
	} else {
		writeJson(ops, &buffer)
	}

	res = buffer.String()

	return
}

func writeJson(ops []llbOp, w io.Writer) {
	enc := json.NewEncoder(os.Stdout)
	for _, op := range ops {
		if err := enc.Encode(op); err != nil {
			panic(err)
		}
	}
}

func writeDot(ops []llbOp, w io.Writer) {
	fmt.Fprintln(w, "digraph {")
	defer fmt.Fprintln(w, "}")
	for _, op := range ops {
		name, shape := attr(op.Digest, op.Op)
		fmt.Fprintf(w, "  %q [label=%q shape=%q];\n", op.Digest, name, shape)
	}
	for _, op := range ops {
		for i, inp := range op.Op.Inputs {
			label := ""
			if eo, ok := op.Op.Op.(*pb.Op_Exec); ok {
				for _, m := range eo.Exec.Mounts {
					if int(m.Input) == i && m.Dest != "/" {
						label = m.Dest
					}
				}
			}
			fmt.Fprintf(w, "  %q -> %q [label=%q];\n", inp.Digest, op.Digest, label)
		}
	}
}

func attr(dgst digest.Digest, op pb.Op) (string, string) {
	switch op := op.Op.(type) {
	case *pb.Op_Source:
		return op.Source.Identifier, "ellipse"
	case *pb.Op_Exec:
		return strings.Join(op.Exec.Meta.Args, " "), "box"
	case *pb.Op_Build:
		return "build", "box3d"
	default:
		return dgst.String(), "plaintext"
	}
}
