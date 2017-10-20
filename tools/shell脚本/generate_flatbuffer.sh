#!/bin/bash -e
WORK_PATH=$GOPATH/src/Clans
PROTO_DIR="$WORK_PATH/flatSchemas"
TOOL_DIR="$WORK_PATH/tools/bin"

GO_OUT_DIR="$WORK_PATH/server/"
CSHARP_OUT_DIR=$FRONTEND_DIR/Assets/RunAway/Proxy/

cd $PROTO_DIR
echo "Generating flatbuffers for Golang"
parallel '{1} --go -o {2} {3}; echo {3/.}' ::: $TOOL_DIR/flatc ::: $GO_OUT_DIR ::: *.fbs

cd $PROTO_DIR
echo "Generating flatbuffers for CSharp"
parallel '{1} --csharp -o {2} {3}; echo {3/.}' ::: $TOOL_DIR/flatc ::: $CSHARP_OUT_DIR ::: *.fbs