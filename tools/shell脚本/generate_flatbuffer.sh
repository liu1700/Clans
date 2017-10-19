#!/bin/bash -e
WORK_PATH=$GOPATH/src/throne
PROTO_DIR="$WORK_PATH/tools/proto"
COMMON_DIR="$PROTO_DIR/common"
TOOL_DIR="$WORK_PATH/tools/tools/"
BACKEND_DIR="$WORK_PATH/server/"
FRONTEND_DIR="/Users/guotao/Projects/throne_client/client" # 本地前端项目路径（自己修改）
UnityAppPath="/Applications/Unity" #自己的unity路径（自己修改）

GO_OUT_DIR="$BACKEND_DIR/throne/app/proto_structs/"
COMMON_GO_OUT_DIR="$BACKEND_DIR/throne/app/common/"
CSHARP_OUT_DIR="$FRONTEND_DIR/proj_unity/Assets/Scripts/Proxy/ProtoVO/" 
MONO_CMD="mono"


cd $PROTO_DIR
echo "Generating Proto for Golang"
protoc --go_out=$COMMON_GO_OUT_DIR --proto_path=$COMMON_DIR $COMMON_DIR/*.proto
protoc --go_out=$GO_OUT_DIR --proto_path=$PROTO_DIR $PROTO_DIR/*.proto


cd $PROTO_DIR
echo "Generating Proto for CSharp"
`rm -r $CSHARP_OUT_DIR/*.cs`
# for file_name in *.proto
# do
#   base_name=${file_name%%.*}
#   $MONO_CMD $TOOL_DIR/ProtoGen/protogen.exe -i:$base_name.proto -o:$CSHARP_OUT_DIR/$base_name.cs
# done
parallel '{1} {2} -i:{4/.}.proto -o:{3}/{4/.}.cs; echo {4/.}' ::: $MONO_CMD ::: $TOOL_DIR/ProtoGen/protogen.exe ::: $CSHARP_OUT_DIR ::: *.proto

cd $COMMON_DIR
for file_name in *.proto
do
  base_name=${file_name%%.*}
  $MONO_CMD $TOOL_DIR/ProtoGen/protogen.exe -i:$base_name.proto -o:$CSHARP_OUT_DIR/$base_name.cs
done

echo "生成lua需要的pb文件,由于刚修改的C#文件，unity会编译脚本，可能会慢一点"
$UnityAppPath/Unity.app/Contents/MacOS/Unity -quit -batchmode -executeMethod SLua.LuaNwtWorkInit.ServerInit path $PROTO_DIR # -logFile log.log  输出log到当前目录，需要时打开
if [ $? != 0 ];
then
	echo "unity执行失败，可能是你的unity已经打开了这个项目，如果是的话请先关闭。不是这个原因的话，可以去sh脚本中打开log，再次运行查看原因。"
	exit -1;
else
	echo "执行结束"
fi


