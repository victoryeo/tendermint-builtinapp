This is the tendermint built in app as created from tendermint core documentation.

#### build our binary
go build

#### run tendermint
TMHOME="/tmp/example" tendermint init

#### run our application
./tendermint-builtinapp -config "/tmp/example/config/config.toml"