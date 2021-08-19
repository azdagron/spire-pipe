# spire-pipe
Experimental low-level CLI utility for issuing SPIRE RPCs from scripts

I hacked together this tool some time ago while writing some automated testing.
It is still quite experimental. You are welcome to use it however you please.

It is most useful when paired with a tool like
[jq](https://stedolan.github.io/jq/) since it requires JSON input and output
(which is martialed to/from protobufs for issuing gRPC calls).

Warning: there is limited support for streaming RPCS.

## Installation

```
go install github.com/azdagron/spire-pipe
```

## Examples

Obtain a list of registration entries (via the SPIRE Server API Unix domain socket):

```
$ jq -n '{}' | spire-pipe rpc entry list-entries | jq .
```

Obtain a list of registration entries (via the SPIRE Server TCP port using an admin SVID obtained over the Workload API):

```
$ jq -n '{}' | spire-pipe rpc entry list-entries --use-workload-api | jq .
```
