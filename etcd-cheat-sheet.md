# etcd cheat-sheet

### Prerequisities
- list of etcd members, each in a form `https://<MEMBER_IP>:2379`
- default naming and paths of certificates
- check all etcd members have at least one snapshot file before the migration

#### Cluster-health
```go
/usr/bin/etcdctl --cert-file /etc/etcd/peer.crt --key-file /etc/etcd/peer.key \
--ca-file /etc/etcd/ca.crt -C https://<MEMBER_IP>:2379 cluster-health
```

#### Endpoint health check (will list endpoint health of members listed in --endpoints)
```sh
ETCDCTL_API=3 /usr/bin/etcdctl --cert /etc/etcd/peer.crt --key /etc/etcd/peer.key --cacert /etc/etcd/ca.crt \
--endpoints 'https://<MEMBER1_IP>:2379,https://<MEMBER2_IP>:2379,https://<MEMBER3_IP>:2379' \
endpoint health
```

#### Endpoint status check (will list endpoint status of members listed in --endpoints)
```sh
ETCDCTL_API=3 /usr/bin/etcdctl --cert /etc/etcd/peer.crt --key /etc/etcd/peer.key --cacert /etc/etcd/ca.crt \
--endpoints 'https://<MEMBER1_IP>:2379,https://<MEMBER2_IP>:2379,https://<MEMBER3_IP>:2379' \
endpoint status -w table
```

The command can return the following information:
```sh
+----------------------------+------------------+---------+---------+-----------+-----------+------------+
|          ENDPOINT          |        ID        | VERSION | DB SIZE | IS LEADER | RAFT TERM | RAFT INDEX |
+----------------------------+------------------+---------+---------+-----------+-----------+------------+
| https://172.16.186.13:2379 | 4469408199a3c995 |   3.2.5 |  6.3 MB |      true |        39 |    1869922 |
| https://172.16.186.22:2379 | 1fcf4328ae68e365 |   3.2.5 |  6.3 MB |     false |        39 |    1869922 |
| https://172.16.186.10:2379 | abaa7fe06d236c0d |   3.2.5 |  6.3 MB |     false |        39 |    1869922 |
+----------------------------+------------------+---------+---------+-----------+-----------+------------+
```

The `DB SIZE` column is an indicator the etcd data got properly propagated after a member got added to the cluster.

#### List all v2 keys
```sh
/usr/bin/etcdctl --cert-file /etc/etcd/peer.crt --key-file /etc/etcd/peer.key \
--ca-file /etc/etcd/ca.crt -C https://<MEMBER_IP>:2379 ls -r
```

#### List all v3 keys
```sh
ETCDCTL_API=3 /usr/bin/etcdctl --cert /etc/etcd/peer.crt --key /etc/etcd/peer.key \
--cacert /etc/etcd/ca.crt --endpoints https://<MEMBER_IP>:2379 get "" --from-key --keys-only
```

#### List all cluster members
```sh
/usr/bin/etcdctl --cert-file /etc/etcd/peer.crt --key-file /etc/etcd/peer.key \
--ca-file /etc/etcd/ca.crt -C https://<MEMBER1_IP>:2379 member list
```

The command can return the following information:
```sh
1fcf4328ae68e365: name=172.16.186.22 peerURLs=https://172.16.186.22:2380 clientURLs=https://172.16.186.22:2379 isLeader=false
4469408199a3c995: name=172.16.186.13 peerURLs=https://172.16.186.13:2380 clientURLs=https://172.16.186.13:2379 isLeader=true
abaa7fe06d236c0d: name=172.16.186.10 peerURLs=https://172.16.186.10:2380 clientURLs=https://172.16.186.10:2379 isLeader=false
```

The `name=*` is a name of a member (that can be used when a new member with the same name is added).
It is not possible to add a member with the same name (the member has to be removed from the cluster first).

#### Add a member

```vi
Caution:
When adding a second member to the cluster, the etcd will stop working until the second member
is fully started. The reason is the 2-member etcd cluster needs to establish a quorom.
Thus, both etcd member services must be up and running. There will be a pause right after
the `etcdctl member add` command is performed and right before the etcd service is run
and the new member joins the cluster. So it is expected to observe the cluster being unreachable
or unhealthy during the pause.
```

1. the cluster must be up and running properly (you can check health status by running `etcdctl cluster-health` and `etcdctl endpoint health)`

2. add a member to the cluster via
   ```sh
   /usr/bin/etcdctl --cert-file /etc/etcd/peer.crt --key-file /etc/etcd/peer.key --ca-file /etc/etcd/ca.crt \
   -C https://<MEMBER1_IP>:2379 member add <MEMBER2_NAME> https://<MEMBER2_IP>:2380
   ```
   The command will print environment variables that needs to be set in `/etc/etcd/etcd.conf` file.

3. Update the `/etc/etcd/etcd.conf` file accordingaly

#### Migrate

```sh
ETCDCTL_API=3 /usr/bin/etcdctl migrate --data-dir=/var/lib/etcd
```
