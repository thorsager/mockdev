# Mockdev

A set of services that will allow you to _mockup_ a networking devices, for testing and such. This works by dumping
conversations between a client and a device, and being able to _replay_ these at will.

I created this project, jus because I needed this for some projects I was working on, and there might be some other
projects out there that does the same, probably even a better job.

## Supported protocols

This project will support multiple protocols such as `SNMP`, `SSH`, `HTTP` and perhaps `TELNET`. I will be implementing
support in the above mentioned order.

## SNMP

The project is able to replay a dump of a OID-tree or just a sub tree. To see how this is configured have a look
at [config.yaml](_examples/configuration/config.yaml). It is quite straight forward.
`snapshot-files` can be created using the [snmp-snapshot](cmd/snmpsnapshot/snmp_snapshot.go) tool.

# Run under docker

```
docker run --rm -p 161:1161/udp ghcr.io/thorsager/mockdev
```

```
docker run --rm -p 161:1161/udp -v `pwd`/_testdata/snmp-config.yaml:/config/mockdev.yaml ghcr.io/thorsager/mockdev
```

## Verify running

```
snmpwalk -v 2c -c public localhost
```

# Creating snapshots

```
docker run -v `pwd`:/tmp ghcr.io/thorsager/mockdev snmp-snapshot -v -n -f -o /tmp/snapshot.txt -c $COMMUNITY $HOST 
```

# Match-groups in HTTP conversations
Match-groups found to the `path-matcher` or `body-matcher` are available in the `response.body` and `response.headeres[]` 
using go-tempting. Groups from the `path-matcher` are available as `{{ .p<number> }} {{ .b<number> }}` where `<number>` 
is the number of the match-group, and `p` denotes that it is groups from the `path-matcher`, `b` denotes that it is 
groups from the `body-matcher`. Number `0` will contain the entire match.

The match-groups are also available in `response.script` and `after-script` where they can be accessed as env-vars named
in the same manor as described above, ex `echo $p1 >> the_file.log`

# Environment variables in conversations
Any environment variable prefixed with `MOCKDEV_` will be available in conversations, when generating response body.
ex. `MOCDEV_FOO` will be available using `{{ .env.FOO }}`. The current bind address, and the bind port is available as:
`{{ .cfg.Address }}` and `{{ .cfg.Port }}` (_note that the `Address` will most likely be `""` meaning that mockdev is
bound to all addresses_)

# Scripted responses
It is possible to generate the entire http-conversation response using the `response.script` function. This expects that
a _complete_ raw HTTP response message is written to `STDOUT`. For further details look at [script.yaml](_examples/configuration/http_conversations/script.yaml)

# Current Time in conversations
The current time in "local" and in "GMT" is available in the context of templates now and can be accessed using 
`{{ .currentTime }}` and `{{ .currentTime_GMT }}`. Please note that date formatting is available in template context,
formatting is done using [time.Format](https://golang.org/pkg/time/#Time.Format). 
Ex. `{{ .currentTime.Format "Mon, 02 Jan 2006 15:04:05 MST" }`

# Breaking conversations
It is now possible to make the match, or the failure to match a conversation break the "conversation matching" and
respond with the conversation breaking the matching. This is done by setting the `break-on` property on a conversation.

Possible values are
  - `no-match` This will break the conversation-matching if the specific conversation is not matched. Please note that
    conversation will only break if _configured_ matchers fail.
  - `match` Will break the matching if the conversation matches.

An example of usage can be found in [config.yaml](_examples/configuration/config.yaml) in the "fake-auth" conversation.


# Thank You
This project builds on [slayercat/GoSNMPServer](https://github.com/slayercat/GoSNMPServer) for all the SNMP serving _(I
have made a [fork](https://github.com/thorsager/GoSNMPServer) for maintenance)_ and the [gliderlabs/ssh](https://github.com/gliderlabs/ssh)
for SSH serving.