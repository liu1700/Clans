package agent_server

import (
	"Clans/server/flats"
	"Clans/server/netPackages"
	"fmt"

	"github.com/google/flatbuffers/go"
)

func RqUserLogin(sess *Session, pack *netPackages.NetPackage, outBuffer *Buffer) {
	rq := flats.GetRootAsRqLogin(pack.Data, 0)

	name := string(rq.Name())
	pw := string(rq.Password())

	fmt.Println("name: ", name, " password ", pw)

	builder := flatbuffers.NewBuilder(0)
	rpName := builder.CreateByteString(rq.Name())

	flats.RpLoginStart(builder)
	flats.RpLoginAddName(builder, rpName)
	rp := flats.RpLoginEnd(builder)
	builder.Finish(rp)

	outBuffer.send(sess, flats.ReponseIdLogin, pack, builder.FinishedBytes())
}
