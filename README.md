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
docker run --rm -p 161:1161/udp $DOCKER_IMAGE
```

```
docker run --rm -p 161:1161/udp -v `pwd`/_testdata/snmp-config.yaml:/config/mockdev.yaml $DOCKER_IMAGE
```

## Verify running

```
snmpwalk -v 2c -c public localhost
```

# Creating snapshots

```
docker run -v `pwd`:/tmp $DOCKER_IMAGE snmp-snapshot -v -n -f -o /tmp/snapshot.txt -c $COMMUNITY $HOST 
```